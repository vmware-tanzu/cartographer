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

package satoken_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/satoken"
	"github.com/vmware-tanzu/cartographer/pkg/satoken/satokenfakes"
)

var _ = Describe("TokenManager", func() {
	var (
		tokenManager                satoken.TokenManager
		fakeClient                  *satokenfakes.FakeInterface
		fakeCoreInterface           *satokenfakes.FakeCoreV1Interface
		fakeServiceAccountInterface *satokenfakes.FakeServiceAccountInterface
		fakeLogger                  *satokenfakes.FakeLogger
		serviceAccount              *v1.ServiceAccount
		stubbedExpirationTime       *metav1.Time
		tokenCache                  map[string]*authenticationv1.TokenRequest
		maxTTL                      = 2 * time.Hour
	)

	BeforeEach(func() {
		fakeLogger = &satokenfakes.FakeLogger{}
		fakeServiceAccountInterface = &satokenfakes.FakeServiceAccountInterface{}

		stubbedExpirationTime = nil
		fakeServiceAccountInterface.CreateTokenStub = func(_ context.Context, saName string, request *authenticationv1.TokenRequest, options metav1.CreateOptions) (*authenticationv1.TokenRequest, error) {
			response := *request
			response.Status = authenticationv1.TokenRequestStatus{
				Token: "abracadabra" + fmt.Sprintf("%d", fakeServiceAccountInterface.CreateTokenCallCount()),
			}
			if stubbedExpirationTime != nil {
				response.Status.ExpirationTimestamp = *stubbedExpirationTime
			} else {
				response.Status.ExpirationTimestamp = metav1.Time{
					Time: time.Now().Add(maxTTL),
				}
			}
			return &response, nil
		}

		fakeCoreInterface = &satokenfakes.FakeCoreV1Interface{}
		fakeCoreInterface.ServiceAccountsReturns(fakeServiceAccountInterface)

		fakeClient = &satokenfakes.FakeInterface{}
		fakeClient.CoreV1Returns(fakeCoreInterface)

		tokenCache = make(map[string]*authenticationv1.TokenRequest)
		tokenManager = satoken.NewManager(fakeClient, fakeLogger, tokenCache)

		serviceAccount = &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sa",
				Namespace: "my-namespace",
				UID:       "sa-uid",
			},
		}
	})

	Describe("GetServiceAccountToken", func() {
		It("obtains a token with 2 hours TTL from the service account in the specified namespace", func() {
			token, err := tokenManager.GetServiceAccountToken(serviceAccount)

			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("abracadabra1"))

			Expect(fakeClient.CoreV1CallCount()).To(Equal(1))
			Expect(fakeCoreInterface.ServiceAccountsCallCount()).To(Equal(1))
			Expect(fakeCoreInterface.ServiceAccountsArgsForCall(0)).To(Equal("my-namespace"))
			Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(1))
			_, saName, tr, createOptions := fakeServiceAccountInterface.CreateTokenArgsForCall(0)
			Expect(saName).To(Equal("my-sa"))
			Expect(createOptions).To(Equal(metav1.CreateOptions{}))
			secondsInTwoHours := int64(maxTTL.Seconds())
			Expect(tr).To(Equal(&authenticationv1.TokenRequest{
				Spec: authenticationv1.TokenRequestSpec{
					ExpirationSeconds: &secondsInTwoHours,
				}}))
		})

		It("caches tokens for the same service account", func() {
			token, err := tokenManager.GetServiceAccountToken(serviceAccount)
			Expect(err).NotTo(HaveOccurred())

			probablyCachedToken, err := tokenManager.GetServiceAccountToken(serviceAccount)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(1))
			Expect(token).To(Equal(probablyCachedToken))
		})

		Describe("cache identity", func() {
			It("is different for differently named service accounts", func() {
				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				differentServiceAccount := *serviceAccount
				differentServiceAccount.ObjectMeta.Name = "different-sa-name"
				differentToken, err := tokenManager.GetServiceAccountToken(&differentServiceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(token).NotTo(Equal(differentToken))
			})

			It("is different for similarly named service accounts in different namespaces", func() {
				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				differentServiceAccount := *serviceAccount
				differentServiceAccount.ObjectMeta.Namespace = "outta-this-world-ns"
				differentToken, err := tokenManager.GetServiceAccountToken(&differentServiceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(token).NotTo(Equal(differentToken))
			})

			It("is different for a service account with the same name and namespace, but a different UID", func() {
				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				differentServiceAccount := *serviceAccount
				differentServiceAccount.ObjectMeta.UID = "hocus-pocus"
				differentToken, err := tokenManager.GetServiceAccountToken(&differentServiceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(token).NotTo(Equal(differentToken))
			})
		})

		Describe("token refreshing", func() {
			It("does not attempt to refresh tokens with an age less than half the TTL", func() {
				stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(maxTTL / 2).Add(time.Minute)}

				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				maybeCachedToken, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(1))
				Expect(token).To(Equal(maybeCachedToken))
			})

			It("refreshes tokens with age greater than half the TTL", func() {
				stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(maxTTL / 2).Add(-1 * time.Minute)}
				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				maybeCachedToken, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(2))
				Expect(token).NotTo(Equal(maybeCachedToken))
			})

			It("returns non-expired tokens when refresh fails, silently logging the error", func() {
				stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(maxTTL / 2).Add(-1 * time.Minute)}
				token, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				fakeServiceAccountInterface.CreateTokenReturns(nil, fmt.Errorf("creating tokens is hard"))
				maybeCachedToken, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(2))
				Expect(token).To(Equal(maybeCachedToken))

				Expect(fakeLogger.ErrorCallCount()).To(Equal(1))
				loggedErr, logMessage, keysAndValues := fakeLogger.ErrorArgsForCall(0)
				Expect(loggedErr).To(MatchError("creating tokens is hard"))
				Expect(logMessage).To(Equal("update token"))
				Expect(keysAndValues).To(Equal([]interface{}{"cacheKey", `"my-sa"/"my-namespace"/"sa-uid"`}))
			})

			It("returns an error if the token is expired and refresh fails", func() {
				stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(-1 * time.Minute)}
				_, err := tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).NotTo(HaveOccurred())

				fakeServiceAccountInterface.CreateTokenReturns(nil, fmt.Errorf("creating tokens is hard"))
				_, err = tokenManager.GetServiceAccountToken(serviceAccount)
				Expect(err).To(MatchError(`token "my-sa"/"my-namespace"/"sa-uid" expired and refresh failed: creating tokens is hard`))

				Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(2))
			})
		})
	})

	Describe("Cleanup", func() {
		It("evicts expired tokens from the cache", func() {
			stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(-1 * time.Minute)}
			_, err := tokenManager.GetServiceAccountToken(serviceAccount)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(1))

			Expect(tokenCache).To(HaveLen(1))
			tokenManager.Cleanup()
			Expect(tokenCache).To(HaveLen(0))
		})

		It("does not evict unexpired tokens from the cache", func() {
			stubbedExpirationTime = &metav1.Time{Time: time.Now().Add(time.Minute)}
			_, err := tokenManager.GetServiceAccountToken(serviceAccount)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeServiceAccountInterface.CreateTokenCallCount()).To(Equal(1))

			Expect(tokenCache).To(HaveLen(1))
			tokenManager.Cleanup()
			Expect(tokenCache).To(HaveLen(1))
		})
	})
})
