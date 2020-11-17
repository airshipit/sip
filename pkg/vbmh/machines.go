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

package vbmh

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	airshipv1 "sipcluster/pkg/api/v1"

	"github.com/PaesslerAG/jsonpath"
	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScheduledState
type ScheduledState string

// Possible Node or VM Roles  for a Tenant
const (
	// ToBeScheduled means that the VM was identified by the scheduler to be selected
	ToBeScheduled ScheduledState = "Selected"

	// Scheduled means the BMH / VM already has a label implying it
	// was previously scheduled
	Scheduled ScheduledState = "Scheduled"

	// NotScheduled, This BMH was not selected to be scheduled.
	// Either because we didnt meet the criteria
	// or simple we reached the count
	NotScheduled ScheduledState = "NotSelected"

	// UnableToSchedule, This BMH has something wrong with it
	// The BMH itself doesnt depict the error situation
	// i.e. teh NetworkData is missing something
	UnableToSchedule ScheduledState = "UnableToSchedule"
)

const (
	BaseAirshipSelector  = "sip.airshipit.org"
	SipScheduleLabelName = "sip-scheduled"
	SipScheduleLabel     = BaseAirshipSelector + "/" + SipScheduleLabelName

	SipScheduled    = SipScheduleLabel + "=true"
	SipNotScheduled = SipScheduleLabel + "=false"

	// This is a placeholder . Need to synchronize with ViNO the constants below
	// Probable pll this or eqivakent values from a ViNO pkg
	RackLabel   = BaseAirshipSelector + "/rack"
	ServerLabel = BaseAirshipSelector + "/server"

	// Thislabekl is associated to group the colletcion of scheduled vBMH's
	// Will represent the Tenant Cluster or Service Function  Cluster
	SipClusterLabelName = "workload-cluster"
	SipClusterLabel     = BaseAirshipSelector + "/" + SipClusterLabelName

	SipNodeTypeLabelName = "note-type"
	SipNodeTypeLabel     = BaseAirshipSelector + "/" + SipNodeTypeLabelName
)

// MAchine represents an individual BMH CR, and teh appropriate
// attributes required to manage the SIP Cluster scheduling and
// rocesing needs about thhem
type Machine struct {
	Bmh            metal3.BareMetalHost
	ScheduleStatus ScheduledState
	// scheduleLabels
	// I expect to build this over time / if not might not be needed
	ScheduleLabels map[string]string
	VmRole         airshipv1.VmRoles
	// Data will contain whatever information is needed from the server
	// IF it ends up een just the IP then maybe we can collapse into a field
	Data *MachineData
}

func (m *Machine) String() string {
	return fmt.Sprintf("Machine {\n\tBmh:%s\n\tScheduleStatus:%s\n\tVmRole:%v\n}\n", m.Bmh.ObjectMeta.Name, m.ScheduleStatus, m.VmRole)
}

func NewMachine(bmh metal3.BareMetalHost, nodeRole airshipv1.VmRoles, schedState ScheduledState) (m *Machine) {
	return &Machine{
		Bmh:            bmh,
		ScheduleStatus: schedState,
		VmRole:         nodeRole,
		Data: &MachineData{
			IpOnInterface: make(map[string]string),
		},
	}
}

type MachineData struct {
	// Collect all IP's for the interfaces defined
	// In the list of Services
	IpOnInterface map[string]string
}

// MachineList contains the list of Scheduled or ToBeScheduled machines
type MachineList struct {
	// ViNO BMH
	Vbmhs map[string]*Machine
	// Keep track  of how many we have mark for scheduled.
	Scheduled map[airshipv1.VmRoles]int
}

func (ml *MachineList) hasMachine(bmh metal3.BareMetalHost) bool {

	if &bmh == nil {
		return false
	}
	fmt.Printf("Schedule.hasMachine  bmh.ObjectMeta.Name:%s ml.Vbmhs[bmh.ObjectMeta.Name] :%v , answer :%t \n", bmh.ObjectMeta.Name, ml.Vbmhs[bmh.ObjectMeta.Name], (ml.Vbmhs[bmh.ObjectMeta.Name] != nil))
	return ml.Vbmhs[bmh.ObjectMeta.Name] != nil
}

func (ml *MachineList) String() string {
	var sb strings.Builder

	for mName, machine := range ml.Vbmhs {

		sb.WriteString("[" + mName + "]:" + machine.String())
	}

	return sb.String()
}

