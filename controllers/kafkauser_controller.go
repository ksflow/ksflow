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
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type KafkaUserReconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	KafkaConnectionConfig  ksfv1.KafkaConnectionConfig
	KafkaPrincipalTemplate *template.Template
}

//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers/finalizers,verbs=update

func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get KafkaUser
	var ku ksfv1.KafkaUser
	if err := r.Get(ctx, req.NamespacedName, &ku); err != nil {
		logger.Error(err, "unable to get KafkaUser")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Reconcile
	kuCopy := ku.DeepCopy()
	err := r.reconcileUser(ctx, kuCopy)
	kuCopy.Status.DeepCopyInto(&ku.Status)

	// Update in-cluster status
	if statusErr := r.Client.Status().Update(ctx, &ku); statusErr != nil {
		if err != nil {
			err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
		} else {
			err = fmt.Errorf("failed to update status: %v", statusErr)
		}
	}

	return ctrl.Result{}, err
}

// reconcileUser handles reconciliation of a KafkaUser
func (r *KafkaUserReconciler) reconcileUser(ctx context.Context, ku *ksfv1.KafkaUser) error {
	ku.Status.LastUpdated = metav1.Now()
	ku.Status.Phase = ksfv1.KsflowPhaseUnknown
	ku.Status.Reason = ""

	// Validate the KafkaUser
	errs := validation.IsDNS1035Label(ku.Name)
	if len(errs) > 0 {
		ku.Status.Phase = ksfv1.KsflowPhaseError
		ku.Status.Reason = fmt.Sprintf("invalid KafkaUser name: %q", errs[0])
		return fmt.Errorf(ku.Status.Reason)
	}

	// Compute the Kafka principal
	var tplBytes bytes.Buffer
	if err := r.KafkaPrincipalTemplate.Execute(&tplBytes, types.NamespacedName{Namespace: ku.Namespace, Name: ku.Name}); err != nil {
		ku.Status.Phase = ksfv1.KsflowPhaseError
		ku.Status.Reason = fmt.Sprintf("unable to render Kafka principal from template: %q", err)
		return fmt.Errorf(ku.Status.Reason)
	}
	kafkaPrincipalString := tplBytes.String()
	if !strings.HasPrefix(kafkaPrincipalString, "User:") && !strings.HasPrefix(kafkaPrincipalString, "Group:") {
		ku.Status.Phase = ksfv1.KsflowPhaseError
		ku.Status.Reason = fmt.Sprintf("unable to render Kafka principal from template, must begin with User: or Group:, was %q", kafkaPrincipalString)
		return fmt.Errorf(ku.Status.Reason)
	}
	ku.Status.KafkaPrincipal = tplBytes.String()

	// Return
	ku.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// parse template for common name to fail fast and also not build template every time
	principalTemplateStr := "User:ANONYMOUS"
	if r.KafkaConnectionConfig.PrincipalTemplate != nil && len(*r.KafkaConnectionConfig.PrincipalTemplate) > 0 {
		principalTemplateStr = *r.KafkaConnectionConfig.PrincipalTemplate
	}
	var err error
	r.KafkaPrincipalTemplate, err = template.New("kafka-principal").Parse(principalTemplateStr)
	if err != nil {
		return err
	}

	// setup
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaUser{}).
		Complete(r)
}
