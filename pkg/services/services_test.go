package services_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	airshipv1 "sipcluster/pkg/api/v1"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sipcluster/pkg/bmh"
	"sipcluster/pkg/services"
	"sipcluster/testutil"
)

const (
	ip1 = "192.168.0.1"
	ip2 = "192.168.0.2"
)

var bmh1 *metal3.BareMetalHost
var bmh2 *metal3.BareMetalHost

var m1 *bmh.Machine
var m2 *bmh.Machine

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
	var machineList *bmh.MachineList
	BeforeEach(func() {
		bmh1, _ = testutil.CreateBMH(1, "default", "control-plane", 1)
		bmh2, _ = testutil.CreateBMH(2, "default", "control-plane", 2)

		bmcUsername := "root"
		bmcPassword := "password"
		bmcSecret := testutil.CreateBMCAuthSecret(bmh1.GetName(), bmh1.GetNamespace(), bmcUsername,
			bmcPassword)
		Expect(k8sClient.Create(context.Background(), bmcSecret)).Should(Succeed())

		bmh1.Spec.BMC.CredentialsName = bmcSecret.Name
		bmh2.Spec.BMC.CredentialsName = bmcSecret.Name

		m1 = &bmh.Machine{
			BMH: *bmh1,
			Data: &bmh.MachineData{
				IPOnInterface: map[string]string{
					"oam-ipv4": ip1,
				},
			},
		}

		m2 = &bmh.Machine{
			BMH: *bmh2,
			Data: &bmh.MachineData{
				IPOnInterface: map[string]string{
					"oam-ipv4": ip2,
				},
			},
		}

		machineList = &bmh.MachineList{
			Machines: map[string]*bmh.Machine{
				bmh1.GetName(): m1,
				bmh2.GetName(): m2,
			},
		}

		//Secret for Template
		TemplateControlPlane, err := ioutil.ReadFile("../../config/manager/loadbalancer/loadBalancerControlPlane.cfg")
		if err == nil {
			lbcontrolplaneTemplateConfigMap := testutil.CreateTemplateConfigMap("loadbalancercontrolplane",
				"loadBalancerControlPlane.cfg", "default", string(TemplateControlPlane))
			Expect(k8sClient.Create(context.Background(), lbcontrolplaneTemplateConfigMap)).Should(Succeed())
		}

		TemplateWorker, err := ioutil.ReadFile("../../config/manager/loadbalancer/loadBalancerWorker.cfg")

		if err == nil {
			lbworkerTemplateConfigMap := testutil.CreateTemplateConfigMap("loadbalancerworker",
				"loadBalancerWorker.cfg", "default", string(TemplateWorker))
			Expect(k8sClient.Create(context.Background(), lbworkerTemplateConfigMap)).Should(Succeed())
		}

	})

	AfterEach(func() {
		opts := []client.DeleteAllOfOption{client.InNamespace("default")}
		Expect(k8sClient.DeleteAllOf(context.Background(), &metal3.BareMetalHost{}, opts...)).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &airshipv1.SIPCluster{}, opts...)).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &corev1.Secret{}, opts...)).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(context.Background(), &corev1.ConfigMap{}, opts...)).Should(Succeed())
	})

	Context("When new SIP cluster is created", func() {
		It("Deploys services", func() {
			By("Getting machine IPs and creating secrets, pods, and nodeport service")

			sipCluster, nodeSSHPrivateKeys := testutil.CreateSIPCluster("default", "default", 1, 1)
			Expect(k8sClient.Create(context.Background(), nodeSSHPrivateKeys)).Should(Succeed())
			machineList = &bmh.MachineList{
				Machines: map[string]*bmh.Machine{
					bmh1.GetName(): m1,
					bmh2.GetName(): m2,
				},
			}

			set := services.NewServiceSet(logger, *sipCluster, machineList, k8sClient)

			serviceList, err := set.ServiceList()
			Expect(serviceList).To(HaveLen(3))
			Expect(err).To(Succeed())
			for _, svc := range serviceList {
				err := svc.Deploy()
				Expect(err).ToNot(HaveOccurred())
			}

			Eventually(func() error {
				return testDeployment(sipCluster, *machineList)
			}, 5, 1).Should(Succeed())
		})

		It("Does not deploy a Jump Host when an invalid SSH key is provided", func() {
			sip, _ := testutil.CreateSIPCluster("default", "default", 1, 1)
			sip.Spec.Services.LoadBalancerControlPlane = []airshipv1.LoadBalancerServiceControlPlane{}
			sip.Spec.Services.LoadBalancerWorker = []airshipv1.LoadBalancerServiceWorker{}
			sip.Spec.Services.JumpHost[0].SSHAuthorizedKeys = []string{
				"sshrsaAAAAAAAAAAAAAAAAAAAAAinvalidkey",
			}

			set := services.NewServiceSet(logger, *sip, machineList, k8sClient)
			serviceList, err := set.ServiceList()
			Expect(err).To(Succeed())

			for _, svc := range serviceList {
				err := svc.Deploy()
				Expect(err).To(HaveOccurred())
			}
		})

	})

})

func testDeployment(sip *airshipv1.SIPCluster, machineList bmh.MachineList) error {
	loadBalancerControlPlaneDeployment := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleControlPlane)) + "-" +
			sip.GetName(),
	}, loadBalancerControlPlaneDeployment)
	if err != nil {
		return err
	}

	loadBalancerWorkerDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleWorker)) + "-" +
			sip.GetName(),
	}, loadBalancerWorkerDeployment)
	if err != nil {
		return err
	}

	loadBalancerControlPlaneSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleControlPlane)) + "-" +
			sip.GetName(),
	}, loadBalancerControlPlaneSecret)
	if err != nil {
		return err
	}

	loadBalancerWorkerSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleWorker)) + "-" +
			sip.GetName(),
	}, loadBalancerWorkerSecret)
	if err != nil {
		return err
	}

	loadBalancerControlPlaneService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleControlPlane)) + "-" +
			sip.GetName(),
	}, loadBalancerControlPlaneService)
	if err != nil {
		return err
	}

	loadBalancerWorkerService := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name: services.LoadBalancerServiceName + "-" + strings.ToLower(string(airshipv1.RoleWorker)) + "-" +
			sip.GetName(),
	}, loadBalancerWorkerService)
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
