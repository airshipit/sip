/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package services

import (
	"fmt"
	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoadBalancer struct {
	Service
}

func (l *LoadBalancer) Deploy(sip airshipv1.SIPCluster, machines *airshipvms.MachineList, c client.Client) error {
	// do something, might decouple this a bit
	// If the  serviucces are defined as Helm Chart , then deploy might be simply

	// Take the data from teh appropriate Machines
	// Prepare the Config
	l.Service.Deploy(sip, machines, c)
	err := l.Prepare(sip, machines, c)
	if err != nil {
		return err
	}
	return nil
}

func (l *LoadBalancer) Prepare(sip airshipv1.SIPCluster, machines *airshipvms.MachineList, c client.Client) error {
	fmt.Printf("%s.Prepare machines:%s \n", l.Service.serviceName, machines)
	for _, machine := range machines.Vbmhs {
		if machine.VmRole == airshipv1.VmMaster {
			fmt.Printf("%s.Prepare for machine:%s ip is %s\n", l.Service.serviceName, machine, machine.Data.IpOnInterface[sip.Spec.InfraServices[l.Service.serviceName].NodeInterface])
		}
	}
	return nil
}

func newLoadBalancer(infraCfg airshipv1.InfraConfig) InfrastructureService {
	return &LoadBalancer{
		Service: Service{
			serviceName: airshipv1.LoadBalancerService,
			config:      infraCfg,
		},
	}
}

/*


:::warning
For the loadbalanced interface a **static asignment** via network data is required. For now, we will not support updates to this field without manual intervention.  In other words, there is no expectation that the SIP operator watches `BareMetalHost` objects and reacts to changes in the future.  The expectation would instead to re-deliver the `SIPCluster` object to force a no-op update to load balancer configuration is updated.
:::


By extracting these IP address from the appropriate/defined interface for each master node, we can build our loadbalancer service endpoint list to feed to haproxy. In other words, the SIP Cluster will now manufacture an haproxy configuration file that directs traffic to all IP endpoints found above over port 6443.  For example:


``` gotpl
global
  log /dev/stdout local0
  log /dev/stdout local1 notice
  daemon
defaults
  log global
  mode tcp
  option dontlognull
  # TODO: tune these
  timeout connect 5000
  timeout client 50000
  timeout server 50000
frontend control-plane
  bind *:6443
  default_backend kube-apiservers
backend kube-apiservers
  option httpchk GET /healthz
{% for i in range(1, number_masters) %}
  server {{ cluster_name }}-{{ i }} {{ vm_master_ip }}:6443 check check-ssl verify none
{% end %}
```

This will be saved as a configmap and mounted into the cluster specific haproxy daemonset across all undercloud control nodes.

We will then create a Kubernetes NodePort `Service` that will direct traffic on the infrastructure `nodePort` defined in the SIP Cluster definition to these haproxy workloads.

At this point, the SIP Cluster controller can now label the VMs appropriately so they'll be scheduled by the Cluster-API process.

*/
