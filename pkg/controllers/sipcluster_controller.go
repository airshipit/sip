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
	"fmt"

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
	log := r.Log.WithValues("sipcluster", req.NamespacedName)

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
		err := r.finalize(sip)
		if err != nil {
			log.Error(err, "unable to finalize")
			return ctrl.Result{}, err
		}

	}
	return ctrl.Result{}, nil
}

func (r *SIPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&airshipv1.SIPCluster{}).
		Complete(r)
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
	machines := &airshipvms.MachineList{}
	fmt.Printf("gatherVBMH.Schedule sip:%v machines:%v\n", sip, machines)
	err := machines.Schedule(sip, r.Client)
	if err != nil {
		return err, machines
	}

	// we extract the information in a generic way
	// So that LB ,  Jump and Ath POD  all leverage the same
	// If there are some issues finnding information then the machines ??
	fmt.Printf("gatherVBMH.Extrapolate sip:%v machines:%v\n", sip, machines)
	err = machines.Extrapolate(sip, r.Client)
	if err != nil {
		return err, machines
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
		err = service.Deploy(machines, r.Client)
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

	// Label the vBMH's
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
	// 1- Let me retrieve all vBMH mapped for this SIP Cluster
	// 2- Let me now select the one's that meet teh scheduling criteria
	// If I schedule successfully then
	// If Not complete schedule , then throw an error.
	machines := &airshipvms.MachineList{}
	fmt.Printf("finalize sip:%v machines:%s\n", sip, machines.String())
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
