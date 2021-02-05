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

	"html/template"
	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ConfigSecretName name of the haproxy config secret name/volume/mount
	/* #nosec */
	ConfigSecretName = "haproxy-config"
	// DefaultBalancerImage is the image that will be used as load balancer
	DefaultBalancerImage = "haproxy:2.3.2"
)

func (lb loadBalancer) Deploy() error {
	if lb.config.Image == "" {
		lb.config.Image = DefaultBalancerImage
	}
	if lb.config.NodePort < 30000 || lb.config.NodePort > 32767 {
		lb.logger.Info("Either NodePort is not defined in the CR or NodePort is not in the required range of 30000-32767")
		return nil
	}

	pod, secret, err := lb.generatePodAndSecret()
	if err != nil {
		return err
	}

	lb.logger.Info("Applying loadbalancer secret", "secret", secret.GetNamespace()+"/"+secret.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: secret.GetName(), Namespace: secret.GetNamespace()}, secret, lb.client)
	if err != nil {
		return err
	}

	lb.logger.Info("Applying loadbalancer pod", "pod", pod.GetNamespace()+"/"+pod.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: pod.GetName(), Namespace: pod.GetNamespace()}, pod, lb.client)
	if err != nil {
		return err
	}

	lbService := lb.generateService()
	lb.logger.Info("Applying loadbalancer service", "service", lbService.GetNamespace()+"/"+lbService.GetName())
	err = applyRuntimeObject(client.ObjectKey{Name: lbService.GetName(), Namespace: lbService.GetNamespace()},
		lbService, lb.client)
	if err != nil {
		return err
	}
	return nil
}

func (lb loadBalancer) generatePodAndSecret() (*corev1.Pod, *corev1.Secret, error) {
	secret, err := lb.generateSecret()
	if err != nil {
		return nil, nil, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lb.sipName.Name + "-load-balancer",
			Namespace: lb.sipName.Namespace,
			Labels:    map[string]string{"lb-name": lb.sipName.Namespace + "-haproxy"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "balancer",
					Image: lb.config.Image,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 6443,
						},
					},
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
	}
	return pod, secret, nil
}

func (lb loadBalancer) generateSecret() (*corev1.Secret, error) {
	p := proxy{
		FrontPort: 6443,
		Backends:  make([]backend, 0),
	}
	for _, machine := range lb.machines.Machines {
		if machine.VMRole == airshipv1.VMMaster {
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
			p.Backends = append(p.Backends, backend{IP: ip, Name: machine.BMH.Name, Port: 6443})
		}
	}
	secretData, err := generateTemplate(p)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lb.sipName.Name + "-load-balancer",
			Namespace: lb.sipName.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"haproxy.cfg": secretData,
		},
	}, nil
}

func (lb loadBalancer) generateService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lb.sipName.Name + "-load-balancer-service",
			Namespace: lb.sipName.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     6443,
					NodePort: int32(lb.config.NodePort),
				},
			},
			Selector: map[string]string{"lb-name": lb.sipName.Namespace + "-haproxy"},
			Type:     corev1.ServiceTypeNodePort,
		},
	}
}

type proxy struct {
	FrontPort int
	Backends  []backend
}

type backend struct {
	IP   string
	Name string
	Port int
}

type loadBalancer struct {
	client   client.Client
	sipName  types.NamespacedName
	logger   logr.Logger
	config   airshipv1.SIPClusterService
	machines *airshipvms.MachineList
}

func newLB(name, namespace string,
	logger logr.Logger,
	config airshipv1.SIPClusterService,
	machines *airshipvms.MachineList,
	client client.Client) loadBalancer {
	return loadBalancer{
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

func (lb loadBalancer) Finalize() error {
	// implete to delete loadbalancer
	return nil
}

func generateTemplate(p proxy) ([]byte, error) {
	tmpl, err := template.New("haproxy-config").Parse(defaultTemplate)
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

var defaultTemplate = `global
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
# apiserver frontend which proxys to the masters
#---------------------------------------------------------------------
frontend apiserver
  bind *:{{ .FrontPort }}
  mode tcp
  option tcplog
  default_backend kube-apiservers

#---------------------------------------------------------------------
# round robin balancing for apiserver
#---------------------------------------------------------------------
backend kube-apiservers
  mode tcp
  balance     roundrobin
  option httpchk GET /readyz
  http-check expect status 200
  option log-health-checks
  # Observed apiserver returns 500 for around 10s when 2nd cp node joins.
  # downinter 2s makes it check more frequently to recover from that state sooner.
  # Also changing fall to 4 so that it takes longer (4 failures) for it to take down a backend.
  default-server check check-ssl verify none inter 5s downinter 2s fall 4 on-marked-down shutdown-sessions
    {{- range .Backends }}
    {{- $backEnd := . }}
    server {{ $backEnd.Name }} {{ $backEnd.IP }}:{{ $backEnd.Port }}
    {{ end -}}`
