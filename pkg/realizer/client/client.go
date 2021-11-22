package client

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientBuilder func(secret *corev1.Secret) (client.Client, error)

func NewClientBuilder(restConfig *rest.Config) ClientBuilder {
	return func(secret *corev1.Secret) (client.Client, error) {
		config, err := AddBearerToken(secret, restConfig)
		if err != nil {
			return nil, fmt.Errorf("adding bearer token: %w", err)
		}

		cl, err := client.New(config, client.Options{})
		if err != nil {
			return nil, fmt.Errorf("creating client: %w", err)
		}

		return cl, nil
	}
}

func AddBearerToken(secret *corev1.Secret, restConfig *rest.Config) (*rest.Config, error) {
	tokenBytes, found := secret.Data[corev1.ServiceAccountTokenKey]
	if !found {
		return nil, fmt.Errorf("couldn't find service account token value")
	}

	newConfig := *restConfig
	newConfig.BearerToken = string(tokenBytes)
	newConfig.BearerTokenFile = ""

	return &newConfig, nil
}
