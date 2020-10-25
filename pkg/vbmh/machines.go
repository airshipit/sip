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

package vms

import (
	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
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

// MAchine represents an individual BMH CR, and teh appropriate
// attributes required to manage the SIP Cluster scheduling and
// rocesing needs about thhem
type Machine struct {
	Bmh           metal3.BareMetalHostSpec
	AcheduleState ScheduledState
	// scheduleLabels
	ScheduleLabels map[string]string
}

type MachineList struct {
	bmhs []Machine

	// I might have some index here of MAchines that have been actualy Selected to be Scheduled.

}
