// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	realizerclient "github.com/vmware-tanzu/cartographer/pkg/realizer/client"
)

var _ = Describe("Pkg/Realizer/Client/Client", func() {
	Describe("AddBearerToken", func() {
		var (
			oldConfig *rest.Config
			secret    *corev1.Secret
			newToken  string
			oldToken  string
			tokenFile string
		)

		BeforeEach(func() {
			oldToken = "some-old-token"
			tokenFile = "some-file-path"
			oldConfig = &rest.Config{
				Host:            "some-host",
				BearerToken:     oldToken,
				BearerTokenFile: tokenFile,
			}

			newToken = "some-new-token"

			secret = &corev1.Secret{
				Data: map[string][]byte{
					corev1.ServiceAccountTokenKey: []byte(newToken),
				},
			}
		})

		It("overwrites the BearerToken in the config and removes the BearerTokenFile (because it supersedes BearerToken)", func() {
			newConfig, err := realizerclient.AddBearerToken(secret, oldConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(newConfig.BearerToken).To(Equal(newToken))
			Expect(newConfig.BearerTokenFile).To(Equal(""))
		})

		It("preserves the rest of the config", func() {
			newConfig, err := realizerclient.AddBearerToken(secret, oldConfig)
			Expect(err).NotTo(HaveOccurred())

			newConfig.BearerToken = oldToken
			newConfig.BearerTokenFile = tokenFile

			Expect(newConfig).To(Equal(oldConfig))
		})
	})
})
