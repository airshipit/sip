/*


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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	"sipcluster/pkg/vbmh"
	"sipcluster/testutil"
)

const (
	testNamespace = "default"
)

var _ = Describe("SIPCluster controller", func() {
	Context("When it detects a new SIPCluster", func() {
		It("Should schedule available nodes", func() {
			By("Labeling nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "master", "worker", "worker", "worker", "worker"}
			for node, role := range nodes {
				vBMH, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test1"
			sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs until SIP has scheduled them to the SIP cluster
			Eventually(func() error {
				expectedLabels := map[string]string{
					vbmh.SipScheduleLabel: "true",
					vbmh.SipClusterLabel:  clusterName,
				}

				var bmh metal3.BareMetalHost
				for node := range nodes {
					Expect(k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      fmt.Sprintf("node0%d", node),
						Namespace: testNamespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
		})

		It("Should not schedule nodes when there is an insufficient number of available master nodes", func() {
			By("Not labeling any nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "worker", "worker", "worker", "worker"}
			for node, role := range nodes {
				vBMH, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test2"
			sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					vbmh.SipScheduleLabel: "false",
				}

				var bmh metal3.BareMetalHost
				for node := range nodes {
					Expect(k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      fmt.Sprintf("node0%d", node),
						Namespace: testNamespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
		})

		It("Should not schedule nodes when there is an insufficient number of available worker nodes", func() {
			By("Not labeling any nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "master", "worker", "worker"}
			testNamespace := "default"
			for node, role := range nodes {
				vBMH, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test4"
			sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					vbmh.SipScheduleLabel: "false",
				}

				var bmh metal3.BareMetalHost
				for node := range nodes {
					Expect(k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      fmt.Sprintf("node0%d", node),
						Namespace: testNamespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
		})

		Context("With per-node scheduling", func() {
			It("Should not schedule two worker nodes to the same server", func() {
				By("Not labeling any nodes")

				// Create vBMH test objects
				var nodes []*metal3.BareMetalHost
				baremetalServer := "r06o001"

				vBMH, networkData := testutil.CreateBMH(0, testNamespace, "master", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(1, testNamespace, "worker", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(2, testNamespace, "worker", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test5"
				sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 1, 2)
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						vbmh.SipScheduleLabel: "false",
					}

					var bmh metal3.BareMetalHost
					for node := range nodes {
						Expect(k8sClient.Get(context.Background(), types.NamespacedName{
							Name:      fmt.Sprintf("node0%d", node),
							Namespace: testNamespace,
						}, &bmh)).Should(Succeed())
					}

					return compareLabels(expectedLabels, bmh.GetLabels())
				}, 30, 5).Should(Succeed())

				cleanTestResources()
			})

			It("Should not schedule two master nodes to the same server", func() {
				By("Not labeling any nodes")

				// Create vBMH test objects
				var nodes []*metal3.BareMetalHost
				baremetalServer := "r06o001"

				vBMH, networkData := testutil.CreateBMH(0, testNamespace, "master", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(1, testNamespace, "master", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(2, testNamespace, "worker", 6)
				vBMH.Labels[vbmh.ServerLabel] = baremetalServer

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test6"
				sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 2, 1)
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						vbmh.SipScheduleLabel: "false",
					}

					var bmh metal3.BareMetalHost
					for node := range nodes {
						Expect(k8sClient.Get(context.Background(), types.NamespacedName{
							Name:      fmt.Sprintf("node0%d", node),
							Namespace: testNamespace,
						}, &bmh)).Should(Succeed())
					}

					return compareLabels(expectedLabels, bmh.GetLabels())
				}, 30, 5).Should(Succeed())

				cleanTestResources()
			})
		})

		Context("With per-rack scheduling", func() {
			It("Should not schedule two worker nodes to the same rack", func() {
				By("Not labeling any nodes")

				// Create vBMH test objects
				var nodes []*metal3.BareMetalHost
				testNamespace := "default"

				vBMH, networkData := testutil.CreateBMH(0, testNamespace, "master", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(1, testNamespace, "worker", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(2, testNamespace, "worker", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test3"
				sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 1, 2)

				masterSpec := sipCluster.Spec.Nodes[airshipv1.VMMaster]
				masterSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.VMMaster] = masterSpec

				workerSpec := sipCluster.Spec.Nodes[airshipv1.VMWorker]
				workerSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.VMWorker] = workerSpec

				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						vbmh.SipScheduleLabel: "false",
					}

					var bmh metal3.BareMetalHost
					for node := range nodes {
						Expect(k8sClient.Get(context.Background(), types.NamespacedName{
							Name:      fmt.Sprintf("node0%d", node),
							Namespace: testNamespace,
						}, &bmh)).Should(Succeed())
					}

					return compareLabels(expectedLabels, bmh.GetLabels())
				}, 30, 5).Should(Succeed())

				cleanTestResources()
			})

			It("Should not schedule two master nodes to the same rack", func() {
				By("Not labeling any nodes")

				// Create vBMH test objects
				var nodes []*metal3.BareMetalHost

				vBMH, networkData := testutil.CreateBMH(0, testNamespace, "master", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(1, testNamespace, "master", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				vBMH, networkData = testutil.CreateBMH(2, testNamespace, "worker", 6)

				nodes = append(nodes, vBMH)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test3"
				sipCluster := testutil.CreateSIPCluster(clusterName, testNamespace, 2, 1)

				masterSpec := sipCluster.Spec.Nodes[airshipv1.VMMaster]
				masterSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.VMMaster] = masterSpec

				workerSpec := sipCluster.Spec.Nodes[airshipv1.VMWorker]
				workerSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.VMWorker] = workerSpec

				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						vbmh.SipScheduleLabel: "false",
					}

					var bmh metal3.BareMetalHost
					for node := range nodes {
						Expect(k8sClient.Get(context.Background(), types.NamespacedName{
							Name:      fmt.Sprintf("node0%d", node),
							Namespace: testNamespace,
						}, &bmh)).Should(Succeed())
					}

					return compareLabels(expectedLabels, bmh.GetLabels())
				}, 30, 5).Should(Succeed())

				cleanTestResources()
			})
		})
	})
})

func compareLabels(expected map[string]string, actual map[string]string) error {
	for k, v := range expected {
		value, exists := actual[k]
		if !exists {
			return fmt.Errorf("label %s=%s missing. Has labels %v", k, v, actual)
		}

		if value != v {
			return fmt.Errorf("label %s=%s does not match expected label %s=%s. Has labels %v", k, value, k,
				v, actual)
		}
	}

	return nil
}

func cleanTestResources() {
	opts := []client.DeleteAllOfOption{client.InNamespace("default")}
	Expect(k8sClient.DeleteAllOf(context.Background(), &metal3.BareMetalHost{}, opts...)).Should(Succeed())
	Expect(k8sClient.DeleteAllOf(context.Background(), &airshipv1.SIPCluster{}, opts...)).Should(Succeed())
	Expect(k8sClient.DeleteAllOf(context.Background(), &corev1.Secret{}, opts...)).Should(Succeed())
}
