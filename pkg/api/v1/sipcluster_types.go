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
	// Nodes defines the set of nodes to schedule for each BMH role.
	Nodes map[BMHRole]NodeSet `json:"nodes,omitempty"`

	// Services defines the services that are deployed when a SIPCluster is provisioned.
	Services SIPClusterServices `json:"services"`
}

// SIPClusterServices defines the services that are deployed when a SIPCluster is provisioned.
type SIPClusterServices struct {
	// LoadBalancer defines the sub-cluster load balancer services.
	LoadBalancerControlPlane []LoadBalancerServiceControlPlane `json:"loadBalancerControlPlane,omitempty"`
	//  LoadBalancer defines the sub-cluster load balancer services.
	LoadBalancerWorker []LoadBalancerServiceWorker `json:"loadBalancerWorker,omitempty"`
	// JumpHost defines the sub-cluster jump host services.
	JumpHost []JumpHostService `json:"jumpHost,omitempty"`
}

func (s SIPClusterServices) GetAll() []SIPClusterService {
	all := []SIPClusterService{}
	for _, s := range s.LoadBalancerControlPlane {
		all = append(all, s.SIPClusterService)
	}
	for _, s := range s.LoadBalancerWorker {
		all = append(all, s.SIPClusterService)
	}
	for _, s := range s.JumpHost {
		all = append(all, s.SIPClusterService)
	}
	return all
}

// JumpHostService is an infrastructure service type that represents the sub-cluster jump-host service.
type JumpHostService struct {
	SIPClusterService `json:",inline"`
	NodePort          int      `json:"nodePort"`
	BMC               *BMCOpts `json:"bmc,omitempty"`
	SSHAuthorizedKeys []string `json:"sshAuthorizedKeys,omitempty"`
	// NodeSSHPrivateKeys holds the name of a Secret in the same namespace as the SIPCluster CR,
	// whose key values each represent an ssh private key that can be used to access the cluster nodes.
	// They are mounted into the jumphost with the secret keys serving as file names relative to a common
	// directory, and then configured as identity files in the SSH config file of the default user.
	NodeSSHPrivateKeys string `json:"nodeSSHPrivateKeys"`
}

/*
LoadBalancerServiceControlPlane is an infrastructure service type that represents the sub-cluster load balancer service.
*/
type LoadBalancerServiceControlPlane struct {
	SIPClusterService `json:",inline"`
	NodePort          int `json:"nodePort"`
}

// LoadBalancerServiceWorker is an infrastructure service type that represents the sub-cluster load balancer service.
type LoadBalancerServiceWorker struct {
	SIPClusterService `json:",inline"`
	// TODO: Remove the inherited single NodePort field via refactoring. It is unused for this
	// service since we have the below node port range instead.
	NodePortRange PortRange `json:"nodePortRange"`
}

// PortRange represents a range of ports.
type PortRange struct {
	// Start is the starting port number in the range.
	Start int `json:"start"`
	// End is the ending port number in the range.
	End int `json:"end"`
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
// - Flavor for the Nodes image
// - Anti-affinity expectations
// - Scale of the group of Nodes
type NodeSet struct {
	// LabelSelector is the BMH label selector to use.
	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
	// TopologyKey is similar to the same named field in the kubernetes Pod anti-affinity API.
	// If two BMHs are labeled with this key and have identical values for that
	// label, they are considered to be in the same topology domain, and thus only one will be scheduled.
	TopologyKey string `json:"topologyKey,omitempty"`
	// Count defines the scale expectations for the Nodes
	Count *NodeCount `json:"count,omitempty"`
}

type SIPClusterService struct {
	Image         string            `json:"image"`
	NodeLabels    map[string]string `json:"nodeLabels,omitempty"`
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
