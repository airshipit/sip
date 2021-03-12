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

package bmh

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	airshipv1 "sipcluster/pkg/api/v1"

	"github.com/PaesslerAG/jsonpath"
	"github.com/go-logr/logr"
	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	kerror "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScheduledState string

// Possible Node or BMH roles  for a Tenant
const (
	// ToBeScheduled means that the BMH was identified by the scheduler to be selected
	ToBeScheduled ScheduledState = "Selected"

	// Scheduled means the BMH  already has a label implying it
	// was previously scheduled
	Scheduled ScheduledState = "Scheduled"

	// NotScheduled, This BMH was not selected to be scheduled.
	// Either because we didnt meet the criteria
	// or simple we reached the count
	NotScheduled ScheduledState = "NotSelected"

	// UnableToSchedule, This BMH has something wrong with it
	// The BMH itself doesnt depict the error situation
	// i.e. the NetworkData is missing something
	UnableToSchedule ScheduledState = "UnableToSchedule"
)

const (
	BaseAirshipSelector = "sip.airshipit.org"

	// This label is applied to all BMHs scheduled to a given SIPCluster.
	SipClusterLabelName = "cluster"
	SipClusterLabel     = BaseAirshipSelector + "/" + SipClusterLabelName

	SipNodeTypeLabelName = "node-type"
	SipNodeTypeLabel     = BaseAirshipSelector + "/" + SipNodeTypeLabelName
)

// Keys used to retrieve credentials from the BMC credentials secret
const (
	keyBMCUsername = "username"
	keyBMCPassword = "password"
)

// MAchine represents an individual BMH CR, and the appropriate
// attributes required to manage the SIP Cluster scheduling and
// rocesing needs about thhem
type Machine struct {
	BMH            metal3.BareMetalHost
	ScheduleStatus ScheduledState
	// scheduleLabels
	// I expect to build this over time / if not might not be needed
	ScheduleLabels map[string]string
	BMHRole        airshipv1.BMHRole
	// Data will contain whatever information is needed from the server
	// IF it ends up een just the IP then maybe we can collapse into a field
	Data *MachineData
}

func (m *Machine) String() string {
	// TODO(howell): cleanup this manual marshaling
	return fmt.Sprintf("Machine {\n\tBmh:%s\n\tScheduleStatus:%s\n\tBMHRole:%v\n}\n",
		m.BMH.ObjectMeta.Name, m.ScheduleStatus, m.BMHRole)
}

func NewMachine(bmh metal3.BareMetalHost, nodeRole airshipv1.BMHRole, schedState ScheduledState) (m *Machine, e error) {
	// Add logic to check if required fields exist.
	if bmh.Spec.NetworkData == nil {
		return nil, &ErrorNetworkDataNotFound{BMH: bmh}
	}
	return &Machine{
		BMH:            bmh,
		ScheduleStatus: schedState,
		BMHRole:        nodeRole,
		Data: &MachineData{
			IPOnInterface: make(map[string]string),
		},
	}, nil
}

type MachineData struct {
	// Collect all IP's for the interfaces defined
	// In the list of Services
	IPOnInterface map[string]string
	BMCUsername   string
	BMCPassword   string
}

// MachineList contains the list of Scheduled or ToBeScheduled machines
type MachineList struct {
	NamespacedName types.NamespacedName
	// Machines
	Machines map[string]*Machine
	// Keep track  of how many we have mark for scheduled.
	ReadyForScheduleCount map[airshipv1.BMHRole]int
	Log                   logr.Logger
}

func (ml *MachineList) hasMachine(bmh metal3.BareMetalHost) bool {
	return ml.Machines[bmh.ObjectMeta.Name] != nil
}

func (ml *MachineList) String() string {
	// TODO(howell): This output probably isn't formatted properly
	var sb strings.Builder
	for mName, machine := range ml.Machines {
		sb.WriteString("[" + mName + "]:" + machine.String())
	}
	return sb.String()
}

func (ml *MachineList) Schedule(sip airshipv1.SIPCluster, c client.Client) error {
	ml.Log.Info("starting scheduling of BaremetalHosts")

	// Initialize the Target list
	ml.init(sip.Spec.Nodes)

	// IDentify BMH's that meet the appropriate selction criteria
	bmhList, err := ml.getBMHs(c)
	if err != nil {
		return err
	}

	//  Identify and Select the BMH I actually will use
	err = ml.identifyNodes(sip, bmhList, c)
	if err != nil {
		return err
	}

	// If I get here the MachineList should have a selected set of  Machine's
	// They are in the ScheduleStatus of ToBeScheduled as well as the Role
	//
	ml.Log.Info("Found machines for scheduling", "count", len(ml.Machines))
	return nil
}

