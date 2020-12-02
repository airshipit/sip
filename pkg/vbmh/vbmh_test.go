package vbmh

import (
	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	mockClient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	airshipv1 "sipcluster/pkg/api/v1"
	"sipcluster/testutil"
)

const (
	// numNodes is the number of test vBMH objects (nodes) created for each test
	numNodes = 7
)

var _ = Describe("MachineList", func() {
	var machineList *MachineList
	BeforeEach(func() {
		nodes := map[string]*Machine{}
		for n := 0; n < numNodes; n++ {
			bmh, _ := testutil.CreateBMH(n, "default", "master", 6)
			nodes[bmh.Name] = NewMachine(*bmh, airshipv1.VmMaster, NotScheduled)
		}

		machineList = &MachineList{
			NamespacedName: types.NamespacedName{
				Name:      "vbmh",
				Namespace: "default",
			},
			Machines: nodes,
			Log:      ctrl.Log.WithName("controllers").WithName("SIPCluster"),
		}

		err := metal3.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should report if it has a machine registered for a BMH object", func() {
		for _, bmh := range machineList.Machines {
			Expect(machineList.hasMachine(bmh.BMH)).Should(BeTrue())
		}

		unregisteredMachine := machineList.Machines["node01"]
		unregisteredMachine.BMH.Name = "foo"
		Expect(machineList.hasMachine(unregisteredMachine.BMH)).Should(BeFalse())
	})

	It("Should produce a list of unscheduled BMH objects", func() {
		// "Schedule" two nodes
		machineList.Machines["node00"].BMH.Labels[SipScheduleLabel] = "true"
		machineList.Machines["node01"].BMH.Labels[SipScheduleLabel] = "true"
		scheduledNodes := []metal3.BareMetalHost{
			machineList.Machines["node00"].BMH,
			machineList.Machines["node01"].BMH,
		}

		var objs []runtime.Object
		for _, machine := range machineList.Machines {
			objs = append(objs, &machine.BMH)
		}

		k8sClient := mockClient.NewFakeClient(objs...)
		bmhList, err := machineList.getBMHs(k8sClient)
		Expect(err).To(BeNil())

		// Validate that the BMH list does not contain scheduled nodes
		for _, bmh := range bmhList.Items {
			for _, scheduled := range scheduledNodes {
				Expect(bmh).ToNot(Equal(scheduled))
				Expect(bmh.Labels[SipScheduleLabel]).To(Equal("false"))
			}
		}

	})

})
