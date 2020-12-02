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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	"sipcluster/pkg/vbmh"
)

var _ = Describe("SIPCluster controller", func() {
	Context("When it detects a new SIPCluster", func() {
		It("Should schedule available nodes", func() {
			By("Labelling nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "master", "worker", "worker", "worker", "worker"}
			namespace := "default"
			for node, role := range nodes {
				vBMH, networkData := createBMH(node, namespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test1"
			sipCluster := createSIPCluster(clusterName, namespace, 3, 4)
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
						Name:      fmt.Sprintf("node%d", node),
						Namespace: namespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
		})

		It("Should not schedule nodes when there is an insufficient number of available master nodes", func() {
			By("Not labelling any nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "worker", "worker", "worker", "worker"}
			namespace := "default"
			for node, role := range nodes {
				vBMH, networkData := createBMH(node, namespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test2"
			sipCluster := createSIPCluster(clusterName, namespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					vbmh.SipScheduleLabel: "false",
				}

				var bmh metal3.BareMetalHost
				for node := range nodes {
					Expect(k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      fmt.Sprintf("node%d", node),
						Namespace: namespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
		})

		It("Should not schedule nodes when there is an insufficient number of available worker nodes", func() {
			By("Not labelling any nodes")

			// Create vBMH test objects
			nodes := []string{"master", "master", "master", "worker", "worker"}
			namespace := "default"
			for node, role := range nodes {
				vBMH, networkData := createBMH(node, namespace, role, 6)
				Expect(k8sClient.Create(context.Background(), vBMH)).Should(Succeed())
				Expect(k8sClient.Create(context.Background(), networkData)).Should(Succeed())
			}

			// Create SIP cluster
			clusterName := "subcluster-test3"
			sipCluster := createSIPCluster(clusterName, namespace, 3, 4)
			Expect(k8sClient.Create(context.Background(), sipCluster)).Should(Succeed())

			// Poll BMHs and validate they are not scheduled
			Consistently(func() error {
				expectedLabels := map[string]string{
					vbmh.SipScheduleLabel: "false",
				}

				var bmh metal3.BareMetalHost
				for node := range nodes {
					Expect(k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      fmt.Sprintf("node%d", node),
						Namespace: namespace,
					}, &bmh)).Should(Succeed())
				}

				return compareLabels(expectedLabels, bmh.GetLabels())
			}, 30, 5).Should(Succeed())

			cleanTestResources()
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

func createBMH(node int, namespace string, role string, rack int) (*metal3.BareMetalHost, *corev1.Secret) {
	rackLabel := fmt.Sprintf("r%d", rack)
	networkDataName := fmt.Sprintf("node%d-network-data", node)
	return &metal3.BareMetalHost{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("node%d", node),
				Namespace: namespace,
				Labels: map[string]string{
					"airshipit.org/vino-flavor": role,
					vbmh.SipScheduleLabel:       "false",
					vbmh.RackLabel:              rackLabel,
					vbmh.ServerLabel:            fmt.Sprintf("stl2%so%d", rackLabel, node),
				},
			},
			Spec: metal3.BareMetalHostSpec{
				NetworkData: &corev1.SecretReference{
					Namespace: namespace,
					Name:      networkDataName,
				},
			},
		}, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      networkDataName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"networkData": []byte("ewoKICAgICJsaW5rcyI6IFsKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJlbm80IiwKICAgICAgICAgICAgIm5hbWUiOiAiZW5vNCIsCiAgICAgICAgICAgICJ0eXBlIjogInBoeSIsCiAgICAgICAgICAgICJtdHUiOiAxNTAwCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJlbnA1OXMwZjEiLAogICAgICAgICAgICAibmFtZSI6ICJlbnA1OXMwZjEiLAogICAgICAgICAgICAidHlwZSI6ICJwaHkiLAogICAgICAgICAgICAibXR1IjogOTEwMAogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAiZW5wMjE2czBmMCIsCiAgICAgICAgICAgICJuYW1lIjogImVucDIxNnMwZjAiLAogICAgICAgICAgICAidHlwZSI6ICJwaHkiLAogICAgICAgICAgICAibXR1IjogOTEwMAogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAiYm9uZDAiLAogICAgICAgICAgICAibmFtZSI6ICJib25kMCIsCiAgICAgICAgICAgICJ0eXBlIjogImJvbmQiLAogICAgICAgICAgICAiYm9uZF9saW5rcyI6IFsKICAgICAgICAgICAgICAgICJlbnA1OXMwZjEiLAogICAgICAgICAgICAgICAgImVucDIxNnMwZjAiCiAgICAgICAgICAgIF0sCiAgICAgICAgICAgICJib25kX21vZGUiOiAiODAyLjNhZCIsCiAgICAgICAgICAgICJib25kX3htaXRfaGFzaF9wb2xpY3kiOiAibGF5ZXIzKzQiLAogICAgICAgICAgICAiYm9uZF9taWltb24iOiAxMDAsCiAgICAgICAgICAgICJtdHUiOiA5MTAwCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40MSIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQxIiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQxLAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40MiIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQyIiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQyLAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40NCIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQ0IiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQ0LAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40NSIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQ1IiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQ1LAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfQogICAgXSwKICAgICJuZXR3b3JrcyI6IFsKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJvYW0taXB2NiIsCiAgICAgICAgICAgICJ0eXBlIjogImlwdjYiLAogICAgICAgICAgICAibGluayI6ICJib25kMC40MSIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjIwMDE6MTg5MDoxMDAxOjI5M2Q6OjE0MCIsCiAgICAgICAgICAgICJyb3V0ZXMiOiBbCiAgICAgICAgICAgICAgICB7CiAgICAgICAgICAgICAgICAgICAgIm5ldHdvcmsiOiAiOjovMCIsCiAgICAgICAgICAgICAgICAgICAgIm5ldG1hc2siOiAiOjovMCIsCiAgICAgICAgICAgICAgICAgICAgImdhdGV3YXkiOiAiMjAwMToxODkwOjEwMDE6MjkzZDo6MSIKICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAib2FtLWlwdjQiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY0IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDEiLAogICAgICAgICAgICAiaXBfYWRkcmVzcyI6ICIzMi42OC41MS4xNDAiLAogICAgICAgICAgICAibmV0bWFzayI6ICIyNTUuMjU1LjI1NS4xMjgiLAogICAgICAgICAgICAiZG5zX25hbWVzZXJ2ZXJzIjogWwogICAgICAgICAgICAgICAgIjEzNS4xODguMzQuMTI0IiwKICAgICAgICAgICAgICAgICIxMzUuMzguMjQ0LjE2IiwKICAgICAgICAgICAgICAgICIxMzUuMTg4LjM0Ljg0IgogICAgICAgICAgICBdLAogICAgICAgICAgICAicm91dGVzIjogWwogICAgICAgICAgICAgICAgewogICAgICAgICAgICAgICAgICAgICJuZXR3b3JrIjogIjAuMC4wLjAiLAogICAgICAgICAgICAgICAgICAgICJuZXRtYXNrIjogIjAuMC4wLjAiLAogICAgICAgICAgICAgICAgICAgICJnYXRld2F5IjogIjMyLjY4LjUxLjEyOSIKICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAicHhlLWlwdjYiLAogICAgICAgICAgICAibGluayI6ICJlbm80IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NiIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogImZkMDA6OTAwOjEwMDoxMzg6OjEyIgogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAicHhlLWlwdjQiLAogICAgICAgICAgICAibGluayI6ICJlbm80IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NCIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjE3Mi4zMC4wLjEyIiwKICAgICAgICAgICAgIm5ldG1hc2siOiAiMjU1LjI1NS4yNTUuMTI4IgogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAic3RvcmFnZS1pcHY2IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDIiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY2IiwKICAgICAgICAgICAgImlwX2FkZHJlc3MiOiAiZmQwMDo5MDA6MTAwOjEzOTo6MTYiCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJzdG9yYWdlLWlwdjQiLAogICAgICAgICAgICAibGluayI6ICJib25kMC40MiIsCiAgICAgICAgICAgICJ0eXBlIjogImlwdjQiLAogICAgICAgICAgICAiaXBfYWRkcmVzcyI6ICIxNzIuMzEuMC4xNiIsCiAgICAgICAgICAgICJuZXRtYXNrIjogIjI1NS4yNTUuMjU1LjEyOCIKICAgICAgICB9LAogICAgICAgIHsKICAgICAgICAgICAgImlkIjogImtzbi1pcHY2IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDQiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY2IiwKICAgICAgICAgICAgImlwX2FkZHJlc3MiOiAiZmQwMDo5MDA6MTAwOjEzYTo6MTIiCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJrc24taXB2NCIsCiAgICAgICAgICAgICJsaW5rIjogImJvbmQwLjQ0IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NCIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjE3Mi4yOS4wLjEyIiwKICAgICAgICAgICAgIm5ldG1hc2siOiAiMjU1LjI1NS4yNTUuMTI4IgogICAgICAgIH0KICAgIF0KfQo="),
			},
			Type: corev1.SecretTypeOpaque,
		}
}

func createSIPCluster(name string, namespace string, masters int, workers int) *airshipv1.SIPCluster {
	return &airshipv1.SIPCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SIPCluster",
			APIVersion: "airship.airshipit.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: airshipv1.SIPClusterSpec{
			Config: &airshipv1.SipConfig{
				ClusterName: name,
			},
			Nodes: map[airshipv1.VmRoles]airshipv1.NodeSet{
				airshipv1.VmMaster: {
					VmFlavor:   "airshipit.org/vino-flavor=master",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VmCount{
						Active:  masters,
						Standby: 0,
					},
				},
				airshipv1.VmWorker: {
					VmFlavor:   "airshipit.org/vino-flavor=worker",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VmCount{
						Active:  workers,
						Standby: 0,
					},
				},
			},
			InfraServices: map[airshipv1.InfraService]airshipv1.InfraConfig{},
		},
		Status: airshipv1.SIPClusterStatus{},
	}
}
