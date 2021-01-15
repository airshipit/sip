package services_test

import (
	"context"

	airshipv1 "sipcluster/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sipcluster/pkg/services"
	"sipcluster/pkg/vbmh"
	"sipcluster/testutil"
)

var _ = Describe("Service Set", func() {
	Context("When new SIP cluster is created", func() {
		It("Deploys services", func() {
			By("Getting machine IPs and creating secrets, pods, and nodeport service")

			bmh1, _ := testutil.CreateBMH(1, "default", "control-plane", 1)
			bmh2, _ := testutil.CreateBMH(2, "default", "control-plane", 2)
			m1 := &vbmh.Machine{
				BMH: *bmh1,
				Data: &vbmh.MachineData{
					IPOnInterface: map[string]string{
						"eno3": "192.168.0.1",
					},
				},
			}
			m2 := &vbmh.Machine{
				BMH: *bmh2,
				Data: &vbmh.MachineData{
					IPOnInterface: map[string]string{
						"eno3": "192.168.0.2",
					},
				},
			}

			sip := testutil.CreateSIPCluster("default", "default", 1, 1)
			machineList := &vbmh.MachineList{
				Machines: map[string]*vbmh.Machine{
					bmh1.GetName(): m1,
					bmh2.GetName(): m2,
				},
			}

			set := services.NewServiceSet(logger, *sip, machineList, k8sClient)

			serviceList := set.ServiceList()
			Expect(serviceList).To(HaveLen(2))
			for _, svc := range serviceList {
				err := svc.Deploy()
				Expect(err).ToNot(HaveOccurred())
			}

			Eventually(func() error {
				return testDeployment(sip)
			}, 5, 1).Should(Succeed())
		})
	})
})

func testDeployment(sip *airshipv1.SIPCluster) error {
	jumpHostPod := &corev1.Pod{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-jump-pod",
	}, jumpHostPod)
	if err != nil {
		return err
	}

	loadBalancerPod := &corev1.Pod{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer",
	}, loadBalancerPod)
	if err != nil {
		return err
	}

	loadBalancerSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer",
	}, loadBalancerSecret)
	if err != nil {
		return err
	}

	loadBalancerService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer-service",
	}, loadBalancerService)
	if err != nil {
		return err
	}

	return nil
}
