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

	// ClusterName is the name of the cluster to associate machines with
	ClusterName string `json:"cluster-name,omitempty"`
	// Nodes are the list of Nodes objects workers, or master that definee eexpectations
	// of the Tenant cluster
	// VMRole is either Control or Workers
	// VMRole VMRoles `json:"vm-role,omitempty"`
	Nodes map[VMRoles]NodeSet `json:"nodes,omitempty"`

	// Infra is the collection of expeected configuration details
	// for the multiple infrastructure services or pods that SIP manages
	//Infra *InfraSet `json:"infra,omitempty"`

	// List of Infrastructure Services
	InfraServices map[InfraService]InfraConfig `json:"infra"`
}

// SIPClusterStatus defines the observed state of SIPCluster
type SIPClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// VMRoles defines the states the provisioner will report
// the tenant has having.
type InfraService string

// Possible Infra Structure Services
const (
	// LoadBalancer Service
	LoadBalancerService InfraService = "loadbalancer"

	// JumpHostService Service
	JumpHostService InfraService = "jumppod"

	// AuthHostService Service
	AuthHostService InfraService = "authpod"
)

// NodeSet are the the list of Nodes objects workers,
// or master that definee eexpectations
// for  the Tenant Clusters
// Includes artifacts to associate with each defined namespace
// Such as :
// - Roles for the Nodes
// - Flavor for theh Nodes image
// - Scheduling expectations
// - Scale of the group of Nodes
//
type NodeSet struct {

	// VMFlavor is  essentially a Flavor label identifying the
	// type of Node that meets the construction reqirements
	VMFlavor string `json:"vm-flavor,omitempty"`
	// PlaceHolder until we define the real expected
	// Implementation
	// Scheduling define constraints the allows the SIP Scheduler
	// to identify the required  BMH's to allow CAPI to build a cluster
	Scheduling SpreadTopology `json:"spreadTopology,omitempty"`
	// Count defines the scale expectations for the Nodes
	Count *VMCount `json:"count,omitempty"`
}

type SpreadTopology string

// Possible Node or VM Roles  for a Tenant
const (
	// RackAntiAffinity means the state is unknown
	RackAntiAffinity SpreadTopology = "per-rack"

	// ServerAntiAffinity means the state is unknown
	ServerAntiAffinity SpreadTopology = "per-node"
)

type InfraConfig struct {
	OptionalData  *OptsConfig       `json:"optional,omitempty"`
	Image         string            `json:"image,omitempty"`
	NodeLabels    map[string]string `json:"nodelabels,omitempty"`
	NodePort      int               `json:"nodePort,omitempty"`
	NodeInterface string            `json:"nodeInterfaceId,omitempty"`
}

type OptsConfig struct {
	SSHKey    string `json:"sshkey,omitempty"`
	ClusterIP string `json:"clusterIP,omitempty"`
}

// VMRoles defines the states the provisioner will report
// the tenant has having.
type VMRoles string

// Possible Node or VM Roles  for a Tenant
const (
	VMMaster VMRoles = "master"
	VMWorker         = "worker"
)

// VMCount
type VMCount struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Active  int `json:"active,omitempty"`
	Standby int `json:"standby,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SIPCluster{}, &SIPClusterList{})
}
