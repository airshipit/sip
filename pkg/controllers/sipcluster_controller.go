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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipsvc "sipcluster/pkg/services"
	airshipvms "sipcluster/pkg/vbmh"
)

// SIPClusterReconciler reconciles a SIPCluster object
type SIPClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=airship.airshipit.org,resources=sipclusters/status,verbs=get;update;patch

// +kubebuilder:rbac:groups="metal3.io",resources=baremetalhosts,verbs=get;update;patch;list

func (r *SIPClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("SIPcluster", req.NamespacedName)

	// Lets retrieve the SIPCluster
	sip := airshipv1.SIPCluster{}
	if err := r.Get(ctx, req.NamespacedName, &sip); err != nil {
		//log.Error(err, "unable to fetch SIP Cluster")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}

	// Check for Deletion
	// name of our custom finalizer
	sipFinalizerName := "sip.airship.airshipit.org/finalizer"
	// Tghis only works if I add a finalizer to CRD TODO
	if sip.ObjectMeta.DeletionTimestamp.IsZero() {
		// machines
		err, machines := r.gatherVBMH(sip)
		if err != nil {
			//log.Error(err, "unable to gather vBMHs")
			return ctrl.Result{}, err
		}

		err = r.deployInfra(sip, machines)
		if err != nil {
			log.Error(err, "unable to deploy infrastructure services")
			return ctrl.Result{}, err
		}

		err = r.finish(sip, machines)
		if err != nil {
			log.Error(err, "unable to finish creation/update ..")
			return ctrl.Result{}, err
		}
	} else {
		// Deleting the SIP , what do we do now
		if containsString(sip.ObjectMeta.Finalizers, sipFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			err := r.finalize(sip)
			if err != nil {
				log.Error(err, "unable to finalize")
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			sip.ObjectMeta.Finalizers = removeString(sip.ObjectMeta.Finalizers, sipFinalizerName)
			if err := r.Update(context.Background(), &sip); err != nil {
				return ctrl.Result{}, err
			}
		}

	}
	return ctrl.Result{}, nil
}

func (r *SIPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airshipv1.SIPCluster{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
// There might be a golang funuction to do this . Will check later
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

/*
### Gather Phase

#### Identity BMH VM's
- Gather BMH's that meet the criteria expected for the groups
- Check for existing labeled BMH's
- Complete the expected scheduling contraints :
    - If master
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
func (r *SIPClusterReconciler) gatherVBMH(sip airshipv1.SIPCluster) (error, *airshipvms.MachineList) {
	// 1- Let me retrieve all BMH  that are unlabeled or already labeled with the target Tenant/CNF
	// 2- Let me now select the one's that meet teh scheduling criteria
	// If I schedule successfully then
	// If Not complete schedule , then throw an error.
	logger := r.Log.WithValues("SIPCluster", sip.GetNamespace()+"/"+sip.GetName())
	logger.Info("starting to gather BaremetalHost machines for SIPcluster")
	machines := &airshipvms.MachineList{}

	// TODO : this is a loop until we succeed or cannot find a schedule
	for {
		logger.Info("gathering machines, so these machines are collected", "machines", machines.String())
		err := machines.Schedule(sip, r.Client)
		if err != nil {
			return err, machines
		}

		// we extract the information in a generic way
		// So that LB ,  Jump and Ath POD  all leverage the same
		// If there are some issues finnding information the vBMH
		// Are flagged Unschedulable
		// Loop and Try to find new vBMH to complete tge schedule
		//fmt.Printf("gatherVBMH.Extrapolate sip:%v machines:%v\n", sip, machines)
		if machines.Extrapolate(sip, r.Client) {
			logger.Info("successfuly extrapolated machines")
			break
		}
	}

	return nil, machines
}

func (r *SIPClusterReconciler) deployInfra(sip airshipv1.SIPCluster, machines *airshipvms.MachineList) error {
	for sName, sConfig := range sip.Spec.InfraServices {
		// Instantiate
		service, err := airshipsvc.NewService(sName, sConfig)
		if err != nil {
			return err
		}

		// Lets deploy the Service
		err = service.Deploy(sip, machines, r.Client)
		if err != nil {
			return err
		}

		// Did it deploy correctly, letcs check

		err = service.Validate()
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
	err := machines.ApplyLabels(sip, r.Client)
	if err != nil {
		return err
	}
	return nil

}

/**

Deal with Deletion andd Finalizers if any is needed
Such as i'e what are we doing with the lables on teh vBMH's
**/
func (r *SIPClusterReconciler) finalize(sip airshipv1.SIPCluster) error {
	logger := r.Log.WithValues("SIPCluster", sip.GetNamespace()+"/"+sip.GetName())
	for sName, sConfig := range sip.Spec.InfraServices {
		// Instantiate
		service, err := airshipsvc.NewService(sName, sConfig)
		if err != nil {
			return err
		}

		// Lets clean  Service specific stuff
		err = service.Finalize(sip, r.Client)
		if err != nil {
			return err
		}
	}
	// Clean Up common servicce stuff
	airshipsvc.FinalizeCommon(sip, r.Client)

	// 1- Let me retrieve all vBMH mapped for this SIP Cluster
	// 2- Let me now select the one's that meet teh scheduling criteria
	// If I schedule successfully then
	// If Not complete schedule , then throw an error.
	machines := &airshipvms.MachineList{}
	logger.Info("finalize sip machines", "machines", machines.String())
	// Update the list of  Machines.
	err := machines.GetCluster(sip, r.Client)
	if err != nil {
		return err
	}
	// Placeholder unsuree whether this is what we want
	err = machines.RemoveLabels(sip, r.Client)
	if err != nil {
		return err
	}
	return nil
}
