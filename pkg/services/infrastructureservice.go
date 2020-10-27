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
	helm "github.com/fluxcd/helm-controller/api/v2beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"
)

// Infrastructure interface should be implemented by each Tenant Required
// Infrastructure Service

// Init   : prepares the Service
// Deploy : deploys the service
// Validate : will make sure that the deployment is successfull
type InfrastructureService interface {
	//
	Deploy(*airshipvms.MachineList, airshipvms.MachineData, client.Client) error
	Validate() error
}

// Generic Service Factory
type Service struct {
	serviceName airshipv1.InfraService
	config      airshipv1.InfraConfig
}

func (s *Service) Deploy(machines *airshipvms.MachineList, machineData airshipvms.MachineData, c client.Client) error {
	// do something, might decouple this a bit
	// If the  serviucces are defined as Helm Chart , then deploy might be simply

	// Take the data from teh appropriate Machines
	// Prepare the Config

	// Prepare the HelmReleasecd
	helmRelease := &helm.HelmRelease{}
	if err := c.Create(context.TODO(), helmRelease); err != nil {
		return err
	}
	return nil
}

func (s *Service) Validate() error {
	// do something, might decouple this a bit
	return nil

}

// Service Factory
func NewService(infraName airshipv1.InfraService, infraCfg airshipv1.InfraConfig) (InfrastructureService, error) {
	if infraName == airshipv1.LoadBalancerService {
		return newLoadBalancer(infraCfg), nil
	} else if infraName == airshipv1.JumpHostService {
		return newJumpHost(infraCfg), nil
	} else if infraName == airshipv1.AuthHostService {
		return newAuthHost(infraCfg), nil
	}
	return nil, ErrInfraServiceNotSupported{}
}
