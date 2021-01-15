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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"
)

const (
	jumpHostContainerName   = "ssh"
	jumpHostPodNameTemplate = "%s-jump-pod"
)

// JumpHost is an InfrastructureService that provides SSH capabilities to access a sub-cluster.
type jumpHost struct {
	client   client.Client
	sipName  types.NamespacedName
	logger   logr.Logger
	config   airshipv1.InfraConfig
	machines *airshipvms.MachineList
}

func newJumpHost(name, namespace string, logger logr.Logger, config airshipv1.InfraConfig,
	machines *airshipvms.MachineList, client client.Client) InfraService {
	return jumpHost{
		sipName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		logger:   logger,
		config:   config,
		machines: machines,
		client:   client,
	}
}

// Deploy creates a JumpHost service in the base cluster.
func (jh jumpHost) Deploy() error {
	jh.logger.Info("deploying jump host", "sub-cluster", jh.sipName.Name)

	jumpHostPodName := fmt.Sprintf(jumpHostPodNameTemplate, jh.sipName.Name)
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jumpHostPodName,
			Namespace: jh.sipName.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  jumpHostContainerName,
					Image: jh.config.Image,
				},
			},
		},
	}

	if err := jh.client.Create(context.Background(), &pod); err != nil {
		return err
	}

	jh.logger.Info("successfully deployed jump host", "sub-cluster", jh.sipName.Name, "jump host pod name",
		jumpHostPodName, "namespace", jh.sipName.Namespace)

	return nil
}

// Finalize removes a deployed JumpHost service.
func (jh jumpHost) Finalize() error {
	// TODO(drewwalters96): Add logic to cleanup SIPCluster JumpHost pod.
	return nil
}

// Type returns the type of infrastructure service: jumphost.
func (jh jumpHost) Type() airshipv1.InfraService {
	return airshipv1.JumpHostService
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