func (ml *MachineList) Schedule(sip airshipv1.SIPCluster, c client.Client) error {

	// Initialize the Target list
	ml.init(sip.Spec.Nodes)

	// IDentify vBMH's that meet the appropriate selction criteria
	bmhList, err := ml.getVBMH(c)
	if err != nil {
		return err
	}

	//  Identify and Select the vBMH I actually will use
	err = ml.identifyNodes(sip, bmhList, c)
	if err != nil {
		return err
	}

	// If I get here the MachineList should have a selected set of  Machine's
	// They are in the ScheduleStatus of ToBeScheduled as well as the Role
	//
	fmt.Printf("Schedule ml.Vbmhs size:%d\n", len(ml.Vbmhs))
	return nil
}

func (ml *MachineList) init(nodes map[airshipv1.VmRoles]airshipv1.NodeSet) {
	// Only Initialize 1st time
	fmt.Printf("Schedule.init len(ml.Vbmhs):%d\n", len(ml.Vbmhs))
	if len(ml.Vbmhs) == 0 {
		mlSize := 0
		mlNodeTypes := 0
		for _, nodeCfg := range nodes {
			mlSize = mlSize + nodeCfg.Count.Active + nodeCfg.Count.Standby
			mlNodeTypes = mlNodeTypes + 1
		}
		//fmt.Printf("Schedule.init mlSize:%d\n", mlSize)
		ml.Scheduled = make(map[airshipv1.VmRoles]int, mlNodeTypes)
		ml.Vbmhs = make(map[string]*Machine, 0)
	}

}

func (ml *MachineList) getVBMH(c client.Client) (*metal3.BareMetalHostList, error) {

	bmhList := &metal3.BareMetalHostList{}

	// I am thinking we can add a Label for unsccheduled.
	// SIP Cluster can change it to scheduled.
	// We can then simple use this to select UNSCHEDULED
	/*
		This possible will not be needed if I figured out how to provide a != label.
		Then we can use DOESNT HAVE A TENANT LABEL
	*/
	scheduleLabels := map[string]string{SipScheduleLabel: "false"}

	err := c.List(context.Background(), bmhList, client.MatchingLabels(scheduleLabels))
	if err != nil {
		fmt.Printf("Schedule.getVBMH bmhList err:%v\n", err)
		return bmhList, err
	}
	fmt.Printf("Schedule.getVBMH bmhList size:%d\n", len(bmhList.Items))
	if len(bmhList.Items) > 0 {
		return bmhList, nil
	}
	return bmhList, fmt.Errorf("Unable to identify vBMH available for scheduling. Selecting  %v ", scheduleLabels)
}

