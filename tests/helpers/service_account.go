package helpers

import (
	"context"
	"fmt"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiserverserviceaccount "k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/client-go/util/keyutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Heavily copied from https://github.com/kubernetes/kubernetes/blob/ea0764452222146c47ec826977f49d7001b0ea8c/pkg/controller/serviceaccount/tokens_controller.go#L362

type ServiceAccountHelper interface {
	CreateServiceAccount(name, namespace string) (*corev1.ServiceAccount, error)
	CreateAuthableSecret(serviceAccount *corev1.ServiceAccount) (*corev1.Secret, error)
}

func NewServiceAccountHelper(tokenSigningKeyFilePath string, cl client.Client) (ServiceAccountHelper, error) {
	signingKeyPEMBytes, err := ioutil.ReadFile(tokenSigningKeyFilePath)
	if err != nil {
		return nil, err
	}
	signingKey, err := keyutil.ParsePrivateKeyPEM(signingKeyPEMBytes)
	if err != nil {
		return nil, err
	}

	generator, err := JWTTokenGenerator("kubernetes/serviceaccount", signingKey)
	if err != nil {
		return nil, err
	}

	return &serviceAccountHelper{
		tokenGenerator: generator,
		client:         cl,
	}, nil
}

type serviceAccountHelper struct {
	tokenGenerator TokenGenerator
	client         client.Client
}

func legacyClaims(serviceAccount corev1.ServiceAccount, secret corev1.Secret) (*jwt.Claims, interface{}) {
	return &jwt.Claims{
			Subject: apiserverserviceaccount.MakeUsername(serviceAccount.Namespace, serviceAccount.Name),
		}, &legacyPrivateClaims{
			Namespace:          serviceAccount.Namespace,
			ServiceAccountName: serviceAccount.Name,
			ServiceAccountUID:  string(serviceAccount.UID),
			SecretName:         secret.Name,
		}
}

type legacyPrivateClaims struct {
	ServiceAccountName string `json:"kubernetes.io/serviceaccount/service-account.name"`
	ServiceAccountUID  string `json:"kubernetes.io/serviceaccount/service-account.uid"`
	SecretName         string `json:"kubernetes.io/serviceaccount/secret.name"`
	Namespace          string `json:"kubernetes.io/serviceaccount/namespace"`
}

func (h *serviceAccountHelper) CreateServiceAccount(name, namespace string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := h.client.Create(context.TODO(), serviceAccount, &client.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return serviceAccount, nil
}

func (h *serviceAccountHelper) CreateAuthableSecret(serviceAccount *corev1.ServiceAccount) (*corev1.Secret, error) {
	if serviceAccount.UID == "" {
		return nil, fmt.Errorf("service account has no uid")
	}
	secretName := fmt.Sprintf("%s-token", serviceAccount.Name)
	serviceAccount.Secrets = []corev1.ObjectReference{
		{
			Name:      secretName,
			Namespace: serviceAccount.Namespace,
		},
	}
	err := h.client.Update(context.TODO(), serviceAccount)
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: serviceAccount.Namespace,
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: serviceAccount.Name,
				corev1.ServiceAccountUIDKey:  string(serviceAccount.UID),
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
		Data: map[string][]byte{},
	}

	// Generate the token
	token, err := h.tokenGenerator.GenerateToken(legacyClaims(*serviceAccount, *secret))
	if err != nil {
		return nil, err
	}

	secret.Data[corev1.ServiceAccountTokenKey] = []byte(token)
	secret.Data[corev1.ServiceAccountNamespaceKey] = []byte(serviceAccount.Namespace)

	err = h.client.Create(context.TODO(), secret, &client.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}
