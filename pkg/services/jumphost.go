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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"golang.org/x/crypto/ssh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	bmh "sipcluster/pkg/bmh"
)

const (
	JumpHostServiceName = "jumphost"

	subPathHosts     = "hosts"
	subPathSSHConfig = "ssh_config"

	sshDir             = "/home/ubuntu/.ssh"
	authorizedKeysFile = "authorized_keys"

	mountPathData               = "/etc/opt/sip"
	mountPathScripts            = "/opt/sip/bin"
	mountPathHosts              = mountPathData + "/" + subPathHosts
	mountPathSSHConfig          = sshDir + "/config"
	mountPathSSH                = sshDir + "/" + authorizedKeysFile
	mountPathNodeSSHPrivateKeys = mountPathData + "/" + nameNodeSSHPrivateKeysVolume

	nameDataVolume               = "data"
	nameScriptsVolume            = "scripts"
	nameAuthorizedKeysVolume     = "authorized-keys"
	nameNodeSSHPrivateKeysVolume = "ssh-private-keys"
)

// JumpHost is an InfrastructureService that provides SSH and power-management capabilities for sub-clusters.
type jumpHost struct {
	client   client.Client
	sipName  types.NamespacedName
	logger   logr.Logger
	config   airshipv1.JumpHostService
	machines *bmh.MachineList
}

func newJumpHost(name, namespace string, logger logr.Logger, config airshipv1.JumpHostService,
	machines *bmh.MachineList, client client.Client) InfraService {
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

	hostAliases := jh.generateHostAliases()

	// TODO: Validate Service becomes ready.
	service := jh.generateService(instance, labels)
	jh.logger.Info("Applying service", "service", service.GetNamespace()+"/"+service.GetName())
	err := applyRuntimeObject(client.ObjectKey{Name: service.GetName(), Namespace: service.GetNamespace()},
		service, jh.client)
	if err != nil {
		return err
	}

	secret, err := jh.generateSecret(instance, labels, hostAliases)
	if err != nil {
		return err
	}

	jh.logger.Info("Applying secret", "secret", secret.GetNamespace()+"/"+secret.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: secret.GetName(), Namespace: secret.GetNamespace()},
		secret, jh.client)
	if err != nil {
		return err
	}

	configMap, err := jh.generateConfigMap(instance, labels)
	if err != nil {
		return err
	}
	jh.logger.Info("Applying configmap", "configmap", configMap.GetNamespace()+"/"+configMap.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: configMap.GetName(), Namespace: configMap.GetNamespace()},
		configMap, jh.client)
	if err != nil {
		return err
	}

	// TODO: Validate Deployment becomes ready.
	deployment := jh.generateDeployment(instance, labels, hostAliases)
	jh.logger.Info("Applying deployment", "deployment", deployment.GetNamespace()+"/"+deployment.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: deployment.GetName(), Namespace: deployment.GetNamespace()},
		deployment, jh.client)
	if err != nil {
		return err
	}

	return nil
}

func (jh jumpHost) generateDeployment(instance string, labels map[string]string,
	hostAliases []corev1.HostAlias) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
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
					InitContainers: []corev1.Container{
						{
							Name:            "ssh-perms",
							Image:           jh.config.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         []string{"/bin/sh", "-c"},
							Args: []string{
								fmt.Sprintf("mkdir -m700 %s", sshDir),
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            JumpHostServiceName,
							Image:           jh.config.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "ssh",
									ContainerPort: 22,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      nameDataVolume,
									MountPath: mountPathHosts,
									SubPath:   subPathHosts,
								},
								{
									Name:      nameScriptsVolume,
									MountPath: mountPathScripts,
								},
								{
									Name:      nameDataVolume,
									MountPath: mountPathSSHConfig,
									SubPath:   subPathSSHConfig,
								},
								{
									Name:      nameNodeSSHPrivateKeysVolume,
									MountPath: mountPathNodeSSHPrivateKeys,
								},
								{
									Name:      nameAuthorizedKeysVolume,
									MountPath: mountPathSSH,
									SubPath:   authorizedKeysFile,
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString("ssh"),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString("ssh"),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       20,
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1000),
					},
					Volumes: []corev1.Volume{
						{
							Name: nameDataVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: instance,
								},
							},
						},
						{
							Name: nameAuthorizedKeysVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance,
									},
									DefaultMode: int32Ptr(0700),
									Items: []corev1.KeyToPath{
										{
											Key:  nameAuthorizedKeysVolume,
											Path: authorizedKeysFile,
											Mode: int32Ptr(0600),
										},
									},
								},
							},
						},
						{
							Name: nameScriptsVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: instance,
									},
									DefaultMode: int32Ptr(0777),
								},
							},
						},
						{
							Name: nameNodeSSHPrivateKeysVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: jh.config.NodeSSHPrivateKeys,
								},
							},
						},
					},
					HostAliases: hostAliases,
				},
			},
		},
	}

	// Set NO_PROXY env variables when Redfish proxy setting is false (Default: false).
	var proxy bool
	if jh.config.BMC != nil {
		proxy = jh.config.BMC.Proxy
	}
	if proxy == false {
		// TODO: We may need to identify the container with Redfish functionality in the future if some
		// containers require communication over a proxy server.
		for _, container := range deployment.Spec.Template.Spec.Containers {
			container.Env = []corev1.EnvVar{
				{
					Name:  "NO_PROXY",
					Value: "*",
				},
				{
					Name:  "no_proxy",
					Value: "*",
				},
			}
		}
	}

	return deployment
}

