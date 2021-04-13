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
	"fmt"
	"strings"

	"html/template"
	airshipv1 "sipcluster/pkg/api/v1"
	bmh "sipcluster/pkg/bmh"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ConfigSecretName name of the haproxy config secret name/volume/mount
	/* #nosec */
	ConfigSecretName = "haproxy-config"
	// DefaultBalancerImage is the image that will be used as load balancer
	DefaultBalancerImage    = "haproxy:2.3.2"
	LoadBalancerServiceName = "loadbalancer"
)

func (lb loadBalancer) Deploy() error {
	if lb.config.Image == "" {
		lb.config.Image = DefaultBalancerImage
	}

	instance := LoadBalancerServiceName + "-" + strings.ToLower(string(lb.bmhRole)) + "-" + lb.sipName.Name
	labels := map[string]string{
		// See https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
		"app.kubernetes.io/part-of":   "sip",
		"app.kubernetes.io/component": LoadBalancerServiceName,
		"app.kubernetes.io/name":      "haproxy",
		"app.kubernetes.io/instance":  instance,
	}

	deployment, secret, err := lb.generateDeploymentAndSecret(instance, labels)
	if err != nil {
		return err
	}

	lb.logger.Info("Applying loadbalancer secret", "secret", secret.GetNamespace()+"/"+secret.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: secret.GetName(), Namespace: secret.GetNamespace()}, secret, lb.client)
	if err != nil {
		return err
	}

	// TODO: Validate Deployment becomes ready.
	lb.logger.Info("Applying loadbalancer deployment", "deployment", deployment.GetNamespace()+"/"+deployment.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: deployment.GetName(), Namespace: deployment.GetNamespace()},
		deployment, lb.client)
	if err != nil {
		return err
	}

	// TODO: Validate Service becomes ready.
	lbService := lb.generateService(instance, labels)
	lb.logger.Info("Applying loadbalancer service", "service", lbService.GetNamespace()+"/"+lbService.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: lbService.GetName(), Namespace: lbService.GetNamespace()},
		lbService, lb.client)
	if err != nil {
		return err
	}
	return nil
}

func (lb loadBalancer) generateDeploymentAndSecret(instance string, labels map[string]string) (*appsv1.Deployment,
	*corev1.Secret, error) {
	secret, err := lb.generateSecret(instance)
	if err != nil {
		return nil, nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: lb.sipName.Namespace,
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
							Name:            LoadBalancerServiceName,
							Image:           lb.config.Image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports:           lb.getContainerPorts(),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      ConfigSecretName,
									MountPath: "/usr/local/etc/haproxy",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: ConfigSecretName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secret.GetName(),
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, secret, nil
}

func (lb loadBalancer) getContainerPorts() []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{}
	for _, servicePort := range lb.servicePorts {
		containerPorts = append(containerPorts, corev1.ContainerPort{
			Name:          servicePort.Name,
			ContainerPort: servicePort.Port,
		})
	}
	return containerPorts
}

func (lb loadBalancer) generateSecret(instance string) (*corev1.Secret, error) {
	p := proxy{
		ContainerPorts: lb.getContainerPorts(),
		Servers:        make([]server, 0),
	}
	for _, machine := range lb.machines.Machines {
		if machine.BMHRole == lb.bmhRole {
			name := machine.BMH.Name
			namespace := machine.BMH.Namespace
			ip, exists := machine.Data.IPOnInterface[lb.config.NodeInterface]
			if !exists {
				lb.logger.Info("Machine does not have backend interface to be forwarded to",
					"interface", lb.config.NodeInterface,
					"machine", namespace+"/"+name,
				)
				continue
			}
			p.Servers = append(p.Servers, server{IP: ip, Name: machine.BMH.Name})
		}
	}
	secretData, err := lb.generateTemplate(p)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: lb.sipName.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"haproxy.cfg": secretData,
		},
	}, nil
}

func (lb loadBalancer) generateService(instance string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance,
			Namespace: lb.sipName.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    lb.servicePorts,
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
}

type proxy struct {
	ContainerPorts []corev1.ContainerPort
	Servers        []server
}

type server struct {
	IP   string
	Name string
}

type loadBalancer struct {
	client       client.Client
	sipName      types.NamespacedName
	logger       logr.Logger
	config       airshipv1.SIPClusterService
	machines     *bmh.MachineList
	bmhRole      airshipv1.BMHRole
	template     string
	servicePorts []corev1.ServicePort
}

type loadBalancerControlPlane struct {
	loadBalancer
	config airshipv1.LoadBalancerServiceControlPlane
}

type loadBalancerWorker struct {
	loadBalancer
	config airshipv1.LoadBalancerServiceWorker
}

