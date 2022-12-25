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
	"os"
	"strings"
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

const (
	SecretRootCAKey           = "ca.crt"
	SecretBootstrapServersKey = "bootstrap.servers"

	UserResourceLabelKey = "ksflow.io/kafka-user"
)

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
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/signers,resourceNames=kubernetes.io/kube-apiserver-client,verbs=approve
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

	// Reconcile
	kuCopy := ku.DeepCopy()
	err := r.reconcileUser(ctx, kuCopy)

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

	// TODO: get secret
	// if not exist, get
	// TODO

	// Ensure temporary secret exists
	var tmpSecret corev1.Secret
	if err := r.Get(ctx, ku.TemporarySecretNamespacedName(), &tmpSecret); err != nil {
		if apierrors.IsNotFound(err) {
			s, serr := r.getNewTemporarySecret(ctx, ku)
			if serr != nil {
				ku.Status.Phase = ksfv1.KsflowPhaseError
				ku.Status.Reason = serr.Error()
				return fmt.Errorf(ku.Status.Reason)
			}
			if serr = r.Create(ctx, s); serr != nil {
				ku.Status.Phase = ksfv1.KsflowPhaseError
				ku.Status.Reason = serr.Error()
				return fmt.Errorf(ku.Status.Reason)
			}
			ku.Status.Reason = "private key generated"
			ku.Status.Phase = ksfv1.KsflowPhaseCreating
		} else {
			ku.Status.Phase = ksfv1.KsflowPhaseError
			ku.Status.Reason = err.Error()
			return fmt.Errorf(ku.Status.Reason)
		}
	}

	// Ensure secret is up-to-date
	updateSecret, err := secretNeedsUpdating(&secret)
	if err != nil {
		ku.Status.Phase = ksfv1.KsflowPhaseError
		ku.Status.Reason = err.Error()
		return fmt.Errorf(ku.Status.Reason)
	}
	if updateSecret {
		s, serr := r.getNewUserSecret(ctx, ku)
		if serr != nil {
			ku.Status.Phase = ksfv1.KsflowPhaseError
			ku.Status.Reason = serr.Error()
			return fmt.Errorf(ku.Status.Reason)
		}
		if serr = r.Update(ctx, s); serr != nil {
			ku.Status.Phase = ksfv1.KsflowPhaseError
			ku.Status.Reason = serr.Error()
			return fmt.Errorf(ku.Status.Reason)
		}
	}

	ku.Status.Phase = ksfv1.KsflowPhaseAvailable
	return nil
}

func (r *KafkaUserReconciler) getPrivateKey(ctx context.Context, ku *ksfv1.KafkaUser) (*rsa.PrivateKey, error) {
	// try to get from main secret if it exists
	privatekey, err := r.getPrivateKeyFromSecret(ctx, ku.SecretName())
	if err != nil {
		return nil, err
	}
	if privatekey != nil {
		return privatekey, nil
	}
	// next try to get from temporary secret if it exists
	privatekey, err = r.getPrivateKeyFromSecret(ctx, ku.TemporarySecretName())
	if err != nil {
		return nil, err
	}
	if privatekey != nil {
		return privatekey, nil
	}

	// neither secret has what we need, so go ahead and create one in the temporary secret
	// TODO
	var secret corev1.Secret
	if err := r.Get(ctx, ku.SecretName(), &secret); err != nil {
		if apierrors.IsNotFound(err) {
			if terr := r.Get(ctx, ku.TemporarySecretName(), &secret); terr != nil {
				if apierrors.IsNotFound(terr) {

				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}
}

// getPrivateKeyFromSecret gets the private key from the secret if the secret is found, if not found returns nil
// returns error if failed for any other reason
func (r *KafkaUserReconciler) getPrivateKeyFromSecret(ctx context.Context, secretName types.NamespacedName) (*rsa.PrivateKey, error) {
	var secret corev1.Secret
	if err := r.Get(ctx, secretName, &secret); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	privateKeyBytes := secret.Data[corev1.TLSPrivateKeyKey]
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (r *KafkaUserReconciler) getNewTemporarySecret(ctx context.Context, ku *ksfv1.KafkaUser) (*corev1.Secret, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	caBytes, err := os.ReadFile(r.KafkaConnectionConfig.KafkaTLSConfig.CAFilePath)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ku.TemporarySecretName().Name,
			Namespace: ku.TemporarySecretName().Namespace,
			Labels: map[string]string{
				UserResourceLabelKey: ku.Name,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			corev1.TLSPrivateKeyKey:   privateKeyBytes,
			SecretRootCAKey:           caBytes,
			SecretBootstrapServersKey: []byte(strings.Join(r.KafkaConnectionConfig.BootstrapServers, ",")),
		},
	}, nil
}

// gets a new user's secret
func (r *KafkaUserReconciler) getNewSecret(ctx context.Context, ku *ksfv1.KafkaUser) (*corev1.Secret, error) {

	certBytes, err := r.getNewUserCertificate(ctx, privateKey, ku)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ku.SecretNamespacedName().Name,
			Namespace: ku.SecretNamespacedName().Namespace,
			Labels: map[string]string{
				UserResourceLabelKey: ku.Name,
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:         certBytes,
			corev1.TLSPrivateKeyKey:   privateKeyBytes,
			SecretRootCAKey:           caBytes,
			SecretBootstrapServersKey: []byte(strings.Join(r.KafkaConnectionConfig.BootstrapServers, ",")),
		},
	}, nil
}

func (r *KafkaUserReconciler) getNewUserCertificate(ctx context.Context, pkey *rsa.PrivateKey, ku *ksfv1.KafkaUser) ([]byte, error) {
	subject, err := r.certificateSigningRequestPKIXName(ku)
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
			Name: ku.CertificateSigningRequestNamespacedName().Name,
			Labels: map[string]string{
				UserResourceLabelKey: fmt.Sprintf("%s.%s", ku.Name, ku.Namespace),
			},
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
		if apierrors.IsAlreadyExists(err) {
			if derr := r.Delete(ctx, csr); derr != nil {
				return nil, derr
			}
			if cerr := r.Create(ctx, csr); cerr != nil {
				return nil, cerr
			}
		} else {
			return nil, err
		}
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
	var signedCSR certv1.CertificateSigningRequest
	// TODO: how do we know this has been signed?  async...
	err = r.Get(ctx, ku.CertificateSigningRequestNamespacedName(), &signedCSR)
	if err != nil {
		return nil, err
	}
	if signedCSR.Status.Certificate == nil {
		return nil, fmt.Errorf("certificate from CSR was not signed")
	}
	return signedCSR.Status.Certificate, nil
}

func secretNeedsUpdating(secret *corev1.Secret) (bool, error) {
	// TODO implement, how do we know since it's different each time?
}

func (r *KafkaUserReconciler) certificateSigningRequestPKIXName(ku *ksfv1.KafkaUser) (pkix.Name, error) {
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
	var err error
	r.CommonNameTemplate, err = template.New("cn").Parse(r.KafkaConnectionConfig.KafkaCSRConfig.Subject.CommonNameTemplate)
	if err != nil {
		return err
	}

	// setup
	return ctrl.NewControllerManagedBy(mgr).
		For(&ksfv1.KafkaUser{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