func (ml *MachineList) init(nodes map[airshipv1.BMHRole]airshipv1.NodeSet) {
	// Only Initialize 1st time
	if len(ml.Machines) == 0 {
		mlSize := 0
		mlNodeTypes := 0
		for _, nodeCfg := range nodes {
			mlSize = mlSize + nodeCfg.Count.Active + nodeCfg.Count.Standby
			mlNodeTypes++
		}
		fmt.Printf("Schedule.init mlSize:%d\n", mlSize)
		ml.ReadyForScheduleCount = make(map[airshipv1.BMHRole]int, mlNodeTypes)
		ml.Machines = make(map[string]*Machine, 0)
	}
}

func (ml *MachineList) getBMHs(c client.Client) (*metal3.BareMetalHostList, error) {
	bmhList := &metal3.BareMetalHostList{}

	// Select BMH not yet labeled as scheduled by SIP
	unscheduledSelector := labels.NewSelector()
	r, err := labels.NewRequirement(SipClusterLabel, selection.DoesNotExist, nil)
	if err == nil {
		unscheduledSelector = unscheduledSelector.Add(*r)
	}

	ml.Log.Info("Getting all available BaremetalHosts that are not scheduled")
	err = c.List(context.Background(), bmhList, client.MatchingLabelsSelector{Selector: unscheduledSelector})
	if err != nil {
		ml.Log.Info("Received an error while getting BaremetalHost list", "error", err.Error())
		return bmhList, err
	}
	ml.Log.Info("Got a list of hosts", "BaremetalHostCount", len(bmhList.Items))
	if len(bmhList.Items) > 0 {
		return bmhList, nil
	}
	return bmhList, fmt.Errorf("Unable to identify BMH available for scheduling. Selecting  %v ", unscheduledSelector)
}

func (ml *MachineList) identifyNodes(sip airshipv1.SIPCluster,
	bmhList *metal3.BareMetalHostList, c client.Client) error {
	// If using the SIP Sheduled label, we now have a list of BMH;'s
	// that are not scheduled
	// Next I need to apply the constraints
	ml.Log.Info("Trying to identify BaremetalHosts that match scheduling parameters",
		"initial BMH count", len(bmhList.Items))
	for nodeRole, nodeCfg := range sip.Spec.Nodes {
		logger := ml.Log.WithValues("role", nodeRole) //nolint:govet
		ml.ReadyForScheduleCount[nodeRole] = 0
		logger.Info("Getting host constraints")
		scheduleSetMap := ml.initScheduleMaps(nodeRole, nodeCfg.TopologyKey)
		logger.Info("Matching hosts against constraints")
		err := ml.scheduleIt(nodeRole, nodeCfg, bmhList, scheduleSetMap, c, GetClusterLabel(sip))
		if err != nil {
			return err
		}
	}
	return nil
}

func (ml *MachineList) initScheduleMaps(role airshipv1.BMHRole,
	topologyKey string) *ScheduleSet {
	logger := ml.Log.WithValues("role", role, "topologyKey", topologyKey)

	logger.Info("Marking schedule set as active")
	return &ScheduleSet{
		active:      true,
		set:         make(map[string]bool),
		topologyKey: topologyKey,
	}
}

func (ml *MachineList) countScheduledAndTobeScheduled(nodeRole airshipv1.BMHRole,
	c client.Client, clusterName string) int {
	bmhList := &metal3.BareMetalHostList{}

	scheduleLabels := map[string]string{
		SipClusterLabel:  clusterName,
		SipNodeTypeLabel: string(nodeRole),
	}

	logger := ml.Log.WithValues("role", nodeRole)
	logger.Info("Getting list of BaremetalHost already scheduled for SIP cluster from kubernetes")
	err := c.List(context.Background(), bmhList, client.MatchingLabels(scheduleLabels))
	if err != nil {
		logger.Info("Received error when getting BaremetalHosts", "error", err.Error())
		return 0
	}

	// TODO Update the Machine List
	// With what is already there.
	logger.Info("Got already scheduled BaremetalHosts from kubernetes", "count", len(bmhList.Items))
	for _, bmh := range bmhList.Items {
		logger := logger.WithValues("BMH name", bmh.GetName()) //nolint:govet
		readyScheduled := !ml.hasMachine(bmh)
		logger.Info("Checking if BMH is already marked to be scheduled", "ready to be scheduled", readyScheduled)
		if readyScheduled {
			logger.Info("BMH host is not yet marked as ready to be scheduled, marking it as ready to be scheduled")
			// Add it to the list.
			m, err := NewMachine(bmh, nodeRole, Scheduled)
			if err != nil {
				logger.Info("BMH did not meet scheduling requirements", "error", err.Error())
				continue
			}
			ml.Machines[bmh.ObjectMeta.Name] = m
			ml.ReadyForScheduleCount[nodeRole]++
		}
	}
	// ReadyForScheduleCount should include:
	// - New added in previous iteratins tagged as ToBeScheduled
	// - Count for those Added previously but now tagged as UnableToSchedule
	// - New BMH Machines already tagged as as Scheduled
	return ml.ReadyForScheduleCount[nodeRole]
}

