package vbmh

import (
	"fmt"
	airshipv1 "sipcluster/pkg/api/v1"
)

// ErrAuthTypeNotSupported is returned when wrong AuthType is provided
type ErrorConstraintNotFound struct {
}

func (e ErrorConstraintNotFound) Error() string {
	return "Invalid or Not found Schedulign Constraint"
}

type ErrorUnableToFullySchedule struct {
	TargetNode   airshipv1.VmRoles
	TargetFlavor string
}

func (e ErrorUnableToFullySchedule) Error() string {
	return fmt.Sprintf("Unable to complete a schedule targetting %v nodes, which a floavor of %v ", e.TargetNode, e.TargetFlavor)
}
