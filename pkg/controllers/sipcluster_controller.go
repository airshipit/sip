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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerror "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipsvc "sipcluster/pkg/services"
	airshipvms "sipcluster/pkg/vbmh"
)

// SIPClusterReconciler reconciles a SIPCluster object
type SIPClusterReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	NamespacedName types.NamespacedName
}

const (
	sipFinalizerName = "sip.airship.airshipit.org/finalizer"
)

// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters/status,verbs=get;update;patch

// +kubebuilder:rbac:groups="metal3.io",resources=baremetalhosts,verbs=get;update;patch;list

func (r *SIPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.NamespacedName = req.NamespacedName
	log := logr.FromContext(ctx)

	sip := airshipv1.SIPCluster{}
	if err := r.Get(ctx, req.NamespacedName, &sip); err != nil {
		log.Error(err, "unable to fetch SIPCluster")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}

	readyCondition := metav1.Condition{
		Status:             metav1.ConditionFalse,
		Reason:             airshipv1.ReasonTypeProgressing,
		Type:               airshipv1.ConditionTypeReady,
		ObservedGeneration: sip.GetGeneration(),
	}

	apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
	if err := r.patchStatus(ctx, &sip); err != nil {
		log.Error(err, "unable to set condition", "condition", readyCondition)
		return ctrl.Result{Requeue: true}, err
	}

	if !sip.ObjectMeta.DeletionTimestamp.IsZero() {
		// SIPCluster is being deleted; handle the finalizers, then stop reconciling
		// TODO(howell): add finalizers to the CRD
		if containsString(sip.ObjectMeta.Finalizers, sipFinalizerName) {
			result, err := r.handleFinalizers(ctx, sip)
			if err != nil {
				readyCondition = metav1.Condition{
					Status:             metav1.ConditionFalse,
					Reason:             airshipv1.ReasonTypeUnableToDecommission,
					Type:               airshipv1.ConditionTypeReady,
					Message:            err.Error(),
					ObservedGeneration: sip.GetGeneration(),
				}

				apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
				if patchStatusErr := r.patchStatus(ctx, &sip); err != nil {
					err = kerror.NewAggregate([]error{err, patchStatusErr})
					log.Error(err, "unable to set condition", "condition", readyCondition)
				}

				log.Error(err, "unable to finalize")
				return ctrl.Result{Requeue: true}, err
			}

			return result, err
		}
		return ctrl.Result{}, nil
	}

	machines, err := r.gatherVBMH(ctx, sip)
	if err != nil {
		readyCondition = metav1.Condition{
			Status:             metav1.ConditionFalse,
			Reason:             airshipv1.ReasonTypeUnschedulable,
			Type:               airshipv1.ConditionTypeReady,
			Message:            err.Error(),
			ObservedGeneration: sip.GetGeneration(),
		}

		apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
		if patchStatusErr := r.patchStatus(ctx, &sip); err != nil {
			err = kerror.NewAggregate([]error{err, patchStatusErr})
			log.Error(err, "unable to set condition", "condition", readyCondition)
		}

		log.Error(err, "unable to gather vBMHs")
		return ctrl.Result{Requeue: true}, err
	}

	err = r.deployInfra(sip, machines, log)
	if err != nil {
		readyCondition = metav1.Condition{
			Status:             metav1.ConditionFalse,
			Reason:             airshipv1.ReasonTypeInfraServiceFailure,
			Type:               airshipv1.ConditionTypeReady,
			Message:            err.Error(),
			ObservedGeneration: sip.GetGeneration(),
		}

		apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
		if patchStatusErr := r.patchStatus(ctx, &sip); err != nil {
			err = kerror.NewAggregate([]error{err, patchStatusErr})
			log.Error(err, "unable to set condition", "condition", readyCondition)
		}

		log.Error(err, "unable to deploy infrastructure services")
		return ctrl.Result{Requeue: true}, err
	}

	err = r.finish(sip, machines)
	if err != nil {
		readyCondition = metav1.Condition{
			Status:             metav1.ConditionFalse,
			Reason:             airshipv1.ReasonTypeUnableToApplyLabels,
			Type:               airshipv1.ConditionTypeReady,
			Message:            err.Error(),
			ObservedGeneration: sip.GetGeneration(),
		}

		apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
		if patchStatusErr := r.patchStatus(ctx, &sip); err != nil {
			err = kerror.NewAggregate([]error{err, patchStatusErr})
			log.Error(err, "unable to set condition", "condition", readyCondition)
		}

		log.Error(err, "unable to finish reconciliation")
		return ctrl.Result{Requeue: true}, err
	}

	readyCondition = metav1.Condition{
		Status:             metav1.ConditionTrue,
		Reason:             airshipv1.ReasonTypeReconciliationSucceeded,
		Type:               airshipv1.ConditionTypeReady,
		ObservedGeneration: sip.GetGeneration(),
	}

	apimeta.SetStatusCondition(&sip.Status.Conditions, readyCondition)
	if patchStatusErr := r.patchStatus(ctx, &sip); err != nil {
		err = kerror.NewAggregate([]error{err, patchStatusErr})
		log.Error(err, "unable to set condition", "condition", readyCondition)

		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *SIPClusterReconciler) patchStatus(ctx context.Context, sip *airshipv1.SIPCluster) error {
	key := client.ObjectKeyFromObject(sip)
	latest := &airshipv1.SIPCluster{}

	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	return r.Client.Status().Patch(ctx, sip, client.MergeFrom(latest))
}

func (r *SIPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airshipv1.SIPCluster{}, builder.WithPredicates(
			predicate.GenerationChangedPredicate{},
		)).
		Complete(r)
}

