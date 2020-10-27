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
}

type MachineList struct {
	bmhs []Machine

	// I might have some index here of MAchines that have been actualy Selected to be Scheduled.

}

func (ml *MachineList) Schedule(nodes *airshipv1.NodeSet, c client.Client) error {
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
		return err
	}
	bmListOptions := &client.ListOptions{
		LabelSelector: bmhSelector,
		Limit:         100,
	}

	err = c.List(context.TODO(), bmList, bmListOptions)
	if err != nil {
		return err
	}

	// If using the SIP Sheduled label, we now have a list of vBMH;'s
	// that are not scheduled
	// Next I need to apply the constraints

	return nil
}
