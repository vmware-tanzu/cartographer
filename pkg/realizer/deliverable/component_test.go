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

package deliverable_test

//
//var _ = Describe("Resource", func() {
//
//	var (
//		ctx                      context.Context
//		resource                 v1alpha1.DeliveryResource
//		deliverable              v1alpha1.Deliverable
//		outputs                  realizer.Outputs
//		deliveryName             string
//		fakeSystemRepo           repositoryfakes.FakeRepository
//		fakeDeliverableRepo      repositoryfakes.FakeRepository
//		clientForBuiltRepository client.Client
//		cacheForBuiltRepository  repository.RepoCache
//		repoCache                repository.RepoCache
//		builtClient              client.Client
//		theSecret                *corev1.Secret
//		secretForBuiltClient     *corev1.Secret
//		r                        realizer.ResourceRealizer
//		deliveryParams           []v1alpha1.BlueprintParam
//	)
//
//	BeforeEach(func() {
//		ctx = context.Background()
//		resource = v1alpha1.DeliveryResource{
//			Name: "resource-1",
//			TemplateRef: v1alpha1.DeliveryTemplateReference{
//				Kind: "ClusterSourceTemplate",
//				Name: "source-template-1",
//			},
//		}
//
//		deliveryName = "delivery-name"
//
//		deliveryParams = []v1alpha1.BlueprintParam{}
//
//		outputs = realizer.NewOutputs()
//
//		fakeSystemRepo = repositoryfakes.FakeRepository{}
//		fakeDeliverableRepo = repositoryfakes.FakeRepository{}
//
//		repositoryBuilder := func(client client.Client, repoCache repository.RepoCache) repository.Repository {
//			clientForBuiltRepository = client
//			cacheForBuiltRepository = repoCache
//			return &fakeDeliverableRepo
//		}
//
//		builtClient = &repositoryfakes.FakeClient{}
//		clientBuilder := func(secret *corev1.Secret, _ bool) (client.Client, discovery.DiscoveryInterface, error) {
//			secretForBuiltClient = secret
//			return builtClient, nil, nil
//		}
//
//		repoCache = &repositoryfakes.FakeRepoCache{} //TODO: can we verify right cache used?
//		resourceRealizerBuilder := realizer.NewResourceRealizerBuilder(repositoryBuilder, clientBuilder, repoCache)
//
//		deliverable = v1alpha1.Deliverable{}
//
//		theSecret = &corev1.Secret{StringData: map[string]string{"blah": "blah"}}
//
//		var err error
//		r, err = resourceRealizerBuilder(theSecret, &deliverable, &fakeSystemRepo, deliveryParams)
//		Expect(err).NotTo(HaveOccurred())
//	})
//
//	It("creates a resource realizer with the existing client, as well as one with the the supplied secret mixed in", func() {
//		Expect(secretForBuiltClient).To(Equal(theSecret))
//		Expect(clientForBuiltRepository).To(Equal(builtClient))
//	})
//
//	It("creates a resource realizer with the existing cache", func() {
//		Expect(cacheForBuiltRepository).To(Equal(repoCache))
//	})
//
//	Describe("Do", func() {
//
//	})
//})
