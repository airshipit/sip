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
			Expect(serviceList).To(HaveLen(1))
			Eventually(func() error {
				return testDeployment(serviceList[0], sip)
			}, 5, 1).Should(Succeed())
		})
	})
})

func testDeployment(sl services.InfraService, sip *airshipv1.SIPCluster) error {
	err := sl.Deploy()
	if err != nil {
		return err
	}

	pod := &corev1.Pod{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer",
	}, pod)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer",
	}, secret)
	if err != nil {
		return err
	}

	service := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      sip.GetName() + "-load-balancer-service",
	}, service)
	if err != nil {
		return err
	}
	return nil
}
