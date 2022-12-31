/*
Copyright 2022 The Ksflow Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"text/template"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	"github.com/twmb/franz-go/pkg/sr"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	KafkaSchemaFinalizerName = "kafka-schema.ksflow.io/finalizer"
)

type KafkaSchemaReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	KafkaSchemaConfig ksfv1.KafkaSchemaConfig
	NameTemplate      *template.Template
}

//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaschemas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaschemas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkaschemas/finalizers,verbs=update

func (r *KafkaSchemaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get KafkaSchema
	var ks ksfv1.KafkaSchema
	if err := r.Get(ctx, req.NamespacedName, &ks); err != nil {
		logger.Error(err, "unable to get KafkaSchema")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create Kafka client
	srClient, err := r.KafkaSchemaConfig.SchemaRegistryConnectionConfig.NewClient()
	if err != nil {
		logger.Error(err, "unable to create Schema Registry client")
		return ctrl.Result{}, err
	}

	// Reconcile
	ksCopy := ks.DeepCopy()
	err = r.reconcileSchema(ksCopy, srClient)
	ksCopy.Status.DeepCopyInto(&ks.Status)

	// Update in-cluster spec w/finalizers
	if !equality.Semantic.DeepEqual(ks.Finalizers, ksCopy.Finalizers) {
		ks.SetFinalizers(ksCopy.Finalizers)
		if specErr := r.Client.Update(ctx, &ks); specErr != nil {
			if err != nil {
				err = fmt.Errorf("failed while updating spec: %v: %v", specErr, err)
			} else {
				err = fmt.Errorf("failed to update spec: %v", specErr)
			}
		}
	}

	// Update in-cluster status
	ksCopy.Status.DeepCopyInto(&ks.Status)
	if statusErr := r.Client.Status().Update(ctx, &ks); statusErr != nil {
		if err != nil {
			err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
		} else {
			err = fmt.Errorf("failed to update status: %v", statusErr)
		}
	}

	return ctrl.Result{}, err
}

// reconcileSchema handles reconciliation of a KafkaSchema
func (r *KafkaSchemaReconciler) reconcileSchema(ks *ksfv1.KafkaSchema, srClient *sr.Client) error {
	ks.Status.LastUpdated = metav1.Now()
	ks.Status.Phase = ksfv1.KsflowPhaseUnknown
	ks.Status.Reason = ""

	// Validate the KafkaSchema
	errs := validation.IsDNS1035Label(ks.Name)
	if len(errs) > 0 {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = fmt.Sprintf("invalid KafkaSchema name: %q", errs[0])
		return fmt.Errorf(ks.Status.Reason)
	}

	// Compute the Kafka schema name
	finalSubjectName, err := ks.FinalName(r.NameTemplate)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = fmt.Sprintf("unable to render Kafka schema from template: %q", err)
		return fmt.Errorf(ks.Status.Reason)
	}
	ks.Status.SubjectName = finalSubjectName

	// Schema deletion & finalizers
	if !ks.DeletionTimestamp.IsZero() {
		ks.Status.Phase = ksfv1.KsflowPhaseDeleting
	}
	ret, err := handleSchemaDeletionAndFinalizers(ks, srClient)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}
	if ret {
		return nil
	}

	// Create or update schema
	if err = r.createOrUpdateSubject(&ks.Spec.KafkaSubjectInClusterConfiguration, ks.Status.SubjectName, srClient); err != nil {
		return err
	}
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}

	// Update status
	var sicc *ksfv1.KafkaSubjectInClusterConfiguration
	sicc, err = r.getSubjectInClusterConfiguration(ks.Status.SubjectName, srClient)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}
	if sicc != nil {
		ks.Status.KafkaSubjectInClusterConfiguration = *sicc
	}
	sCount, err := schemaCount(ks.Status.SubjectName, srClient)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}
	ks.Status.SchemaCount = sCount

	err = r.subjectIsUpToDate(ks.Spec.KafkaSubjectInClusterConfiguration, ks.Status.KafkaSubjectInClusterConfiguration)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseUpdating
		ks.Status.Reason = err.Error()
		return err
	}

	ks.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

// subjectIsUpToDate returns an error indicating why the subject is not up-to-date, or nil if it is up-to-date
func (r *KafkaSchemaReconciler) subjectIsUpToDate(specKSICC ksfv1.KafkaSubjectInClusterConfiguration, statusKSICC ksfv1.KafkaSubjectInClusterConfiguration) error {
	if !r.KafkaSchemaConfig.IgnoreSchemaMode && specKSICC.Mode != statusKSICC.Mode {
		return fmt.Errorf("spec mode %q does not match status mode %q", specKSICC.Mode.ToFranz().String(), statusKSICC.Mode.ToFranz().String())
	}
	if specKSICC.Type != statusKSICC.Type {
		return fmt.Errorf("spec type %q does not match status type %q", specKSICC.Type.ToFranz().String(), statusKSICC.Type.ToFranz().String())
	}
	if specKSICC.CompatibilityLevel != statusKSICC.CompatibilityLevel {
		return fmt.Errorf("spec compatibility level %q does not match status compatibility level %q", specKSICC.CompatibilityLevel.ToFranz().String(), statusKSICC.CompatibilityLevel.ToFranz().String())
	}
	if specKSICC.Schema != statusKSICC.Schema {
		return fmt.Errorf(`spec schema does not match status schema`)
	}
	if len(specKSICC.References) != len(statusKSICC.References) {
		return fmt.Errorf(`spec has %d references, which does not match status which has %d references`, len(specKSICC.References), len(statusKSICC.References))
	}
	for i := range specKSICC.References {
		if specKSICC.References[i].Name != statusKSICC.References[i].Name {
			return fmt.Errorf(`spec reference name %s does not match status reference name %s`, specKSICC.References[i].Name, statusKSICC.References[i].Name)
		}
		if specKSICC.References[i].Subject != statusKSICC.References[i].Subject {
			return fmt.Errorf(`spec reference subject %s does not match status reference subject %s`, specKSICC.References[i].Subject, statusKSICC.References[i].Subject)
		}
		if specKSICC.References[i].Version != statusKSICC.References[i].Version {
			return fmt.Errorf(`spec reference version %d does not match status reference version %d`, specKSICC.References[i].Version, statusKSICC.References[i].Version)
		}
	}
	return nil
}

// handleSchemaDeletionAndFinalizers updates finalizers if necessary and handles deletion of kafka schemas
// returns false if processing should continue, true if we should finish reconcile
func handleSchemaDeletionAndFinalizers(ks *ksfv1.KafkaSchema, srClient *sr.Client) (bool, error) {
	if ks.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(ks, KafkaSchemaFinalizerName) {
			controllerutil.AddFinalizer(ks, KafkaSchemaFinalizerName)
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(ks, KafkaSchemaFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := deleteSubjectFromRegistry(ks.Status.SubjectName, srClient); err != nil {
				return true, err
			}
			exists, err := subjectExists(ks.Status.SubjectName, srClient)
			if err != nil {
				return true, err
			}
			if exists {
				// return err so that it uses exponential backoff (ref: https://github.com/kubernetes-sigs/controller-runtime/issues/808#issuecomment-639845414)
				return true, errors.New("waiting for schema to finish deleting")
			}
			controllerutil.RemoveFinalizer(ks, KafkaSchemaFinalizerName)
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

func (r *KafkaSchemaReconciler) createOrUpdateSubject(
	desired *ksfv1.KafkaSubjectInClusterConfiguration,
	subjectName string,
	srClient *sr.Client) error {

	var sicc *ksfv1.KafkaSubjectInClusterConfiguration
	sicc, err := r.getSubjectInClusterConfiguration(subjectName, srClient)
	if err != nil {
		return err
	}
	if sicc != nil {
		if err = r.updateSubjectInRegistry(desired, sicc, subjectName, srClient); err != nil {
			return err
		}
	} else {
		if err = r.createSubjectInRegistry(desired, subjectName, srClient); err != nil {
			return err
		}
	}
	return nil
}

// schemaCount returns the number of schemas for the subject
func schemaCount(subjectName string, srClient *sr.Client) (*int32, error) {
	schemas, err := srClient.Schemas(context.Background(), subjectName, sr.HideDeleted)
	if err != nil {
		return nil, err
	}
	return pointer.Int32(int32(len(schemas))), nil
}

// subjectExists returns true if the subject exists
func subjectExists(subjectName string, srClient *sr.Client) (bool, error) {
	subjects, err := srClient.Subjects(context.Background(), sr.ShowDeleted)
	if err != nil {
		return false, err
	}
	for _, s := range subjects {
		if s == subjectName {
			return true, nil
		}
	}
	return false, nil
}

// getSubjectInClusterConfiguration retrieves the current observed state for the given subjectName by making any necessary calls to the Registry
func (r *KafkaSchemaReconciler) getSubjectInClusterConfiguration(subjectName string, srClient *sr.Client) (*ksfv1.KafkaSubjectInClusterConfiguration, error) {
	kscc := ksfv1.KafkaSubjectInClusterConfiguration{}

	found, err := subjectExists(subjectName, srClient)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	if !r.KafkaSchemaConfig.IgnoreSchemaMode {
		modeResults := srClient.Mode(context.Background(), subjectName)
		for _, mr := range modeResults {
			if mr.Err != nil {
				return nil, mr.Err
			}
			if mr.Subject != subjectName {
				return nil, fmt.Errorf("received unexpected subject name when querying modes, was %q, expected %q", mr.Subject, subjectName)
			}
			kscc.Mode = ksfv1.KafkaSchemaMode(mr.Mode.String())
		}
		if kscc.Mode == ksfv1.KafkaSchemaModeUnknown {
			return nil, fmt.Errorf("empty mode results for subjectname %q", subjectName)
		}
	}

	compatibilityLevelResults := srClient.CompatibilityLevel(context.Background(), subjectName)
	for _, clr := range compatibilityLevelResults {
		if clr.Err != nil {
			return nil, clr.Err
		}
		if clr.Subject != subjectName {
			return nil, fmt.Errorf("received unexpected subject name when querying compatibility levels, was %q, expected %q", clr.Subject, subjectName)
		}
		kscc.CompatibilityLevel = ksfv1.KafkaSchemaCompatibilityLevel(clr.Level.String())
	}
	if kscc.CompatibilityLevel == ksfv1.KafkaSchemaCompatibilityLevelUnknown {
		return nil, fmt.Errorf("empty compatibility level results for subjectname %q", subjectName)
	}

	subjectSchema, err := srClient.SchemaByVersion(context.Background(), subjectName, -1, sr.HideDeleted)
	if err != nil {
		return nil, err
	}
	kscc.Type = ksfv1.KafkaSchemaType(subjectSchema.Type.String())
	kscc.Schema = subjectSchema.Schema.Schema
	kscc.References = subjectSchema.References

	return &kscc, nil
}

// updateSubjectInRegistry compares the desired state (coming from spec) and the observed state (coming from the status),
// making any necessary calls to the Registry to bring them closer together.
func (r *KafkaSchemaReconciler) updateSubjectInRegistry(
	desired *ksfv1.KafkaSubjectInClusterConfiguration,
	observed *ksfv1.KafkaSubjectInClusterConfiguration,
	subjectName string,
	srClient *sr.Client) error {

	if !r.KafkaSchemaConfig.IgnoreSchemaMode && desired.Mode != observed.Mode {
		mode := desired.Mode.ToFranz()
		modeResults := srClient.SetMode(context.Background(), mode, true, subjectName)
		if len(modeResults) == 0 {
			return errors.New("empty mode results")
		}
		for _, mr := range modeResults {
			if mr.Err != nil {
				return mr.Err
			}
			if mr.Subject != subjectName || mr.Mode != mode {
				return fmt.Errorf("unexpected mode result, %v", mr)
			}
		}
	}

	if desired.CompatibilityLevel != observed.CompatibilityLevel {
		compatLevel := desired.CompatibilityLevel.ToFranz()
		compatResults := srClient.SetCompatibilityLevel(context.Background(), compatLevel, subjectName)
		if len(compatResults) == 0 {
			return errors.New("empty compatibility level results")
		}
		for _, cr := range compatResults {
			if cr.Err != nil {
				return cr.Err
			}
			if cr.Subject != subjectName || cr.Level != compatLevel {
				return fmt.Errorf("unexpected compatibility level result, %v", cr)
			}
		}
	}

	refAreEqual := len(desired.References) == len(observed.References)
	if refAreEqual {
		for i, d := range desired.References {
			o := observed.References[i]
			if d.Name != o.Name || d.Subject != o.Subject || d.Version != o.Version {
				refAreEqual = false
				break
			}
		}
	}
	if desired.Type != observed.Type || desired.Schema != observed.Schema || !refAreEqual {
		schema := sr.Schema{
			Schema:     desired.Schema,
			Type:       desired.Type.ToFranz(),
			References: desired.References,
		}
		_, err := srClient.CreateSchema(context.Background(), subjectName, schema)
		if err != nil {
			return err
		}
	}

	return nil
}

// createSubjectInRegistry creates the specified kafka subject using the provided Schema Registry client
func (r *KafkaSchemaReconciler) createSubjectInRegistry(kscc *ksfv1.KafkaSubjectInClusterConfiguration, subjectName string, srClient *sr.Client) error {
	schema := sr.Schema{
		Schema:     kscc.Schema,
		Type:       kscc.Type.ToFranz(),
		References: kscc.References,
	}
	_, err := srClient.CreateSchema(context.Background(), subjectName, schema)
	if err != nil {
		return err
	}

	if !r.KafkaSchemaConfig.IgnoreSchemaMode {
		mode := kscc.Mode.ToFranz()
		modeResults := srClient.SetMode(context.Background(), mode, true, subjectName)
		if len(modeResults) == 0 {
			return errors.New("empty mode results")
		}
		for _, mr := range modeResults {
			if mr.Err != nil {
				return mr.Err
			}
			if mr.Subject != subjectName || mr.Mode != mode {
				return fmt.Errorf("unexpected mode result, %v", mr)
			}
		}
	}

	compatLevel := kscc.CompatibilityLevel.ToFranz()
	compatResults := srClient.SetCompatibilityLevel(context.Background(), compatLevel, subjectName)
	if len(compatResults) == 0 {
		return errors.New("empty compatibility level results")
	}
	for _, cr := range compatResults {
		if cr.Err != nil {
			return cr.Err
		}
		if cr.Subject != subjectName || cr.Level != compatLevel {
			return fmt.Errorf("unexpected compatibility level result, %v", cr)
		}
	}

	return nil
}

// deleteSubjectFromRegistry deletes the specified subject using the provided Schema Registry client
func deleteSubjectFromRegistry(subjectName string, srClient *sr.Client) error {
	notSoftDeleted, err := srClient.Subjects(context.Background(), sr.HideDeleted)
	if err != nil {
		return err
	}
	for _, v := range notSoftDeleted {
		if v == subjectName {
			_, err = srClient.DeleteSubject(context.Background(), subjectName, sr.SoftDelete)
			if err != nil {
				return err
			}
		}
	}

	_, err = srClient.DeleteSubject(context.Background(), subjectName, sr.HardDelete)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaSchemaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ntString := r.KafkaSchemaConfig.NameTemplate
	if len(ntString) == 0 {
		return errors.New("nameTemplate for kafkaSchemas was empty")
	}

	var err error
	if r.NameTemplate, err = template.New("kafka-schema").Parse(ntString); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaSchema{}).
		Complete(r)
}
