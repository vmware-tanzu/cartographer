// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	"github.com/ghodss/yaml"
	clientcmd "k8s.io/client-go/tools/clientcmd/api/v1"
)

type KubeconfigRestricted struct {
	result string
}

// NewKubeconfigRestricted takes kubeconfig yaml as input and returns kubeconfig yaml with certain fields restricted (removed).
// Developers may find it informative to view their own config at ~/.kube/config
func NewKubeconfigRestricted(input string) (*KubeconfigRestricted, error) {
	var inputConfig clientcmd.Config

	err := yaml.Unmarshal([]byte(input), &inputConfig)
	if err != nil {
		return nil, fmt.Errorf("Parsing kubeconfig: %s", err)
	}

	if len(inputConfig.Clusters) == 0 {
		return nil, fmt.Errorf("Expected to find at least one cluster in kubeconfig")
	}

	resultConfig := clientcmd.Config{
		Kind:           inputConfig.Kind,
		APIVersion:     inputConfig.APIVersion,
		CurrentContext: inputConfig.CurrentContext,
	}

	for _, inputCluster := range inputConfig.Clusters {
		resultConfig.Clusters = append(resultConfig.Clusters, clientcmd.NamedCluster{
			Name: inputCluster.Name,
			Cluster: clientcmd.Cluster{
				Server:                   inputCluster.Cluster.Server,
				TLSServerName:            inputCluster.Cluster.TLSServerName,
				InsecureSkipTLSVerify:    inputCluster.Cluster.InsecureSkipTLSVerify,
				CertificateAuthorityData: inputCluster.Cluster.CertificateAuthorityData,
				ProxyURL:                 inputCluster.Cluster.ProxyURL,
			},
		})
	}

	for _, inputAI := range inputConfig.AuthInfos {
		resultConfig.AuthInfos = append(resultConfig.AuthInfos, clientcmd.NamedAuthInfo{
			Name: inputAI.Name,
			AuthInfo: clientcmd.AuthInfo{
				ClientCertificateData: inputAI.AuthInfo.ClientCertificateData,
				ClientKeyData:         inputAI.AuthInfo.ClientKeyData,
				Token:                 inputAI.AuthInfo.Token,
				Impersonate:           inputAI.AuthInfo.Impersonate,
				ImpersonateGroups:     inputAI.AuthInfo.ImpersonateGroups,
				ImpersonateUserExtra:  inputAI.AuthInfo.ImpersonateUserExtra,
				Username:              inputAI.AuthInfo.Username,
				Password:              inputAI.AuthInfo.Password,
				AuthProvider:          inputAI.AuthInfo.AuthProvider,
			},
		})
	}

	for _, inputCtx := range inputConfig.Contexts {
		resultConfig.Contexts = append(resultConfig.Contexts, clientcmd.NamedContext{
			Name: inputCtx.Name,
			Context: clientcmd.Context{
				Cluster:   inputCtx.Context.Cluster,
				AuthInfo:  inputCtx.Context.AuthInfo,
				Namespace: inputCtx.Context.Namespace,
			},
		})
	}

	bs, err := yaml.Marshal(resultConfig)
	if err != nil {
		return nil, fmt.Errorf("Marshaling kubeconfig: %s", err)
	}

	return &KubeconfigRestricted{string(bs)}, nil
}

func (r *KubeconfigRestricted) AsYAML() string { return r.result }
