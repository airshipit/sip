package bmh

import (
	"fmt"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	airshipv1 "sipcluster/pkg/api/v1"
)

// ErrorConstraintNotFound is returned when wrong AuthType is provided
type ErrorConstraintNotFound struct {
}

func (e ErrorConstraintNotFound) Error() string {
	return "Invalid or Not found Schedulign Constraint"
}

type ErrorUnableToFullySchedule struct {
	TargetNode          airshipv1.BMHRole
	TargetLabelSelector metav1.LabelSelector
}

func (e ErrorUnableToFullySchedule) Error() string {
	return fmt.Sprintf("Unable to complete a schedule with a target of %v nodes, with a label selector of %v",
		e.TargetNode, e.TargetLabelSelector)
}

type ErrorHostIPNotFound struct {
	HostName    string
	IPInterface string
	Message     string
}

func (e ErrorHostIPNotFound) Error() string {
	return fmt.Sprintf("Unable to identify the vBMH Host %v IP address on interface %v required by "+
		"Infrastructure Service %s", e.HostName, e.IPInterface, e.Message)
}

// ErrorUknownSpreadTopology is returned when wrong AuthType is provided
type ErrorUknownSpreadTopology struct {
	Topology airshipv1.SpreadTopology
}

func (e ErrorUknownSpreadTopology) Error() string {
	return fmt.Sprintf("Uknown spread topology '%s'", e.Topology)
}

// ErrorNetworkDataNotFound is returned when NetworkData metadata is missing from BMH
type ErrorNetworkDataNotFound struct {
	BMH metal3.BareMetalHost
}

func (e ErrorNetworkDataNotFound) Error() string {
	return fmt.Sprintf("vBMH Host %v does not define NetworkData, but is required for scheduling.", e.BMH)
}

// ErrMalformedManagementCredentials occurs when a BMC credentials secret does not contain username and password fields.
type ErrMalformedManagementCredentials struct {
	SecretName string
}

func (e ErrMalformedManagementCredentials) Error() string {
	return fmt.Sprintf("secret %s contains malformed management credentials. Must contain '%s' and '%s' fields.",
		e.SecretName, keyBMCUsername, keyBMCPassword)
}
