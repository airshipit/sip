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

	"github.com/go-logr/logr"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	airshipv1 "sipcluster/pkg/api/v1"
	airshipvms "sipcluster/pkg/vbmh"
)

// InfraService generalizes inftracture services
type InfraService interface {
	Deploy() error
	Finalize() error
}

// ServiceSet provides access to infrastructure services
type ServiceSet struct {
	logger   logr.Logger
	sip      airshipv1.SIPCluster
	machines *airshipvms.MachineList
	client   client.Client
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

func (ss ServiceSet) Finalize() error {
	return nil
}

// ServiceList returns all services defined in Set
func (ss ServiceSet) ServiceList() ([]InfraService, error) {
	serviceList := []InfraService{}
	services := ss.sip.Spec.Services
	for _, svc := range services.LoadBalancer {
		serviceList = append(serviceList,
			newLB(ss.sip.GetName(),
				ss.sip.GetNamespace(),
				ss.logger,
				svc,
				ss.machines,
				ss.client))
	}
	for _, svc := range services.Auth {
		return nil, ErrInfraServiceNotSupported{svc}
	}
	for _, svc := range services.JumpHost {
		serviceList = append(serviceList,
			newJumpHost(ss.sip.GetName(),
				ss.sip.GetNamespace(),
				ss.logger,
				svc,
				ss.machines,
				ss.client))
	}
	return serviceList, nil
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

func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }
