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
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"
)

// InfraService generalizes inftracture services
type InfraService interface {
	Deploy() error
	Finalize() error
	Type() airshipv1.InfraService
}

// ServiceSet provides access to infrastructure services
type ServiceSet struct {
	logger   logr.Logger
	sip      airshipv1.SIPCluster
	machines *airshipvms.MachineList
	client   client.Client

	services map[airshipv1.InfraService]InfraService
}

// NewServiceSet returns new instance of ServiceSet
func NewServiceSet(
	logger logr.Logger,
	sip airshipv1.SIPCluster,
	machines *airshipvms.MachineList,
	client client.Client) ServiceSet {
	logger = logger.WithValues("SIPCluster", types.NamespacedName{Name: sip.GetNamespace(), Namespace: sip.GetName()})

	return ServiceSet{
		logger:   logger,
		sip:      sip,
		client:   client,
		machines: machines,
	}
}

// LoadBalancer returns loadbalancer service
func (ss ServiceSet) LoadBalancer() (InfraService, error) {
	lb, ok := ss.services[airshipv1.LoadBalancerService]
	if !ok {
		ss.logger.Info("sip cluster doesn't have loadbalancer infrastructure service defined")
	}
	return lb, fmt.Errorf("loadbalancer service is not defined for sip cluster '%s'/'%s'",
		ss.sip.GetNamespace(),
		ss.sip.GetName())
}

func (ss ServiceSet) Finalize() error {
	serviceNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ss.sip.Spec.ClusterName,
		},
	}
	return ss.client.Delete(context.TODO(), serviceNamespace)
}

func CreateNS(serviceNamespaceName string, c client.Client) error {
	ns := &corev1.Namespace{}
	key := client.ObjectKey{Name: serviceNamespaceName}
	if err := c.Get(context.Background(), key, ns); err == nil {
		// Namespace already exists
		return nil
	}

	serviceNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceNamespaceName,
		},
	}
	return c.Create(context.TODO(), serviceNamespace)
}

// ServiceList returns all services defined in Set
func (ss ServiceSet) ServiceList() []InfraService {
	var serviceList []InfraService
	for _, serviceConfig := range ss.sip.Spec.InfraServices {
		switch serviceConfig.ServiceType {
		case airshipv1.LoadBalancerService:
			serviceList = append(serviceList,
				newLB(ss.sip.GetName(),
					ss.sip.Spec.ClusterName,
					ss.logger,
					serviceConfig,
					ss.machines,
					ss.client))
		case airshipv1.JumpHostService:
			serviceList = append(serviceList,
				newJumpHost(ss.sip.GetName(),
					ss.sip.Spec.ClusterName,
					ss.logger,
					serviceConfig,
					ss.machines,
					ss.client))
		default:
			ss.logger.Info("serviceType unsupported", "serviceType", serviceConfig.ServiceType)
		}
	}
	return serviceList
}

func applyRuntimeObject(key client.ObjectKey, obj client.Object, c client.Client) error {
	ctx := context.Background()
	switch err := c.Get(ctx, key, obj); {
	case apierror.IsNotFound(err):
		return c.Create(ctx, obj)
	case err == nil:
		return c.Update(ctx, obj)
	default:
		return err
	}
}
