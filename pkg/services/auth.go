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
	bmh "sipcluster/pkg/bmh"

	v1alpha3 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DexServiceName = "dex-aio"
)

// Auth uses Dex as an Open ID Connect (OIDC) Authentication service.
type auth struct {
	client   client.Client
	sipName  types.NamespacedName
	logger   logr.Logger
	config   airshipv1.AuthService
	machines *bmh.MachineList
}

func newAuth(name, namespace string,
	logger logr.Logger,
	config airshipv1.AuthService,
	machines *bmh.MachineList,
	client client.Client) InfraService {
	return auth{
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

// Deploy creates an Auth service.
func (au auth) Deploy() error {
	clusterIssuerName := DexServiceName + "-" + au.sipName.Namespace + "-" + au.sipName.Name

	clusterIssuer := au.generateCI(clusterIssuerName)
	au.logger.Info("Applying ClusterIssuer", "ClusterIssuer", clusterIssuer.GetName())
	err := applyRuntimeObject(client.ObjectKey{Name: clusterIssuer.GetName()},
		clusterIssuer, au.client)
	if err != nil {
		return err
	}

	return nil
}

// generateCI generates a ClusterIssuer object for Auth Service.
func (au auth) generateCI(clusterIssuerName string) *v1alpha3.ClusterIssuer {
	return &v1alpha3.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha3.SchemeGroupVersion.String(),
			Kind:       "ClusterIssuer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterIssuerName,
		},
		Spec: v1alpha3.IssuerSpec{
			IssuerConfig: v1alpha3.IssuerConfig{
				CA: &v1alpha3.CAIssuer{SecretName: au.config.CaSecret},
			},
		},
		Status: v1alpha3.IssuerStatus{},
	}
}

// Finalize to remove the deployed auth service.
func (au auth) Finalize() error {
	// TODO: Add logic to cleanup auth service.
	return nil
}
