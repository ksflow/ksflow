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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	finalSchemaName, err := ks.FinalName(r.NameTemplate)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = fmt.Sprintf("unable to render Kafka schema from template: %q", err)
		return fmt.Errorf(ks.Status.Reason)
	}
	ks.Status.SchemaName = finalSchemaName

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
	if err = createOrUpdateSchema(&ks.Spec.KafkaSchemaInClusterConfiguration, ks.Status.SchemaName, srClient); err != nil {
		return err
	}
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}

	// Update status
	var ticc *ksfv1.KafkaSchemaInClusterConfiguration
	ticc, err = getSchemaInClusterConfiguration(ks.Status.SchemaName, srClient)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseError
		ks.Status.Reason = err.Error()
		return err
	}
	if ticc != nil {
		ks.Status.KafkaSchemaInClusterConfiguration = *ticc
	}

	err = schemaIsUpToDate(ks.Spec.KafkaSchemaInClusterConfiguration, ks.Status.KafkaSchemaInClusterConfiguration)
	if err != nil {
		ks.Status.Phase = ksfv1.KsflowPhaseUpdating
		ks.Status.Reason = err.Error()
		return err
	}

	ks.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

// createSchemaInRegistry creates the specified kafka schema using the provided Schema Registry client
func createSchemaInRegistry(kss *ksfv1.KafkaSchemaInClusterConfiguration, subjectName string, srClient *sr.Client) error {
	mode := kss.Mode.ToFranz()
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

	compatLevel := kss.CompatibilityLevel.ToFranz()
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

	schema := sr.Schema{
		Schema:     kss.Schema,
		Type:       kss.Type.ToFranz(),
		References: kss.References,
	}
	_, err := srClient.CreateSchema(context.Background(), subjectName, schema)
	if err != nil {
		return err
	}

	return nil
}

// deleteSchemaFromRegistry deletes the specified schema using the provided Schema Registry client
func deleteSchemaFromRegistry(subjectName string, srClient *sr.Client) error {
	// listing is not efficient if many subjects... revisit this
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
