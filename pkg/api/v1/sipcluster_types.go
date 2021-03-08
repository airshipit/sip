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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true

// SIPClusterList contains a list of SIPCluster
type SIPClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SIPCluster `json:"items"`
}

// +kubebuilder:object:root=true

// SIPCluster is the Schema for the sipclusters API
// +kubebuilder:subresource:status
type SIPCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SIPClusterSpec   `json:"spec,omitempty"`
	Status SIPClusterStatus `json:"status,omitempty"`
}

// SIPClusterSpec defines the desired state of a SIPCluster
type SIPClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make manifests to regenerate code after modifying this file

	// Nodes defines the set of nodes to schedule for each BMH role.
	Nodes map[BMHRole]NodeSet `json:"nodes,omitempty"`

	// Services defines the services that are deployed when a SIPCluster is provisioned.
	Services SIPClusterServices `json:"services"`
}

// SIPClusterServices defines the services that are deployed when a SIPCluster is provisioned.
type SIPClusterServices struct {
	// LoadBalancer defines the sub-cluster load balancer services.
	LoadBalancer []SIPClusterService `json:"loadBalancer,omitempty"`
	// Auth defines the sub-cluster authentication services.
	Auth []SIPClusterService `json:"auth,omitempty"`
	// JumpHost defines the sub-cluster jump host services.
	JumpHost []JumpHostService `json:"jumpHost,omitempty"`
}

func (s SIPClusterServices) GetAll() []SIPClusterService {
	all := []SIPClusterService{}
	for _, s := range s.LoadBalancer {
		all = append(all, s)
	}
	for _, s := range s.Auth {
		all = append(all, s)
	}
	for _, s := range s.JumpHost {
		all = append(all, s.SIPClusterService)
	}
	return all
}

// JumpHostService is an infrastructure service type that represents the sub-cluster jump-host service.
type JumpHostService struct {
	SIPClusterService `json:",inline"`
	BMC               *BMCOpts `json:"bmc,omitempty"`
	SSHAuthorizedKeys []string `json:"sshAuthorizedKeys,omitempty"`
	// NodeSSHPrivateKeys holds the name of a Secret in the same namespace as the SIPCluster CR,
	// whose key values each represent an ssh private key that can be used to access the cluster nodes.
	// They are mounted into the jumphost with the secret keys serving as file names relative to a common
	// directory, and then configured as identity files in the SSH config file of the default user.
	NodeSSHPrivateKeys string `json:"nodeSSHPrivateKeys"`
}

// SIPClusterStatus defines the observed state of SIPCluster
type SIPClusterStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	// ConditionTypeReady indicates whether a resource is available for utilization
	ConditionTypeReady string = "Ready"

	// ReasonTypeInfraServiceFailure indicates that a resource has a specified condition because SIP was unable
	// to configure infrastructure services for the SIPCluster.
	ReasonTypeInfraServiceFailure string = "InfraServiceFailure"

	// ReasonTypeProgressing indicates that a resource has a specified condition because SIP is processing it.
	ReasonTypeProgressing string = "Progressing"

	// ReasonTypeUnableToApplyLabels indicates that a resource has a specified condition because SIP was unable to
	// apply labels to BMHs for the SIPCluster.
	ReasonTypeUnableToApplyLabels string = "UnableToApplyLabels"

	// ReasonTypeUnableToDecommission indicates that a resource has a specified condition because SIP was unable to
	// decommission the existing SIPCluster.
	ReasonTypeUnableToDecommission string = "UnableToDecommission"

	// ReasonTypeUnschedulable indicates that a resource has a specified condition because SIP was unable to
	// schedule BMHs for the SIPCluster.
	ReasonTypeUnschedulable string = "Unschedulable"

	// ReasonTypeReconciliationSucceeded indicates that a resource has a specified condition because SIP completed
	// reconciliation of the SIPCluster.
	ReasonTypeReconciliationSucceeded string = "ReconciliationSucceeded"
)

// NodeSet are the the list of Nodes objects workers,
// or ControlPlane that define expectations
// for  the Tenant Clusters
// Includes artifacts to associate with each defined namespace
// Such as :
// - Roles for the Nodes
// - Flavor for theh Nodes image
// - Scheduling expectations
// - Scale of the group of Nodes
//
type NodeSet struct {

	// LabelSelector is the BMH label selector to use.
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
	// PlaceHolder until we define the real expected
	// Implementation
	// Scheduling define constraints that allow the SIP Scheduler
	// to identify the required BMH's to allow CAPI to build a cluster
	Scheduling SpreadTopology `json:"spreadTopology,omitempty"`
	// Count defines the scale expectations for the Nodes
	Count *NodeCount `json:"count,omitempty"`
}

// +kubebuilder:validation:Enum=PerRack;PerHost
type SpreadTopology string

const (
	// RackAntiAffinity means the scheduling should target separate racks.
	RackAntiAffinity SpreadTopology = "PerRack"

	// HostAntiAffinity means the scheduling should target separate hosts.
	HostAntiAffinity SpreadTopology = "PerHost"
)

type SIPClusterService struct {
	Image         string            `json:"image"`
	NodeLabels    map[string]string `json:"nodeLabels,omitempty"`
	NodePort      int               `json:"nodePort"`
	NodeInterface string            `json:"nodeInterfaceId,omitempty"`
	ClusterIP     *string           `json:"clusterIP,omitempty"`
}

// BMCOpts contains options for BMC communication.
type BMCOpts struct {
	Proxy bool `json:"proxy,omitempty"`
}

// BMHRole defines the states the provisioner will report
// the tenant has having.
type BMHRole string

// Possible BMH Roles for a Tenant
const (
	RoleControlPlane BMHRole = "ControlPlane"
	RoleWorker               = "Worker"
)

// NodeCount
type NodeCount struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Active  int `json:"active,omitempty"`
	Standby int `json:"standby,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SIPCluster{}, &SIPClusterList{})
}
