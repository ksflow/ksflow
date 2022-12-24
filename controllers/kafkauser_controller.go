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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"text/template"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ksfv1 "github.com/ksflow/ksflow/api/v1alpha1"
)

// KafkaUserReconciler reconciles a KafkaUser object
type KafkaUserReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	KafkaConnectionConfig ksfv1.KafkaConnectionConfig
	CommonNameTemplate    *template.Template
}

//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ksflow.io,resources=kafkausers/finalizers,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/signers,resourceNames=ksflow.io/user-controller,verbs=approve;sign
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get KafkaUser
	var ku ksfv1.KafkaUser
	if err := r.Get(ctx, req.NamespacedName, &ku); err != nil {
		logger.Error(err, "unable to get KafkaUser")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get Secret
	var secret corev1.Secret
	if err := r.Get(ctx, ku.SecretNamespacedName(), &secret); err != nil {
		if apierrors.IsNotFound(err) {
			r.Create(ctx)
			// TODO: create & return
		} else {
			logger.Error(err, "unable to get Secret")
			return ctrl.Result{}, err
		}
	}

	// Get CertificateSigningRequest
	var csr certv1.CertificateSigningRequest
	if err := r.Get(ctx, ku.CertificateSigningRequestNamespacedName(), &csr); err != nil {
		if apierrors.IsNotFound(err) {
			if csr.Status.Certificate != nil {

			}
			// TODO: create & return
		} else {
			logger.Error(err, "unable to get CertificateSigningRequest")
			return ctrl.Result{}, err
		}
	}

	// Reconcile
	kuCopy := ku.DeepCopy()
	err := r.reconcileUser(kuCopy)

	// Update in-cluster status
	kuCopy.Status.DeepCopyInto(&ku.Status)
	if statusErr := r.Client.Status().Update(ctx, &ku); statusErr != nil {
		if err != nil {
			err = fmt.Errorf("failed while updating status: %v: %v", statusErr, err)
		} else {
			err = fmt.Errorf("failed to update status: %v", statusErr)
		}
	}

	return ctrl.Result{}, err
}

func (r *KafkaUserReconciler) reconcileUser(kafkaUser *ksfv1.KafkaUser) error {
	kafkaUser.Status.LastUpdated = metav1.Now()
	kafkaUser.Status.Phase = ksfv1.KsflowPhaseUnknown
	kafkaUser.Status.Reason = ""

	errs := validation.IsDNS1035Label(kafkaUser.Name)
	if len(errs) > 0 {
		kafkaUser.Status.Phase = ksfv1.KsflowPhaseError
		kafkaUser.Status.Reason = fmt.Sprintf("invalid KafkaUser name: %q", errs[0])
		return fmt.Errorf(kafkaUser.Status.Reason)
	}

	// Secret create or update
	// TODO: create secret with ssl and bootstrap.servers

	// Update status
	// TODO: update status

	kafkaUser.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

func (r *KafkaUserReconciler) getCSR(ctx context.Context, ku *ksfv1.KafkaUser) (*certv1.CertificateSigningRequest, error) {
	logger := log.FromContext(ctx)

	var csr certv1.CertificateSigningRequest
	err := r.Get(ctx, ku.CertificateSigningRequestNamespacedName(), &csr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			csr = certv1.CertificateSigningRequest{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       certv1.CertificateSigningRequestSpec{},
				Status:     certv1.CertificateSigningRequestStatus{},
			}
			if cerr := r.Create(ctx, &csr); cerr != nil {
				return nil, cerr
			}
		} else {
			return nil, err
		}
	}
	return &csr, nil
}

func (r *KafkaUserReconciler) getSecret(ctx context.Context, ku *ksfv1.KafkaUser) (*corev1.Secret, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	subject, err := r.CertificateSigningRequestPKIXName(ku)
	if err != nil {
		return nil, err
	}
	certificateRequest := x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	certificateRequestBytes, err := x509.CreateCertificateRequest(rand.Reader, &certificateRequest, privateKey)
	if err != nil {
		return nil, err
	}
	csr := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: ku.CertificateSigningRequestName(),
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request: pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE REQUEST",
				Bytes: certificateRequestBytes,
			}),
			SignerName:        r.KafkaConnectionConfig.KafkaCSRConfig.SignerName,
			ExpirationSeconds: r.KafkaConnectionConfig.KafkaCSRConfig.ExpirationSeconds,
			Usages:            []certv1.KeyUsage{certv1.UsageClientAuth},
		},
	}
	err = r.Create(ctx, csr)
	if err != nil {
		return nil, err
	}
	csr.Status.Conditions = append(csr.Status.Conditions, certv1.CertificateSigningRequestCondition{
		Type:           certv1.CertificateApproved,
		Reason:         "User is valid",
		Message:        "Approved by ksflow user controller",
		LastUpdateTime: metav1.Now(),
	})
	err = r.Update(ctx, csr)
	if err != nil {
		return nil, err
	}
	signedCsr, err = r.Get(ctx, &csr)
}

func (r *KafkaUserReconciler) CertificateSigningRequestPKIXName(ku *ksfv1.KafkaUser) (pkix.Name, error) {
	var tplBytes bytes.Buffer
	err := r.CommonNameTemplate.Execute(&tplBytes, types.NamespacedName{Namespace: ku.Namespace, Name: ku.Name})
	if err != nil {
		return pkix.Name{}, err
	}

	return pkix.Name{
		Country:            r.KafkaConnectionConfig.KafkaCSRConfig.Subject.Countries,
		Organization:       r.KafkaConnectionConfig.KafkaCSRConfig.Subject.Organizations,
		OrganizationalUnit: r.KafkaConnectionConfig.KafkaCSRConfig.Subject.OrganizationalUnits,
		Locality:           r.KafkaConnectionConfig.KafkaCSRConfig.Subject.Localities,
		Province:           r.KafkaConnectionConfig.KafkaCSRConfig.Subject.Provinces,
		StreetAddress:      r.KafkaConnectionConfig.KafkaCSRConfig.Subject.StreetAddresses,
		PostalCode:         r.KafkaConnectionConfig.KafkaCSRConfig.Subject.PostalCodes,
		CommonName:         tplBytes.String(),
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// parse template for common name to fail fast and also not build template every time
	cn := r.KafkaConnectionConfig.KafkaCSRConfig.Subject.CommonNameTemplate
	var err error
	r.CommonNameTemplate, err = template.New("commonName").Parse(cn)
	if err != nil {
		return err
	}

	// setup
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaUser{}).
		Owns(&certv1.CertificateSigningRequest{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
