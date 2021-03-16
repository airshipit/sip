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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	bmhpkg "sipcluster/pkg/bmh"
	"sipcluster/testutil"
)

const (
	testNamespace = "default"
)

var _ = Describe("SIPCluster controller", func() {

	AfterEach(func() {
		opts := []client.DeleteAllOfOption{client.InNamespace(testNamespace)}
		Expect(k8sClient.DeleteAllOf(context.Background(), &metal3.BareMetalHost{}, opts...)).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &airshipv1.SIPCluster{}, opts...)).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &corev1.Secret{}, opts...)).Should(Succeed())
	})

	Context("When it detects a new SIPCluster", func() {
		It("Should schedule available nodes", func() {
			By("Labeling nodes")

			// Create BMH test objects
			nodes := []airshipv1.BMHRole{airshipv1.RoleControlPlane, airshipv1.RoleControlPlane, airshipv1.RoleControlPlane,
				airshipv1.RoleWorker, airshipv1.RoleWorker, airshipv1.RoleWorker, airshipv1.RoleWorker}
			bmcUsername := "root"
			bmcPassword := "test"
			for node, role := range nodes {
				bmh, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				bmcSecret := testutil.CreateBMCAuthSecret(bmh.Name, bmh.Namespace, bmcUsername,
					bmcPassword)

				bmh.Spec.BMC.CredentialsName = bmcSecret.Name

				Expect(k8sClient.Create(context.Background(), bmcSecret)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

			}

			// Create SIP cluster
			clusterName := "subcluster-test1"
			sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs until SIP has scheduled them to the SIP cluster
			Eventually(func() error {
				expectedLabels := map[string]string{
					bmhpkg.SipScheduleLabel: "true",
					bmhpkg.SipClusterLabel:  bmhpkg.GetClusterLabel(*sipCluster),
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
		})

		It("Should not schedule nodes when there is an insufficient number of available ControlPlane nodes", func() {
			By("Not labeling any nodes")

			// Create BMH test objects
			nodes := []airshipv1.BMHRole{airshipv1.RoleControlPlane, airshipv1.RoleControlPlane, airshipv1.RoleWorker,
				airshipv1.RoleWorker, airshipv1.RoleWorker, airshipv1.RoleWorker}
			for node, role := range nodes {
				bmh, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test2"
			sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					bmhpkg.SipScheduleLabel: "false",
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

			// Validate SIP CR ready condition has been updated
			var sipCR airshipv1.SIPCluster
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      clusterName,
				Namespace: testNamespace,
			}, &sipCR)).To(Succeed())

			Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
				airshipv1.ConditionTypeReady)).To(BeTrue())
		})

		It("Should not schedule nodes when there is an insufficient number of available Worker nodes", func() {
			By("Not labeling any nodes")

			// Create BMH test objects
			nodes := []airshipv1.BMHRole{airshipv1.RoleControlPlane, airshipv1.RoleControlPlane, airshipv1.RoleControlPlane,
				airshipv1.RoleWorker, airshipv1.RoleWorker}
			testNamespace := "default"
			for node, role := range nodes {
				bmh, networkData := testutil.CreateBMH(node, testNamespace, role, 6)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test4"
			sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					bmhpkg.SipScheduleLabel: "false",
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

			// Validate SIP CR ready condition has been updated
			var sipCR airshipv1.SIPCluster
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      clusterName,
				Namespace: testNamespace,
			}, &sipCR)).To(Succeed())

			Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
				airshipv1.ConditionTypeReady)).To(BeTrue())
		})

		Context("With per-node scheduling", func() {
			It("Should not schedule two Worker nodes to the same server", func() {
				By("Not labeling any nodes")

				// Create BMH test objects
				var nodes []*metal3.BareMetalHost
				baremetalServer := "r06o001"

				bmh, networkData := testutil.CreateBMH(0, testNamespace, airshipv1.RoleControlPlane, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(1, testNamespace, airshipv1.RoleWorker, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(2, testNamespace, airshipv1.RoleWorker, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test5"
				sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 1, 2)
				Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						bmhpkg.SipScheduleLabel: "false",
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

				// Validate SIP CR ready condition has been updated
				var sipCR airshipv1.SIPCluster
				Expect(k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      clusterName,
					Namespace: testNamespace,
				}, &sipCR)).To(Succeed())

				Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
					airshipv1.ConditionTypeReady)).To(BeTrue())
			})

			It("Should not schedule two ControlPlane nodes to the same server", func() {
				By("Not labeling any nodes")

				// Create BMH test objects
				var nodes []*metal3.BareMetalHost
				baremetalServer := "r06o001"

				bmh, networkData := testutil.CreateBMH(0, testNamespace, airshipv1.RoleControlPlane, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(1, testNamespace, airshipv1.RoleControlPlane, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(2, testNamespace, airshipv1.RoleWorker, 6)
				bmh.Labels[bmhpkg.ServerLabel] = baremetalServer

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test6"
				sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 2, 1)
				Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						bmhpkg.SipScheduleLabel: "false",
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

				// Validate SIP CR ready condition has been updated
				var sipCR airshipv1.SIPCluster
				Expect(k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      clusterName,
					Namespace: testNamespace,
				}, &sipCR)).To(Succeed())

				Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
					airshipv1.ConditionTypeReady)).To(BeTrue())
			})
		})

		Context("With per-rack scheduling", func() {
			It("Should not schedule two Worker nodes to the same rack", func() {
				By("Not labeling any nodes")

				// Create BMH test objects
				var nodes []*metal3.BareMetalHost
				testNamespace := "default"

				bmh, networkData := testutil.CreateBMH(0, testNamespace, airshipv1.RoleControlPlane, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(1, testNamespace, airshipv1.RoleWorker, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(2, testNamespace, airshipv1.RoleWorker, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test3"
				sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 1, 2)

				controlPlaneSpec := sipCluster.Spec.Nodes[airshipv1.RoleControlPlane]
				controlPlaneSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.RoleControlPlane] = controlPlaneSpec

				workerSpec := sipCluster.Spec.Nodes[airshipv1.RoleWorker]
				workerSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.RoleWorker] = workerSpec

				Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						bmhpkg.SipScheduleLabel: "false",
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

				// Validate SIP CR ready condition has been updated
				var sipCR airshipv1.SIPCluster
				Expect(k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      clusterName,
					Namespace: testNamespace,
				}, &sipCR)).To(Succeed())

				Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
					airshipv1.ConditionTypeReady)).To(BeTrue())
			})

			It("Should not schedule two ControlPlane nodes to the same rack", func() {
				By("Not labeling any nodes")

				// Create BMH test objects
				var nodes []*metal3.BareMetalHost

				bmh, networkData := testutil.CreateBMH(0, testNamespace, airshipv1.RoleControlPlane, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(1, testNamespace, airshipv1.RoleControlPlane, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				bmh, networkData = testutil.CreateBMH(2, testNamespace, airshipv1.RoleWorker, 6)

				nodes = append(nodes, bmh)
				Expect(k8sClient.Create(context.Background(), bmh)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())

				// Create SIP cluster
				clusterName := "subcluster-test3"
				sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster(clusterName, testNamespace, 2, 1)

				controlPlaneSpec := sipCluster.Spec.Nodes[airshipv1.RoleControlPlane]
				controlPlaneSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.RoleControlPlane] = controlPlaneSpec

				workerSpec := sipCluster.Spec.Nodes[airshipv1.RoleWorker]
				workerSpec.Scheduling = airshipv1.RackAntiAffinity
				sipCluster.Spec.Nodes[airshipv1.RoleWorker] = workerSpec

				Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

				// Poll BMHs and validate they are not scheduled
				Consistently(func() error {
					expectedLabels := map[string]string{
						bmhpkg.SipScheduleLabel: "false",
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

				// Validate SIP CR ready condition has been updated
				var sipCR airshipv1.SIPCluster
				Expect(k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      clusterName,
					Namespace: testNamespace,
				}, &sipCR)).To(Succeed())

				Expect(apimeta.IsStatusConditionFalse(sipCR.Status.Conditions,
					airshipv1.ConditionTypeReady)).To(BeTrue())

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
