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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"
)

const (
	JumpHostServiceName = "jumphost"
)

// JumpHost is an InfrastructureService that provides SSH capabilities to access a sub-cluster.
type jumpHost struct {
	client   client.Client
	sipName  types.NamespacedName
	logger   logr.Logger
	config   airshipv1.JumpHostService
	machines *airshipvms.MachineList
}

func newJumpHost(name, namespace string, logger logr.Logger, config airshipv1.JumpHostService,
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
	instance := JumpHostServiceName + "-" + jh.sipName.Name
	labels := map[string]string{
		// See https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
		"app.kubernetes.io/part-of":   "sip",
		"app.kubernetes.io/name":      JumpHostServiceName,
		"app.kubernetes.io/component": JumpHostServiceName,
		"app.kubernetes.io/instance":  instance,
	}

	// TODO: Validate Deployment becomes ready.
	deployment := jh.generateDeployment(instance, labels)
	jh.logger.Info("Applying deployment", "deployment", deployment.GetNamespace()+"/"+deployment.GetName())
	err := applyRuntimeObject(client.ObjectKey{Name: deployment.GetName(), Namespace: deployment.GetNamespace()},
		deployment, jh.client)
	if err != nil {
		return err
	}

	// TODO: Validate Service becomes ready.
	service := jh.generateService(instance, labels)
	jh.logger.Info("Applying service", "service", service.GetNamespace()+"/"+service.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: service.GetName(), Namespace: service.GetNamespace()},
		service, jh.client)
	if err != nil {
		return err
	}

	return nil
}

func (jh jumpHost) generateDeployment(instance string, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: jh.sipName.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  JumpHostServiceName,
							Image: jh.config.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "ssh",
									ContainerPort: 22,
								},
							},
						},
					},
					HostAliases: jh.generateHostAliases(),
				},
			},
		},
	}
}

func (jh jumpHost) generateHostAliases() []corev1.HostAlias {
	hostAliases := []corev1.HostAlias{}
	for _, machine := range jh.machines.Machines {
		namespace := machine.BMH.Namespace
		name := machine.BMH.Name
		ip, exists := machine.Data.IPOnInterface[jh.config.NodeInterface]
		if !exists {
			jh.logger.Info("Machine does not have ip to be aliased",
				"interface", jh.config.NodeInterface,
				"machine", namespace+"/"+name,
			)
			continue
		}
		hostname := machine.BMH.Name
		hostAliases = append(hostAliases, corev1.HostAlias{IP: ip, Hostnames: []string{hostname}})
	}
	return hostAliases
}

func (jh jumpHost) generateService(instance string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: jh.sipName.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "ssh",
					Port:     22,
					NodePort: int32(jh.config.NodePort),
				},
			},
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
}

// Finalize removes a deployed JumpHost service.
func (jh jumpHost) Finalize() error {
	// TODO(drewwalters96): Add logic to cleanup SIPCluster JumpHost pod.
	return nil
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