func (jh jumpHost) generateConfigMap(instance string, labels map[string]string) (*corev1.ConfigMap, error) {
	for _, key := range jh.config.SSHAuthorizedKeys {
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			return nil, ErrInvalidAuthorizedKeyFormat{SSHErr: err.Error(), Key: key}
		}
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: jh.sipName.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			nameAuthorizedKeysVolume: strings.Join(jh.config.SSHAuthorizedKeys, "\n"),
			"host":                   fmt.Sprintf(rebootScript, mountPathHosts),
		},
	}, nil
}

func (jh jumpHost) generateSecret(instance string, labels map[string]string, hostAliases []corev1.HostAlias) (
	*corev1.Secret, error) {
	hostData, err := generateHostList(*jh.machines)
	if err != nil {
		return nil, err
	}
	sshConfig, err := jh.generateSSHConfig(hostAliases)
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: jh.sipName.Namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			subPathHosts:     hostData,
			subPathSSHConfig: sshConfig,
		},
	}

	return secret, nil
}

func (jh jumpHost) generateSSHConfig(hostAliases []corev1.HostAlias) ([]byte, error) {
	key := types.NamespacedName{
		Namespace: jh.sipName.Namespace,
		Name:      jh.config.NodeSSHPrivateKeys,
	}
	secret := &corev1.Secret{}
	if err := jh.client.Get(context.Background(), key, secret); err != nil {
		return nil, err
	}

	identityFiles := []string{}
	for k := range secret.Data {
		identityFiles = append(identityFiles, mountPathNodeSSHPrivateKeys+"/"+k)
	}
	hostNames := []string{}
	for _, hostAlias := range hostAliases {
		hostNames = append(hostNames, hostAlias.Hostnames[0])
	}

	tmpl, err := template.New("ssh-config").Parse(sshConfigTemplate)
	if err != nil {
		return nil, err
	}

	data := sshConfigTemplateData{
		IdentityFiles: identityFiles,
		HostNames:     hostNames,
	}

	w := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(w, data); err != nil {
		return nil, err
	}

	rendered := w.Bytes()
	return rendered, nil
}

type sshConfigTemplateData struct {
	IdentityFiles []string
	HostNames     []string
}

const sshConfigTemplate = `
Host *
{{- range .IdentityFiles }}
	IdentityFile {{ . }}
{{ end -}}

{{- range .HostNames }}
Host {{ . }}
	HostName {{ . }}
{{ end -}}
`

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

type host struct {
	Name string `json:"name"`
	BMC  bmc    `json:"bmc"`
}

type bmc struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// generateHostList creates a list of hosts in JSON format to be mounted as a config map to the jump host pod and used
// to power cycle sub-cluster nodes.
func generateHostList(machineList bmh.MachineList) ([]byte, error) {
	hosts := make([]host, 0)
	for name, machine := range machineList.Machines {
		managementIP, err := getManagementIP(machine.BMH.Spec.BMC.Address)
		if err != nil {
			return nil, err
		}

		h := host{
			Name: name,
			BMC: bmc{
				IP:       managementIP,
				Username: machine.Data.BMCUsername,
				Password: machine.Data.BMCPassword,
			},
		}

		hosts = append(hosts, h)
	}

	out, err := json.Marshal(hosts)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// getManagementIP parses the BMC IP address from a Redfish fully qualified domain name. For example, the input
// redfish+https://127.0.0.1/redfish/v1/Systems/System.Embedded.1 yields 127.0.0.1.
func getManagementIP(redfishURL string) (string, error) {
	parsedURL, err := url.Parse(redfishURL)
	if err != nil {
		return "", ErrMalformedRedfishAddress{Address: redfishURL}
	}

	return parsedURL.Host, nil
}

var rebootScript = `#!/bin/sh

# Support Infrastructure Provider (SIP) Host Utility
# DO NOT MODIFY: generated by SIP

HOSTS_FILE="%s"

LIST_COMMAND="list"
REBOOT_COMMAND="reboot"

help() {
  echo "Support Infrastructure Provider (SIP) Host Utility"
  echo ""
  echo "Usage: ${LIST_COMMAND}                      list hosts"
  echo "       ${REBOOT_COMMAND} [host name]        reboot host"
}

dep_check() {
  if [ "$(which jq)" = "" ]; then
    echo "Missing package 'jq'. Update your JumpHost image to include 'jq' and 'redfishtool'."
    exit 1
  fi

  if [ "$(which redfishtool)" = "" ]; then
    echo "Missing package 'redfishtool'. Update your JumpHost image to include 'jq' and 'redfishtool'."
    exit 1
  fi
}

get_bmc_info() {
  for host in $(jq -r -c '.[]' ${HOSTS_FILE}); do
    if [ "$(echo "$host" | jq -r '.name')" = "$1" ]; then
      addr=$(echo "$host" | jq -r '.bmc.ip')
      user=$(echo "$host" | jq -r '.bmc.username')
      pass=$(echo "$host" | jq -r '.bmc.password')
    fi
  done
}

reboot() {
  get_bmc_info "$1"
  if [ "${addr}" = "" ] || [ "${user}" = "" ] || [ "$pass" = "" ]; then
    echo "Invalid host '$1'. Use the '${LIST_COMMAND}' command to view hosts."
    exit 1
  fi

  echo "Rebooting host '$1'"
  redfishtool -r "${addr}" -u "${user}" -p "${pass}" \
    Systems reset GracefulRestart -vvvvv
  exit 0
}

case $1 in
  "${LIST_COMMAND}")
    dep_check
    jq -r '.[].name' ${HOSTS_FILE}
    ;;
  "${REBOOT_COMMAND}")
    if [ "$2" = "" ]; then
      printf "Host name required.\n\n"
      help
      exit 1
    fi
    dep_check
    reboot "$2"
    ;;
  *)
    help
    ;;
esac
`

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
