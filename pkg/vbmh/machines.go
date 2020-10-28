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
	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	airshipv1 "sipcluster/pkg/api/v1"
	//corev1 "k8s.io/api/core/v1"
	//rbacv1 "k8s.io/api/rbac/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
)

const (
	BaseAirshipSelector = "airshipit.org"
	SipScheduled        = BaseAirshipSelector + "/sip-scheduled in (True, true)"
	SipNotScheduled     = BaseAirshipSelector + "/sip-scheduled in (False, false)"
)

// MAchine represents an individual BMH CR, and teh appropriate
// attributes required to manage the SIP Cluster scheduling and
// rocesing needs about thhem
type Machine struct {
	Bmh           metal3.BareMetalHost
	ScheduleState ScheduledState
	// scheduleLabels
	ScheduleLabels map[string]string
	Data           *MachineData
}

type MachineData struct {
	// Some Data
}
type MachineList struct {
	bmhs []Machine

	// I might have some index here of MAchines that have been actualy Selected to be Scheduled.

}

func (ml *MachineList) Schedule(nodes map[airshipv1.VmRoles]airshipv1.NodeSet, c client.Client) error {
	// Initialize teh Target list
	ml.bmhs = ml.init(nodes)

	// IDentify vBMH's that meet the appropriate selction criteria
	bmList, err := ml.getVBMH(c)
	if err != nil {
		return err
	}

	//  Identify and Select the vBMH I actually will use
	err = ml.identifyNodes(nodes, bmList)
	if err != nil {
		return err
	}

	// If I get here the MachineList should have a set of properly labeled Machine's
	// We will also Flag these machines adequately so that we can extrapolate the data
	// However we wil not be labelling yet.

	return nil
}

func (ml *MachineList) init(nodes map[airshipv1.VmRoles]airshipv1.NodeSet) []Machine {
	mlSize := 0
	for _, nodeCfg := range nodes {
		mlSize = mlSize + nodeCfg.Count.Active + nodeCfg.Count.Standby
	}
	return make([]Machine, mlSize)

}

func (ml *MachineList) getVBMH(c client.Client) (*metal3.BareMetalHostList, error) {
	bmList := &metal3.BareMetalHostList{}
	// I am thinking we can add a Label for unsccheduled.
	// SIP Cluster can change it to scheduled.
	// We can then simple use this to select UNSCHEDULED
	/*
		This possible will not be needed if I figured out how to provide a != label.
		Then we can use DOESNT HAVE A TENANT LABEL
	*/
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{SipNotScheduled: "False"}}
	bmhSelector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return bmList, err
	}
	bmListOptions := &client.ListOptions{
		LabelSelector: bmhSelector,
		Limit:         100,
	}

	err = c.List(context.TODO(), bmList, bmListOptions)
	if err != nil {
		return bmList, err
	}
	return bmList, nil
}

func (ml *MachineList) identifyNodes(nodes map[airshipv1.VmRoles]airshipv1.NodeSet, bmList *metal3.BareMetalHostList) error {
	// If using the SIP Sheduled label, we now have a list of vBMH;'s
	// that are not scheduled
	// Next I need to apply the constraints
	for nodeRole, nodeCfg := range nodes {

		for nodeIdx := 0; nodeIdx < (nodeCfg.Count.Active + nodeCfg.Count.Standby); nodeIdx++ {
			err := ml.scheduler(nodeRole, nodeCfg.Scheduling, bmList)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ml *MachineList) scheduler(nodeRole airshipv1.VmRoles, constraints []string, bmList *metal3.BareMetalHostList) error {
	// nodeRole will drive labels
	return nil
}