func (ml *MachineList) scheduleIt(nodeRole airshipv1.BMHRole, nodeCfg airshipv1.NodeSet,
	bmList *metal3.BareMetalHostList, scheduleSet *ScheduleSet,
	c client.Client, clusterName string) error {
	logger := ml.Log.WithValues("role", nodeRole)
	validBmh := true
	// Count the expectations stated in the CR
	// 	Reduce from the list of BMH's already scheduled and  labeled with the Cluster Name
	// 	Reduce from the number of Machines I have identified  already to be Labeled
	totalNodes := nodeCfg.Count.Active + nodeCfg.Count.Standby
	nodeTarget := totalNodes - ml.countScheduledAndTobeScheduled(nodeRole, c, clusterName)

	logger.Info("BMH count that need to be scheduled for SIP cluster discouting nodes ready to be scheduled",
		"BMH count to be scheduled", nodeTarget)
	// Nothing to schedule
	if nodeTarget == 0 {
		return nil
	}
	logger.Info("Checking list of BMH initially received as not scheduled anywhere yet")
	for _, bmh := range bmList.Items {
		logger := logger.WithValues("BaremetalHost Name", bmh.GetName()) //nolint:govet

		if !ml.hasMachine(bmh) {
			logger.Info("BaremetalHost not yet marked as ready to be scheduled")
			topologyKey := nodeCfg.TopologyKey
			// Do I care about this constraint
			logger := logger.WithValues("topologyKey", topologyKey) //nolint:govet
			if scheduleSet.Active() {
				logger.Info("constraint is active")
				// Check if bmh has the label
				topologyDomain, match, err := scheduleSet.GetLabels(labels.Set(bmh.Labels), &nodeCfg.LabelSelector)
				if err != nil {
					return err
				}
				logger.Info("Checked BMH topology key and label selector",
					"topology domain", topologyDomain,
					"label selector match", match)
				validBmh = match
				// If it does match the label selector
				if topologyDomain != "" && match {
					// If its in the list already for the constraint , theen this bmh is disqualified. Skip it
					if scheduleSet.Exists(topologyDomain) {
						logger.Info("Topology domain has already been scheduled to, skipping it")
						continue
					} else {
						scheduleSet.Add(topologyDomain)
					}
				}
			}
			// All the constraints have been checked
			// Only if its not in the list already
			if validBmh {
				// Lets add it to the list as a schedulable thing
				m, err := NewMachine(bmh, nodeRole, ToBeScheduled)
				if err != nil {
					logger.Info("Skipping BMH host as it did not meet creation requirements", "error", err.Error())
					continue
				}
				ml.Machines[bmh.ObjectMeta.Name] = m
				ml.ReadyForScheduleCount[nodeRole]++
				// TODO Probable should remove the bmh from the
				// list so if there are other node targets they
				// dont even take it into account
				nodeTarget--
				logger.Info("Marked node as ready to be scheduled", "BMH count to be scheduled", nodeTarget)
				if nodeTarget == 0 {
					break
				}
			} else {
				logger.Info("BMH didn't pass scheduling test", "BMH count to be scheduled", nodeTarget)
			}
			// ...
			validBmh = true
		}
	}

	if nodeTarget > 0 {
		logger.Info("Failed to get enough BMHs to complete scheduling")
		return ErrorUnableToFullySchedule{
			TargetNode:          nodeRole,
			TargetLabelSelector: nodeCfg.LabelSelector,
		}
	}
	return nil
}