func (ml *MachineList) identifyNodes(sip airshipv1.SIPCluster, bmhList *metal3.BareMetalHostList, c client.Client) error {
	// If using the SIP Sheduled label, we now have a list of vBMH;'s
	// that are not scheduled
	// Next I need to apply the constraints

	// This willl be a poor mans simple scheduler
	// Only deals with AntiAffinity at :
	// - Racks  : Dont select two machines in the same rack
	// - Server : Dont select two machines in the same server
	fmt.Printf("Schedule.identifyNodes bmList size:%d\n", len(bmhList.Items))
	for nodeRole, nodeCfg := range sip.Spec.Nodes {
		ml.Scheduled[nodeRole] = 0
		scheduleSetMap, err := ml.initScheduleMaps(nodeCfg.Scheduling)
		if err != nil {
			return err
		}
		err = ml.scheduleIt(nodeRole, nodeCfg, bmhList, scheduleSetMap, c, sip.Spec.Config)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Schedule.identifyNodes %s size:%d\n", ml.String(), len(ml.Vbmhs))
	return nil
}

func (ml *MachineList) initScheduleMaps(constraints []airshipv1.SchedulingOptions) (map[airshipv1.SchedulingOptions]*ScheduleSet, error) {
	setMap := make(map[airshipv1.SchedulingOptions]*ScheduleSet)

	for _, constraint := range constraints {
		if constraint == airshipv1.RackAntiAffinity {
			setMap[constraint] = &ScheduleSet{
				active:    true,
				set:       make(map[string]bool),
				labelName: RackLabel,
			}
		}
		if constraint == airshipv1.ServerAntiAffinity {
			setMap[constraint] = &ScheduleSet{
				active:    true,
				set:       make(map[string]bool),
				labelName: ServerLabel,
			}
		}
	}

	fmt.Printf("Schedule.initScheduleMaps  setMap:%v\n", setMap)
	if len(setMap) > 0 {
		return setMap, nil
	}
	return setMap, ErrorConstraintNotFound{}
}

func (ml *MachineList) countScheduledAndTobeScheduled(nodeRole airshipv1.VmRoles, c client.Client, sipCfg *airshipv1.SipConfig) int {

	bmhList := &metal3.BareMetalHostList{}
	scheduleLabels := map[string]string{
		SipScheduleLabel: "true",
		SipClusterLabel:  sipCfg.ClusterName,
		SipNodeTypeLabel: string(nodeRole),
	}

	fmt.Printf("Schedule.scheduleIt.countScheduledAndTobeScheduled  scheduleLabels:%v  \n", scheduleLabels)
	err := c.List(context.Background(), bmhList, client.MatchingLabels(scheduleLabels))
	if err != nil {
		fmt.Printf("Schedule.scheduleIt.countScheduledAndTobeScheduled  err:%s 0 \n", err)
		return 0
	}
	// TODO Update the Machine List
	// With what is already there.
	fmt.Printf("Schedule.scheduleIt.countScheduledAndTobeScheduled  FOUND vBMH's:%d ml.Scheduled[%v]:%d\nMachineList %s\n", len(bmhList.Items), nodeRole, ml.Scheduled[nodeRole], ml.String())
	for _, bmh := range bmhList.Items {
		if !ml.hasMachine(bmh) {
			// Add it to the list.
			ml.Vbmhs[bmh.ObjectMeta.Name] = NewMachine(bmh, nodeRole, Scheduled)
			ml.Scheduled[nodeRole] = ml.Scheduled[nodeRole] + 1
		}
	}
	// Scheduled should include:
	// - NEw added in previous iteratins tagged as ToBeScheduled
	// - Coun tfor those Added previously but now tagged as UnableToSchedule
	// - New vBMH Machines already tagged as as Scheduled
	fmt.Printf("Schedule.scheduleIt.countScheduledAndTobeScheduled AFTER PROCESS ml.Scheduled[%v]:%d\n MachineList %s\n", nodeRole, ml.Scheduled[nodeRole], ml.String())
	return ml.Scheduled[nodeRole]
}

func (ml *MachineList) scheduleIt(nodeRole airshipv1.VmRoles, nodeCfg airshipv1.NodeSet, bmList *metal3.BareMetalHostList,
	scheduleSetMap map[airshipv1.SchedulingOptions]*ScheduleSet, c client.Client, sipCfg *airshipv1.SipConfig) error {
	validBmh := true
	// Count the expectations stated in the CR
	// 	Reduce from the list of BMH's already scheduled and  labeled with the Cluster Name
	// 	Reduce from the number of Machines I have identified  already to be Labeled
	nodeTarget := (nodeCfg.Count.Active + nodeCfg.Count.Standby) - ml.countScheduledAndTobeScheduled(nodeRole, c, sipCfg)
	fmt.Printf("Schedule.scheduleIt  nodeRole:%v nodeTarget:%d nodeCfg.VmFlavor:%s  ml.Vbmhs len:%d \n", nodeRole, nodeTarget, nodeCfg.VmFlavor, len(ml.Vbmhs))
	// Nothing to schedule
	if nodeTarget == 0 {
		return nil
	}

	for _, bmh := range bmList.Items {
		fmt.Printf("---------------\n Schedule.scheduleIt  bmh.ObjectMeta.Name:%s \n", bmh.ObjectMeta.Name)

		if !ml.hasMachine(bmh) {
			for _, constraint := range nodeCfg.Scheduling {
				// Do I care about this constraint

				if scheduleSetMap[constraint].Active() {
					// Check if bmh has the label
					// There is a func (host *BareMetalHost) getLabel(name string) string {
					// Not sure why its not Public, so using our own method
					cLabelValue, cFlavorMatches := scheduleSetMap[constraint].GetLabels(bmh.Labels, nodeCfg.VmFlavor)

					// If it doesnt match the flavor its not valid
					validBmh = cFlavorMatches
					// If it does match the flavor
					if cLabelValue != "" && cFlavorMatches {
						// If its in the list already for the constraint , theen this bmh is disqualified. Skip it
						if scheduleSetMap[constraint].Exists(cLabelValue) {
							validBmh = false
							break
						} else {
							scheduleSetMap[constraint].Add(cLabelValue)
						}
					}
					//fmt.Printf("Schedule.scheduleIt cLabelValue:%s, cFlavorMatches:%t scheduleSetMap[%v]:%v\n", cLabelValue, cFlavorMatches, constraint, scheduleSetMap[constraint])

				}
			}
			fmt.Printf("Schedule.scheduleIt validBmh:%t, bmh.ObjectMeta.Name:%s  ml.Vbmhs len:%d\n", validBmh, bmh.ObjectMeta.Name, len(ml.Vbmhs))
			// All the constraints have been checked
			// Only if its not in the list already
			if validBmh {
				// Lets add it to the list as a schedulable thing
				ml.Vbmhs[bmh.ObjectMeta.Name] = NewMachine(bmh, nodeRole, ToBeScheduled)
				ml.Scheduled[nodeRole] = ml.Scheduled[nodeRole] + 1
				fmt.Printf("---------------\nSchedule.scheduleIt ADDED machine:%s \n", ml.Vbmhs[bmh.ObjectMeta.Name].String())
				// TODO Probable should remove the bmh from the list so if there are other node targets they dont even take it into account
				nodeTarget = nodeTarget - 1
				if nodeTarget == 0 {
					break
				}
			}
			// ...
			validBmh = true
		}

	}

	fmt.Printf("Schedule.scheduleIt nodeRole:%v nodeTarget:%d\n %s\n", nodeRole, nodeTarget, ml.String())
	if nodeTarget > 0 {
		return ErrorUnableToFullySchedule{
			TargetNode:   nodeRole,
			TargetFlavor: nodeCfg.VmFlavor,
		}
	}
	return nil
}

// Extrapolate
// The intention is to extract the IP information from the referenced networkData field for the BareMetalHost
func (ml *MachineList) Extrapolate(sip airshipv1.SIPCluster, c client.Client) bool {
	// Lets get the data for all selected BMH's.
	extrapolateSuccess := true
	fmt.Printf("Schedule.Extrapolate  ml.Vbmhs:%d\n", len(ml.Vbmhs))
	for _, machine := range ml.Vbmhs {
		fmt.Printf("Schedule.Extrapolate  machine.Data.IpOnInterface len:%d machine:%v \n", len(machine.Data.IpOnInterface), machine)

		// Skip if I alread extrapolated tehh data for this machine
		if len(machine.Data.IpOnInterface) > 0 {
			continue
		}
		bmh := machine.Bmh
		// Identify Network Data Secret name

		networkDataSecret := &corev1.Secret{}
		//fmt.Printf("Schedule.Extrapolate Namespace:%s  Name:%s\n", bmh.Spec.NetworkData.Namespace, bmh.Spec.NetworkData.Name)
		// c is a created client.
		err := c.Get(context.Background(), client.ObjectKey{
			Namespace: bmh.Spec.NetworkData.Namespace,
			Name:      bmh.Spec.NetworkData.Name,
		}, networkDataSecret)
		if err != nil {
			machine.ScheduleStatus = UnableToSchedule
			ml.Scheduled[machine.VmRole] = ml.Scheduled[machine.VmRole] - 1
			extrapolateSuccess = false
		}
		//fmt.Printf("Schedule.Extrapolate  networkDataSecret:%v\n", networkDataSecret)
		// Assuming there might be other data
		// Retrieve IP's for Service defined Network Interfaces
		err = ml.getIp(machine, networkDataSecret, sip.Spec.InfraServices)
		if err != nil {
			// Lets mark the machine as NotScheduleable.
			// Update the count of what I have found so far,
			machine.ScheduleStatus = UnableToSchedule
			ml.Scheduled[machine.VmRole] = ml.Scheduled[machine.VmRole] - 1
			extrapolateSuccess = false
		}

	}
	fmt.Printf("Schedule.Extrapolate  extrapolateSuccess:%t\n", extrapolateSuccess)
	return extrapolateSuccess
}

/***

{
    "links": [
        {
            "id": "eno4",
            "name": "eno4",
            "type": "phy",
            "mtu": 1500
        },
        {
            "id": "enp59s0f1",
            "name": "enp59s0f1",
            "type": "phy",
            "mtu": 9100
        },
        {
            "id": "enp216s0f0",
            "name": "enp216s0f0",
            "type": "phy",
            "mtu": 9100
        },
        {
            "id": "bond0",
            "name": "bond0",
            "type": "bond",
            "bond_links": [
                "enp59s0f1",
                "enp216s0f0"
            ],
            "bond_mode": "802.3ad",
            "bond_xmit_hash_policy": "layer3+4",
            "bond_miimon": 100,
            "mtu": 9100
        },
        {
            "id": "bond0.41",
            "name": "bond0.41",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 41,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.42",
            "name": "bond0.42",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 42,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.44",
            "name": "bond0.44",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 44,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.45",
            "name": "bond0.45",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 45,
            "mtu": 9100,
            "vlan_mac_address": null
        }
    ],
    "networks": [
        {
            "id": "oam-ipv6",
            "type": "ipv6",
            "link": "bond0.41",
            "ip_address": "2001:1890:1001:293d::139",
            "routes": [
                {
                    "network": "::/0",
                    "netmask": "::/0",
                    "gateway": "2001:1890:1001:293d::1"
                }
            ]
        },
        {
            "id": "oam-ipv4",
            "type": "ipv4",
            "link": "bond0.41",
            "ip_address": "32.68.51.139",
            "netmask": "255.255.255.128",
            "dns_nameservers": [
                "135.188.34.124",
                "135.38.244.16",
                "135.188.34.84"
            ],
            "routes": [
                {
                    "network": "0.0.0.0",
                    "netmask": "0.0.0.0",
                    "gateway": "32.68.51.129"
                }
            ]
        },
        {
            "id": "pxe-ipv6",
            "link": "eno4",
            "type": "ipv6",
            "ip_address": "fd00:900:100:138::11"
        },
        {
            "id": "pxe-ipv4",
            "link": "eno4",
            "type": "ipv4",
            "ip_address": "172.30.0.11",
            "netmask": "255.255.255.128"
        },
        {
            "id": "storage-ipv6",
            "link": "bond0.42",
            "type": "ipv6",
            "ip_address": "fd00:900:100:139::15"
        },
        {
            "id": "storage-ipv4",
            "link": "bond0.42",
            "type": "ipv4",
            "ip_address": "172.31.1.15",
            "netmask": "255.255.255.128"
        },
        {
            "id": "ksn-ipv6",
            "link": "bond0.44",
            "type": "ipv6",
            "ip_address": "fd00:900:100:13a::11"
        },
        {
            "id": "ksn-ipv4",
            "link": "bond0.44",
            "type": "ipv4",
            "ip_address": "172.29.0.11",
            "netmask": "255.255.255.128"
        }
    ]
}
***/

func (ml *MachineList) getIp(machine *Machine, networkDataSecret *corev1.Secret, infraServices map[airshipv1.InfraService]airshipv1.InfraConfig) error {
	var secretData interface{}
	// Now I have the Secret
	// Lets find the IP's for all Interfaces defined in Cfg
	foundIp := false
	for svcName, svcCfg := range infraServices {
		// Did I already find teh IP for these interface
		if machine.Data.IpOnInterface[svcCfg.NodeInterface] == "" {
			json.Unmarshal(networkDataSecret.Data["networkData"], &secretData)
			//fmt.Printf("Schedule.Extrapolate.getIp secretData:%v\n", secretData)

			queryFilter := fmt.Sprintf("$..networks[? (@.id==\"%s\")].ip_address", svcCfg.NodeInterface)
			//fmt.Printf("Schedule.Extrapolate.getIp queryFilter:%v\n", queryFilter)
			ip_address, err := jsonpath.Get(queryFilter, secretData)

			if err == nil {
				foundIp = true
				for _, value := range ip_address.([]interface{}) {
					machine.Data.IpOnInterface[svcCfg.NodeInterface] = value.(string)
				}

			}
			// Skip if error
			// Should signal that I need to exclude this machine
			// Which also means I am now short potentially.
			fmt.Printf("Schedule.Extrapolate.getIp machine.Data.IpOnInterface[%s]:%v\n", svcCfg.NodeInterface, machine.Data.IpOnInterface[svcCfg.NodeInterface])
		}

		if !foundIp {
			return &ErrorHostIpNotFound{
				HostName:    machine.Bmh.ObjectMeta.Name,
				ServiceName: svcName,
				IPInterface: svcCfg.NodeInterface,
			}
		}
	}
	return nil
}

/*
  ScheduleSet is a simple object to encapsulate data that
  helps our poor man scheduler
*/
type ScheduleSet struct {
	// Defines if this set is actually active
	active bool
	// Holds list of elements in teh Set
	set map[string]bool
	// Holds the label name that identifies the constraint
	labelName string
}

func (ss *ScheduleSet) Active() bool {
	return ss.active
}
func (ss *ScheduleSet) Exists(value string) bool {
	if len(ss.set) > 0 {
		return ss.set[value]
	}
	return false
}

func (ss *ScheduleSet) Add(labelValue string) {
	ss.set[labelValue] = true
}
func (ss *ScheduleSet) GetLabels(labels map[string]string, flavorLabel string) (string, bool) {
	//fmt.Printf("Schedule.scheduleIt.GetLabels  labels:%v, flavorLabel:%s\n", labels, flavorLabel)
	if labels == nil {
		return "", false
	}
	cl := strings.Split(flavorLabel, "=")
	if len(cl) > 0 {
		return labels[ss.labelName], labels[cl[0]] == cl[1]
	}
	return labels[ss.labelName], false
}

/*
ApplyLabel : marks the appropriate machine labels to teh vBMH's that
have benn selected by the scheduling.
This is done only after the Infrastcuture Services have been  deployed
*/
func (ml *MachineList) ApplyLabels(sip airshipv1.SIPCluster, c client.Client) error {

	fmt.Printf("ApplyLabels  %s size:%d\n", ml.String(), len(ml.Vbmhs))
	for _, machine := range ml.Vbmhs {
		// Only Add LAbels to Machines that are not amrked to be scheduled
		if machine.ScheduleStatus == ToBeScheduled {
			bmh := &machine.Bmh
			fmt.Printf("ApplyLabels bmh.ObjectMeta.Name:%s\n", bmh.ObjectMeta.Name)
			bmh.Labels[SipClusterLabel] = sip.Spec.Config.ClusterName
			bmh.Labels[SipScheduleLabel] = "true"
			bmh.Labels[SipNodeTypeLabel] = string(machine.VmRole)

			// This is bombing when it find 1 error
			// Might be better to acculumalte the errors, and
			// Allow it  to continue.
			err := c.Update(context.Background(), bmh)
			if err != nil {
				fmt.Printf("ApplyLabel bmh:%s err:%v\n", bmh.ObjectMeta.Name, err)
				return err
			}
		}
	}
	return nil
}

/*
RemoveLabels
*/
func (ml *MachineList) RemoveLabels(sip airshipv1.SIPCluster, c client.Client) error {

	fmt.Printf("ApplyLabels  %s size:%d\n", ml.String(), len(ml.Vbmhs))
	for _, machine := range ml.Vbmhs {

		bmh := &machine.Bmh
		fmt.Printf("RemoveLabels bmh.ObjectMeta.Name:%s\n", bmh.ObjectMeta.Name)
		bmh.Labels[SipClusterLabel] = "" // REMOVE IT TODO This only blanks it out doesnt remove the label
		bmh.Labels[SipScheduleLabel] = "false"
		bmh.Labels[SipNodeTypeLabel] = "" // REMOVE IT

		// This is bombing when it find 1 error
		// Might be better to acculumalte the errors, and
		// Allow it  to continue.
		err := c.Update(context.Background(), bmh)
		if err != nil {
			fmt.Printf("RemoveLabels bmh:%s err:%v\n", bmh.ObjectMeta.Name, err)
			return err
		}
	}
	return nil
}

func (ml *MachineList) GetCluster(sip airshipv1.SIPCluster, c client.Client) error {

	// Initialize teh Target list
	ml.init(sip.Spec.Nodes)

	bmhList := &metal3.BareMetalHostList{}
	scheduleLabels := map[string]string{
		SipScheduleLabel: "true",
		SipClusterLabel:  sip.Spec.Config.ClusterName,
	}

	err := c.List(context.Background(), bmhList, client.MatchingLabels(scheduleLabels))
	if err != nil {
		return err
	}

	for _, bmh := range bmhList.Items {
		ml.Vbmhs[bmh.ObjectMeta.Name] = &Machine{
			Bmh:            bmh,
			ScheduleStatus: Scheduled,
			VmRole:         airshipv1.VmRoles(bmh.Labels[SipNodeTypeLabel]),
			Data: &MachineData{
				IpOnInterface: make(map[string]string),
			},
		}
	}

	fmt.Printf("GetCluster %s \n", ml.String())
	return nil
}
