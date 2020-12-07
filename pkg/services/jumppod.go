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
	airshipv1 "sipcluster/pkg/api/v1"
)

type JumpHost struct {
	Service
}

func newJumpHost(infraCfg airshipv1.InfraConfig) InfrastructureService {
	return &JumpHost{
		Service: Service{
			serviceName: airshipv1.JumpHostService,
			config:      infraCfg,
		},
	}
}

/*

The SIP Cluster operator will manufacture a jump host pod specifically for this
tenant cluster.  Much like we did above for master nodes by extracting IP
addresses, we would need to extract the `oam-ipv4` ip address for all nodes and
create a configmap to bind mount into the pod so it understands what host IPs
represent the clusters.

The expectation is the Jump Pod runs `sshd` protected by `uam` to allow
operators to SSH directly to the Jump Pod and authenticate via UAM to
immediately access their cluster.

It will provide the following functionality over SSH:

- The Jump Pod will be fronted by a `NodePort` service to allow incoming ssh.
- The Jump Pod will be UAM secured (for SSH)
- Bind mount in cluster-specific SSH key for cluster
- Ability to Power Cycle the cluster VMs
- A kubectl binary and kubeconfig (cluster-admin) for the cluster
- SSH access to the cluster node VMs
- Libvirt console logs for the VMs
  - We will secure libvirt with tls and provide keys to every jump host
    with curated interfaces to extract logs remotely for all VMs for their
    clusters.

*/