// ExtrapolateServiceAddresses extracts the IP addresses of each network interface mapped to a service in the SIPCluster
// CR by inspecting each BMH's Network Data Secret.
func (ml *MachineList) ExtrapolateServiceAddresses(sip airshipv1.SIPCluster, c client.Client) error {
	// NOTE: At this point in the scheduling algorithm, the list of Machines in the MachineList each have BMH
	// objects that meet the SIPCluster CR topology and role constraints.

	var extrapolateErrs error
	for _, machine := range ml.Machines {
		// Skip machines whose service addresses have been extracted
		if len(machine.Data.IPOnInterface) > 0 {
			continue
		}

		// Retrieve BMH Network Data Secret
		networkDataSecret := &corev1.Secret{}
		err := c.Get(context.Background(), client.ObjectKey{
			Namespace: machine.BMH.Spec.NetworkData.Namespace,
			Name:      machine.BMH.Spec.NetworkData.Name,
		}, networkDataSecret)
		if err != nil {
			ml.Log.Error(err, "unable to retrieve BMH Network Data Secret", "BMH", machine.BMH.Name,
				"Secret", machine.BMH.Spec.NetworkData.Name,
				"Secret Namespace", machine.BMH.Spec.NetworkData.Namespace)

			machine.ScheduleStatus = UnableToSchedule
			ml.ReadyForScheduleCount[machine.BMHRole]--
			extrapolateErrs = kerror.NewAggregate([]error{extrapolateErrs, err})

			continue
		}

		// Parse the interface IP addresses from the BMH's Network Data
		err = ml.getIP(machine, networkDataSecret, sip.Spec.Services)
		if err != nil {
			ml.Log.Error(err, "unable to parse BMH Network Data Secret", "BMH", machine.BMH.Name,
				"Secret", machine.BMH.Spec.NetworkData.Name,
				"Secret Namespace", machine.BMH.Spec.NetworkData.Namespace)

			machine.ScheduleStatus = UnableToSchedule
			ml.ReadyForScheduleCount[machine.BMHRole]--
			extrapolateErrs = kerror.NewAggregate([]error{extrapolateErrs, err})
		}
	}

	return extrapolateErrs
}

