package services_test

import (
	"context"
	"encoding/json"

	airshipv1 "sipcluster/pkg/api/v1"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sipcluster/pkg/services"
	"sipcluster/pkg/vbmh"
	"sipcluster/testutil"
)

const (
	ip1 = "192.168.0.1"
	ip2 = "192.168.0.2"
)

var bmh1 *metal3.BareMetalHost
var bmh2 *metal3.BareMetalHost

// Re-declared from services package for testing purposes
type host struct {
	Name string `json:"name"`
	BMC  bmc    `json:"bmc"`
}

type bmc struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var _ = Describe("Service Set", func() {
	Context("When new SIP cluster is created", func() {
		It("Deploys services", func() {
			By("Getting machine IPs and creating secrets, pods, and nodeport service")

			bmh1, _ = testutil.CreateBMH(1, "default", "control-plane", 1)
			bmh2, _ = testutil.CreateBMH(2, "default", "control-plane", 2)

			bmcUsername := "root"
			bmcPassword := "password"
			bmcSecret := testutil.CreateBMCAuthSecret(bmh1.GetName(), bmh1.GetNamespace(), bmcUsername,
				bmcPassword)
			Expect(k8sClient.Create(context.Background(), bmcSecret)).Should(Succeed())

			bmh1.Spec.BMC.CredentialsName = bmcSecret.Name
			bmh2.Spec.BMC.CredentialsName = bmcSecret.Name

			m1 := &vbmh.Machine{
				BMH: *bmh1,
				Data: &vbmh.MachineData{
					IPOnInterface: map[string]string{
						"eno3": ip1,
					},
				},
			}

			m2 := &vbmh.Machine{
				BMH: *bmh2,
				Data: &vbmh.MachineData{
					IPOnInterface: map[string]string{
						"eno3": ip2,
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

			serviceList, err := set.ServiceList()
			Expect(serviceList).To(HaveLen(2))
			Expect(err).To(Succeed())
			for _, svc := range serviceList {
				err := svc.Deploy()
				Expect(err).ToNot(HaveOccurred())
			}

			Eventually(func() error {
				return testDeployment(sip, *machineList)
			}, 5, 1).Should(Succeed())
		})
	})
})

func testDeployment(sip *airshipv1.SIPCluster, machineList vbmh.MachineList) error {
	loadBalancerDeployment := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.LoadBalancerServiceName + "-" + sip.GetName(),
	}, loadBalancerDeployment)
	if err != nil {
		return err
	}

	loadBalancerSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.LoadBalancerServiceName + "-" + sip.GetName(),
	}, loadBalancerSecret)
	if err != nil {
		return err
	}

	loadBalancerService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.LoadBalancerServiceName + "-" + sip.GetName(),
	}, loadBalancerService)
	if err != nil {
		return err
	}

	jumpHostDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.JumpHostServiceName + "-" + sip.GetName(),
	}, jumpHostDeployment)
	if err != nil {
		return err
	}

	jumpHostHostAliases := jumpHostDeployment.Spec.Template.Spec.HostAliases
	Expect(jumpHostHostAliases).To(ConsistOf(
		corev1.HostAlias{
			IP:        ip1,
			Hostnames: []string{bmh1.GetName()},
		},
		corev1.HostAlias{
			IP:        ip2,
			Hostnames: []string{bmh2.GetName()},
		},
	))

	jumpHostService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.JumpHostServiceName + "-" + sip.GetName(),
	}, jumpHostService)
	if err != nil {
		return err
	}

	jumpHostSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      services.JumpHostServiceName + "-" + sip.GetName(),
	}, jumpHostSecret)
	if err != nil {
		return err
	}

	var hosts []host
	err = json.Unmarshal(jumpHostSecret.Data["hosts"], &hosts)
	Expect(err).To(BeNil())
	for _, host := range hosts {
		for _, machine := range machineList.Machines {
			if host.Name == machine.BMH.Name {
				Expect(host.BMC.Username).To(Equal(machine.Data.BMCUsername))
				Expect(host.BMC.Password).To(Equal(machine.Data.BMCPassword))
			}
		}
	}

	return nil
}