func (r *SIPClusterReconciler) handleFinalizers(ctx context.Context, sip airshipv1.SIPCluster) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	err := r.finalize(ctx, sip)
	if err != nil {
		log.Error(err, "unable to finalize")
		return ctrl.Result{}, err
	}

	// remove the finalizer from the list and update it.
	sip.ObjectMeta.Finalizers = removeString(sip.ObjectMeta.Finalizers, sipFinalizerName)
	return ctrl.Result{}, r.Update(context.Background(), &sip)
}

// containsString is a helper function to check whether the string s is in the slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString is a helper function to remove the string s from the slice.
// if s is not in the slice, the original slice is returned
func removeString(slice []string, s string) []string {
	for i, item := range slice {
		if item == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

/*
### Gather Phase

#### Identity BMH VM's
- Gather BMH's that meet the criteria expected for the groups
- Check for existing labeled BMH's
- Complete the expected scheduling contraints :
    - If ControlPlane
        -  collect into list of bmh's to label
    - If worker
        - collect into list of bmh's to label
#### Extract Info from Identified BMH
-  identify and extract  the IP address ands other info as needed (***)
    -  Use it as part of the service infrastucture configuration
- At this point I have a list of BMH's, and I have the extrapolated data I need for configuring services.

### Service Infrastructure Deploy Phase
- Create or Updated the [LB|admin pod] with the appropriate configuration

### Label Phase
- Label the collected hosts.
- At this point SIPCluster is done processing a given CR, and can move on the next.
*/

// machines
func (r *SIPClusterReconciler) gatherVBMH(ctx context.Context, sip airshipv1.SIPCluster) (
	*airshipvms.MachineList, error) {
	// 1- Let me retrieve all BMH  that are unlabeled or already labeled with the target Tenant/CNF
	// 2- Let me now select the one's that meet the scheduling criteria
	// If I schedule successfully then
	// If Not complete schedule , then throw an error.
	logger := logr.FromContext(ctx)
	logger.Info("starting to gather BaremetalHost machines for SIPcluster")
	machines := &airshipvms.MachineList{
		Log:            logger.WithName("machines"),
		NamespacedName: r.NamespacedName,
	}
	// TODO : this is a loop until we succeed or cannot find a schedule
	for {
		logger.Info("gathering machines", "machines", machines.String())

		// NOTE: Schedule executes the scheduling algorithm to find hosts that meet the topology and role
		// constraints.
		err := machines.Schedule(sip, r.Client)
		if err != nil {
			return machines, err
		}

		if err = machines.ExtrapolateServiceAddresses(sip, r.Client); err != nil {
			logger.Error(err, "unable to retrieve infrastructure service IP addresses from selected BMHs."+
				"Selecting replacement hosts.")

			continue
		}

		if err = machines.ExtrapolateBMCAuth(sip, r.Client); err != nil {
			logger.Error(err, "unable to retrieve BMC auth info from selected BMHs. Selecting replacement"+
				"hosts.")

			continue
		}

		break
	}

	return machines, nil
}

func (r *SIPClusterReconciler) deployInfra(sip airshipv1.SIPCluster, machines *airshipvms.MachineList,
	logger logr.Logger) error {
	newServiceSet := airshipsvc.NewServiceSet(logger, sip, machines, r.Client)
	serviceList, err := newServiceSet.ServiceList()
	if err != nil {
		return err
	}
	for _, svc := range serviceList {
		err := svc.Deploy()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
finish shoulld  take care of any wrpa up tasks..
*/
func (r *SIPClusterReconciler) finish(sip airshipv1.SIPCluster, machines *airshipvms.MachineList) error {
	// UnLabel the vBMH's
	return machines.ApplyLabels(sip, r.Client)
}

/**
Deal with Deletion and Finalizers if any is needed
Such as i'e what are we doing with the lables on the vBMH's
**/
func (r *SIPClusterReconciler) finalize(ctx context.Context, sip airshipv1.SIPCluster) error {
	logger := logr.FromContext(ctx)
	machines := &airshipvms.MachineList{}
	serviceSet := airshipsvc.NewServiceSet(logger, sip, machines, r.Client)
	serviceList, err := serviceSet.ServiceList()
	if err != nil {
		return err
	}
	for _, svc := range serviceList {
		err = svc.Finalize()
		if err != nil {
			return err
		}
	}
	err = serviceSet.Finalize()
	if err != nil {
		return err
	}

	// 1- Let me retrieve all vBMH mapped for this SIP Cluster
	// 2- Let me now select the one's that meet the scheduling criteria
	// If I schedule successfully then
	// If Not complete schedule , then throw an error.
	logger.Info("finalize sip machines", "machines", machines.String())
	// Update the list of  Machines.
	err = machines.GetCluster(sip, r.Client)
	if err != nil {
		return err
	}

	// Placeholder - unsure whether this is what we want
	err = machines.RemoveLabels(r.Client)
	if err != nil {
		return err
	}
	return nil
}