// ExtrapolateBMCAuth extracts the BMC authentication information in each BMH's BMC Credentials Secret.
func (ml *MachineList) ExtrapolateBMCAuth(sip airshipv1.SIPCluster, c client.Client) error {
	// NOTE: At this point in the scheduling algorithm, the list of Machines in the MachineList each have BMH
	// objects that meet the SIPCluster CR topology and role constraints.

	var extrapolateErrs error
	for _, machine := range ml.Machines {
		// Retrieve BMC credentials Secret
		bmcCredsSecret := &corev1.Secret{}
		err := c.Get(context.Background(), client.ObjectKey{
			Namespace: machine.BMH.Namespace,
			Name:      machine.BMH.Spec.BMC.CredentialsName,
		}, bmcCredsSecret)
		if err != nil {
			ml.Log.Error(err, "unable to retrieve BMH BMC credentials Secret", "BMH", machine.BMH.Name,
				"Secret", machine.BMH.Spec.BMC.CredentialsName,
				"Secret Namespace", machine.BMH.Namespace)

			machine.ScheduleStatus = UnableToSchedule
			ml.ReadyForScheduleCount[machine.BMHRole]--
			extrapolateErrs = kerror.NewAggregate([]error{extrapolateErrs, err})

			continue
		}

		// Parse BMC credentials from Secret
		err = ml.getMangementCredentials(machine, bmcCredsSecret)
		if err != nil {
			ml.Log.Error(err, "unable to parse BMH BMC credentials Secret", "BMH", machine.BMH.Name,
				"Secret", machine.BMH.Spec.BMC.CredentialsName,
				"Secret Namespace", machine.BMH.Namespace)

			machine.ScheduleStatus = UnableToSchedule
			ml.ReadyForScheduleCount[machine.BMHRole]--
			extrapolateErrs = kerror.NewAggregate([]error{extrapolateErrs, err})
		}
	}

	return extrapolateErrs
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

func (ml *MachineList) getIP(machine *Machine, networkDataSecret *corev1.Secret,
	services airshipv1.SIPClusterServices) error {
	var secretData interface{}
	// Now I have the Secret
	// Lets find the IP's for all Interfaces defined in Cfg
	foundIP := false
	for _, svcCfg := range services.GetAll() {
		// Did I already find the IP for these interface
		if machine.Data.IPOnInterface[svcCfg.NodeInterface] == "" {
			err := json.Unmarshal(networkDataSecret.Data["networkData"], &secretData)
			if err != nil {
				return err
			}
			fmt.Printf("Schedule.Extrapolate.getIP secretData:%v\n", secretData)

			queryFilter := fmt.Sprintf("$..networks[? (@.id==\"%s\")].ip_address", svcCfg.NodeInterface)
			fmt.Printf("Schedule.Extrapolate.getIP queryFilter:%v\n", queryFilter)
			ipAddress, err := jsonpath.Get(queryFilter, secretData)

			if err == nil {
				foundIP = true
				for _, value := range ipAddress.([]interface{}) {
					machine.Data.IPOnInterface[svcCfg.NodeInterface] = value.(string) //nolint:errcheck
				}
			}
			// Skip if error
			// Should signal that I need to exclude this machine
			// Which also means I am now short potentially.
			fmt.Printf("Schedule.Extrapolate.getIP machine.Data.IpOnInterface[%s]:%v\n",
				svcCfg.NodeInterface, machine.Data.IPOnInterface[svcCfg.NodeInterface])
		}

		if !foundIP {
			return &ErrorHostIPNotFound{
				HostName:    machine.BMH.ObjectMeta.Name,
				IPInterface: svcCfg.NodeInterface,
			}
		}
	}
	return nil
}

// getManagementCredentials retrieves BMC credentials from a Kubernetes secret.
func (ml *MachineList) getMangementCredentials(machine *Machine, secret *corev1.Secret) error {
	username, exists := secret.Data[keyBMCUsername]
	if !exists {
		return ErrMalformedManagementCredentials{SecretName: secret.Name}
	}
	machine.Data.BMCUsername = string(username)

	password, exists := secret.Data[keyBMCPassword]
	if !exists {
		return ErrMalformedManagementCredentials{SecretName: secret.Name}
	}
	machine.Data.BMCPassword = string(password)

	return nil
}

// ScheduleSet is a simple object to encapsulate data that helps our poor man scheduler
type ScheduleSet struct {
	// Defines if this set is actually active
	active bool
	// Holds list of elements in the Set
	set map[string]bool
	// Holds the topology key that identifies the constraint
	topologyKey string
}

func (ss *ScheduleSet) Active() bool {
	return ss.active
}

func (ss *ScheduleSet) Exists(value string) bool {
	return ss.set[value]
}

func (ss *ScheduleSet) Add(labelValue string) {
	ss.set[labelValue] = true
}
func (ss *ScheduleSet) GetLabels(labels labels.Labels, labelSelector *metav1.LabelSelector) (string, bool, error) {
	fmt.Printf("Schedule.scheduleIt.GetLabels labels:%v, labelSelector:%s\n", labels, labelSelector)

	match := false
	if labels == nil {
		return "", match, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err == nil {
		match = selector.Matches(labels)
	}
	return labels.Get(ss.topologyKey), match, err
}

// ApplyLabels adds the appropriate labels to the BMHs that are ready to be scheduled
func (ml *MachineList) ApplyLabels(sip airshipv1.SIPCluster, c client.Client) error {
	fmt.Printf("ApplyLabels %s size:%d\n", ml.String(), len(ml.Machines))
	for _, machine := range ml.Machines {
		if machine.ScheduleStatus == ToBeScheduled {
			bmh := &machine.BMH
			fmt.Printf("ApplyLabels bmh.ObjectMeta.Name:%s\n", bmh.ObjectMeta.Name)
			bmh.Labels[SipClusterLabel] = GetClusterLabel(sip)
			bmh.Labels[SipNodeTypeLabel] = string(machine.BMHRole)

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

// RemoveLabels removes sip related labels
func (ml *MachineList) RemoveLabels(c client.Client) error {
	fmt.Printf("RemoveLabels %s size:%d\n", ml.String(), len(ml.Machines))
	for _, machine := range ml.Machines {
		bmh := &machine.BMH
		fmt.Printf("RemoveLabels bmh.ObjectMeta.Name:%s\n", bmh.ObjectMeta.Name)
		delete(bmh.Labels, SipClusterLabel)
		delete(bmh.Labels, SipNodeTypeLabel)

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
	// Initialize the Target list
	ml.init(sip.Spec.Nodes)

	bmhList := &metal3.BareMetalHostList{}
	scheduleLabels := map[string]string{
		SipClusterLabel: GetClusterLabel(sip),
	}

	err := c.List(context.Background(), bmhList, client.MatchingLabels(scheduleLabels))
	if err != nil {
		return err
	}

	for _, bmh := range bmhList.Items {
		ml.Machines[bmh.ObjectMeta.Name] = &Machine{
			BMH:            bmh,
			ScheduleStatus: Scheduled,
			BMHRole:        airshipv1.BMHRole(bmh.Labels[SipNodeTypeLabel]),
			Data: &MachineData{
				IPOnInterface: make(map[string]string),
			},
		}
	}

	fmt.Printf("GetCluster %s \n", ml.String())
	return nil
}

func GetClusterLabel(sip airshipv1.SIPCluster) string {
	return fmt.Sprintf("%s_%s", sip.GetNamespace(), sip.GetName())
}