func newLBControlPlane(name, namespace string,
	logger logr.Logger,
	config airshipv1.LoadBalancerServiceControlPlane,
	machines *bmh.MachineList,
	client client.Client) loadBalancerControlPlane {
	servicePorts := []corev1.ServicePort{
		{
			Name:     "http",
			Port:     6443,
			NodePort: int32(config.NodePort),
		},
	}
	return loadBalancerControlPlane{loadBalancer{
		sipName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		logger:       logger,
		config:       config.SIPClusterService,
		machines:     machines,
		client:       client,
		bmhRole:      airshipv1.RoleControlPlane,
		template:     templateControlPlane,
		servicePorts: servicePorts,
	},
		config,
	}
}

func newLBWorker(name, namespace string,
	logger logr.Logger,
	config airshipv1.LoadBalancerServiceWorker,
	machines *bmh.MachineList,
	client client.Client) loadBalancerWorker {
	servicePorts := []corev1.ServicePort{}
	for port := config.NodePortRange.Start; port <= config.NodePortRange.End; port++ {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:     fmt.Sprintf("port-%d", port),
			Port:     int32(port),
			NodePort: int32(port),
		})
	}
	return loadBalancerWorker{loadBalancer{
		sipName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		logger:       logger,
		config:       config.SIPClusterService,
		machines:     machines,
		client:       client,
		bmhRole:      airshipv1.RoleWorker,
		template:     templateWorker,
		servicePorts: servicePorts,
	},
		config,
	}
}

func (lb loadBalancer) Finalize() error {
	// implete to delete loadbalancer
	return nil
}

func (lb loadBalancer) generateTemplate(p proxy) ([]byte, error) {
	tmpl, err := template.New("haproxy-config").Parse(lb.template)
	if err != nil {
		return nil, err
	}

	w := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(w, p); err != nil {
		return nil, err
	}

	rendered := w.Bytes()
	return rendered, nil
}

var templateControlPlane = `global
  log stdout format raw local0 notice
  daemon

defaults
  mode                    http
  log                     global
  option                  httplog
  option                  dontlognull
  retries                 1
  # Configures the timeout for a connection request to be left pending in a queue
  # (connection requests are queued once the maximum number of connections is reached).
  timeout queue           30s
  # Configures the timeout for a connection to a backend server to be established.
  timeout connect         30s
  # Configures the timeout for inactivity during periods when we would expect
  # the client to be speaking. For usability of 'kubectl exec', the timeout should
  # be long enough to cover inactivity due to idleness of interactive sessions.
  timeout client          600s
  # Configures the timeout for inactivity during periods when we would expect
  # the server to be speaking. For usability of 'kubectl log -f', the timeout should
  # be long enough to cover inactivity due to the lack of new logs.
  timeout server          600s

#---------------------------------------------------------------------
{{- $servers := .Servers }}
{{- range .ContainerPorts }}
{{- $containerPort := . }}
frontend {{ $containerPort.Name }}-frontend
  bind *:{{ $containerPort.ContainerPort }}
  mode tcp
  option tcplog
  default_backend {{ $containerPort.Name }}-backend
backend {{ $containerPort.Name }}-backend
  mode tcp
  balance     roundrobin
  option httpchk GET /readyz
  http-check expect status 200
  option log-health-checks
  # Observed apiserver returns 500 for around 10s when 2nd cp node joins.
  # downinter 2s makes it check more frequently to recover from that state sooner.
  # Also changing fall to 4 so that it takes longer (4 failures) for it to take down a backend.
  default-server check check-ssl verify none inter 5s downinter 2s fall 4 on-marked-down shutdown-sessions
{{- range $servers }}
{{- $server := . }}
  server {{ $server.Name }} {{ $server.IP }}:{{ $containerPort.ContainerPort }}
{{ end -}}
{{ end -}}`

var templateWorker = `global
log stdout format raw local0 notice
daemon

defaults
mode                    tcp
log                     global
option                  tcplog
option                  dontlognull
retries                 1
# Configures the timeout for a connection request to be left pending in a queue
# (connection requests are queued once the maximum number of connections is reached).
timeout queue           30s
# Configures the timeout for a connection to a backend server to be established.
timeout connect         30s
# Configures the timeout for inactivity during periods when we would expect
# the client to be speaking.
timeout client          600s
# Configures the timeout for inactivity during periods when we would expect
# the server to be speaking.
timeout server          600s

#---------------------------------------------------------------------
{{- $servers := .Servers }}
{{- range .ContainerPorts }}
{{- $containerPort := . }}
frontend {{ $containerPort.Name }}-frontend
  bind *:{{ $containerPort.ContainerPort }}
  default_backend {{ $containerPort.Name }}-backend
backend {{ $containerPort.Name }}-backend
  balance     roundrobin
  option tcp-check
  tcp-check connect
  option log-health-checks
default-server check
{{- range $servers }}
{{- $server := . }}
  server {{ $server.Name }} {{ $server.IP }}:{{ $containerPort.ContainerPort }}
{{ end -}}
{{ end -}}`
