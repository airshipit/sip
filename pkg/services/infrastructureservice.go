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
	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Infrastructure interface should be implemented by each Tenant Required
// Infrastructure Service

// Init   : prepares the Service
// Deploy : deploys the service
// Validate : will make sure that the deployment is successful
type InfrastructureService interface {
	//
	Deploy(airshipv1.SIPCluster, *airshipvms.MachineList, client.Client) error
	Validate() error
	Finalize(airshipv1.SIPCluster, client.Client) error
}

// Generic Service Factory
type Service struct {
	serviceName airshipv1.InfraService
	config      airshipv1.InfraConfig
}

func (s *Service) Deploy(sip airshipv1.SIPCluster, machines *airshipvms.MachineList, c client.Client) error {
	// do something, might decouple this a bit
	// If the  serviucces are defined as Helm Chart , then deploy might be simply

	// Lets make sure that the namespace is in place.
	// will be called the name of the cluster.
	if err := s.createNS(sip.Spec.Config.ClusterName, c); err != nil {
		return err
	}
	// Take the data from the appropriate Machines
	// Prepare the Config
	fmt.Printf("Deploy Service:%v \n", s.serviceName)
	return nil
}

func (s *Service) createNS(serviceNamespaceName string, c client.Client) error {
	// Get Namespace
	// If not foundn then ccreate it
	ns := &corev1.Namespace{}
	// c is a created client.
	err := c.Get(context.Background(), client.ObjectKey{
		Name: serviceNamespaceName,
	}, ns)

	if err != nil {
		serviceNamespace := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceNamespaceName,
			},
		}
		if err := c.Create(context.TODO(), serviceNamespace); err != nil {
			return err
		}
	}

	return nil
}
func (s *Service) Validate() error {
	// do something, might decouple this a bit
	fmt.Printf("Validate Service:%v \n", s.serviceName)

	return nil
}

func (s *Service) Finalize(sip airshipv1.SIPCluster, c client.Client) error {
	return nil
}

func FinalizeCommon(sip airshipv1.SIPCluster, c client.Client) error {
	serviceNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: sip.Spec.Config.ClusterName,
		},
	}
	if err := c.Delete(context.TODO(), serviceNamespace); err != nil {
		return err
	}

	return nil
}

// Service Factory
func NewService(infraName airshipv1.InfraService, infraCfg airshipv1.InfraConfig) (InfrastructureService, error) {
	switch infraName {
	case airshipv1.LoadBalancerService:
		return newLoadBalancer(infraCfg), nil
	case airshipv1.JumpHostService:
		return newJumpHost(infraCfg), nil
	case airshipv1.AuthHostService:
		return newAuthHost(infraCfg), nil
	}
	return nil, ErrInfraServiceNotSupported{}
}
