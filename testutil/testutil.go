package testutil

import (
	"fmt"

	"github.com/onsi/gomega"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	airshipv1 "sipcluster/pkg/api/v1"
)

var bmhRoleToLabelValue = map[airshipv1.BMHRole]string{
	airshipv1.RoleControlPlane: "control-plane",
	airshipv1.RoleWorker:       "worker",
}

func UnscheduledSelector() labels.Selector {
	sel := labels.NewSelector()
	r, err := labels.NewRequirement(sipClusterLabel, selection.DoesNotExist, nil)
	gomega.Expect(err).Should(gomega.Succeed())
	return sel.Add(*r)
}

// NOTE(aw442m): These constants have been redefined from the bmh package in order to avoid an import cycle.
const (
	sipRackLabel    = "sip.airshipit.org/rack"
	sipClusterLabel = "sip.airshipit.org/cluster"
	sipServerLabel  = "sip.airshipit.org/server"

	bmhLabel = "example.org/bmh-label"

	sshPrivateKeyBase64 = "DUMMY_DATA"

	networkDataContent = `
{
    "links": [
        {
            "id": "eno4",
            "name": "eno4",
            "type": "phy",
            "mtu": 1500
        },
        {
            "id": "enp59s0f1",
            "name": "enp59s0f1",
            "type": "phy",
            "mtu": 9100
        },
        {
            "id": "enp216s0f0",
            "name": "enp216s0f0",
            "type": "phy",
            "mtu": 9100
        },
        {
            "id": "bond0",
            "name": "bond0",
            "type": "bond",
            "bond_links": [
                "enp59s0f1",
                "enp216s0f0"
            ],
            "bond_mode": "802.3ad",
            "bond_xmit_hash_policy": "layer3+4",
            "bond_miimon": 100,
            "mtu": 9100
        },
        {
            "id": "bond0.41",
            "name": "bond0.41",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 41,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.42",
            "name": "bond0.42",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 42,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.44",
            "name": "bond0.44",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 44,
            "mtu": 9100,
            "vlan_mac_address": null
        },
        {
            "id": "bond0.45",
            "name": "bond0.45",
            "type": "vlan",
            "vlan_link": "bond0",
            "vlan_id": 45,
            "mtu": 9100,
            "vlan_mac_address": null
        }
    ],
    "networks": [
        {
            "id": "oam-ipv6",
            "type": "ipv6",
            "link": "bond0.41",
            "ip_address": "2001:1890:1001:293d::139",
            "routes": [
                {
                    "network": "::/0",
                    "netmask": "::/0",
                    "gateway": "2001:1890:1001:293d::1"
                }
            ]
        },
        {
            "id": "oam-ipv4",
            "type": "ipv4",
            "link": "bond0.41",
            "ip_address": "32.68.51.139",
            "netmask": "255.255.255.128",
            "dns_nameservers": [
                "135.188.34.124",
                "135.38.244.16",
                "135.188.34.84"
            ],
            "routes": [
                {
                    "network": "0.0.0.0",
                    "netmask": "0.0.0.0",
                    "gateway": "32.68.51.129"
                }
            ]
        },
        {
            "id": "pxe-ipv6",
            "link": "eno4",
            "type": "ipv6",
            "ip_address": "fd00:900:100:138::11"
        },
        {
            "id": "pxe-ipv4",
            "link": "eno4",
            "type": "ipv4",
            "ip_address": "172.30.0.11",
            "netmask": "255.255.255.128"
        },
        {
            "id": "storage-ipv6",
            "link": "bond0.42",
            "type": "ipv6",
            "ip_address": "fd00:900:100:139::15"
        },
        {
            "id": "storage-ipv4",
            "link": "bond0.42",
            "type": "ipv4",
            "ip_address": "172.31.1.15",
            "netmask": "255.255.255.128"
        },
        {
            "id": "ksn-ipv6",
            "link": "bond0.44",
            "type": "ipv6",
            "ip_address": "fd00:900:100:13a::11"
        },
        {
            "id": "ksn-ipv4",
            "link": "bond0.44",
            "type": "ipv4",
            "ip_address": "172.29.0.11",
            "netmask": "255.255.255.128"
        }
    ]
}`
)

// CreateBMH initializes a BaremetalHost with specific parameters for use in test cases.
func CreateBMH(node int, namespace string, role airshipv1.BMHRole, rack int) (*metal3.BareMetalHost, *corev1.Secret) {
	rackLabel := fmt.Sprintf("r%d", rack)
	networkDataName := fmt.Sprintf("node%d-network-data", node)
	return &metal3.BareMetalHost{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("node0%d", node),
				Namespace: namespace,
				Labels: map[string]string{
					bmhLabel:       bmhRoleToLabelValue[role],
					sipRackLabel:   rackLabel,
					sipServerLabel: fmt.Sprintf("stl2%so%d", rackLabel, node),
				},
			},
			Spec: metal3.BareMetalHostSpec{
				NetworkData: &corev1.SecretReference{
					Namespace: namespace,
					Name:      networkDataName,
				},
				BMC: metal3.BMCDetails{
					Address: "redfish+https://32.68.51.12/redfish/v1/Systems/System.Embedded.1",
				},
			},
		}, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      networkDataName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"networkData": []byte(networkDataContent),
			},
			Type: corev1.SecretTypeOpaque,
		}
}

// CreateSIPCluster initializes a SIPCluster with specific parameters for use in test cases.
func CreateSIPCluster(name string, namespace string, controlPlanes int, workers int) (
	*airshipv1.SIPCluster, *corev1.Secret) {
	sshPrivateKeySecretName := fmt.Sprintf("%s-ssh-private-key", name)
	return &airshipv1.SIPCluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       "SIPCluster",
				APIVersion: "airship.airshipit.org/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: airshipv1.SIPClusterSpec{
				Nodes: map[airshipv1.BMHRole]airshipv1.NodeSet{
					airshipv1.RoleControlPlane: {
						LabelSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								bmhLabel: bmhRoleToLabelValue[airshipv1.RoleControlPlane],
							},
						},
						Scheduling: airshipv1.HostAntiAffinity,
						Count: &airshipv1.NodeCount{
							Active:  controlPlanes,
							Standby: 0,
						},
					},
					airshipv1.RoleWorker: {
						LabelSelector: metav1.LabelSelector{
							MatchLabels: map[string]string{
								bmhLabel: bmhRoleToLabelValue[airshipv1.RoleWorker],
							},
						},
						Scheduling: airshipv1.HostAntiAffinity,
						Count: &airshipv1.NodeCount{
							Active:  workers,
							Standby: 0,
						},
					},
				},
				Services: airshipv1.SIPClusterServices{
					LoadBalancer: []airshipv1.SIPClusterService{
						{
							NodeInterface: "eno3",
							NodePort:      30000,
						},
					},
					JumpHost: []airshipv1.JumpHostService{
						{
							SIPClusterService: airshipv1.SIPClusterService{
								Image:         "quay.io/airshipit/jump-host",
								NodePort:      30001,
								NodeInterface: "eno3",
							},
							SSHAuthorizedKeys: []string{
								"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCyaozS8kZRw2a1d0O4YXhxtJlDPThqIZilGCsXLbukIFOyMUmMTwQAtwWp5epwU1+5ponC2uBENB6xCCj3cl5Rd43d2/B6HxyAPQGKo6/zKYGAKW2nzYDxSWMl6NUSsiJAyXUA7ZlNZQe0m8PmaferlkQyLLZo3NJpizz6U6ZCtxvj43vEl7NYWnLUEIzGP9zMqltIGnD4vYrU9keVKKXSsp+DkApnbrDapeigeGATCammy2xRrUQDuOvGHsfnQbXr2j0onpTIh0PiLrXLQAPDg8UJRgVB+ThX+neI3rQ320djzRABckNeE6e4Kkwzn+QdZsmA2SDvM9IU7boK1jVQlgUPp7zF5q3hbb8Rx7AadyTarBayUkCgNlrMqth+tmTMWttMqCPxJRGnhhvesAHIl55a28Kzz/2Oqa3J9zwzbyDIwlEXho0eAq3YXEPeBhl34k+7gOt/5Zdbh+yacFoxDh0LrshQgboAijcVVaXPeN0LsHEiVvYIzugwIvCkoFMPWoPj/kEGzPY6FCkVneDA7VoLTCoG8dlrN08Lf05/BGC7Wllm66pTNZC/cKXP+cjpQn1iEuiuPxnPldlMHx9sx2y/BRoft6oT/GzqkNy1NTY/xI+MfmxXnF5kwSbcTbzZQ9fZ8xjh/vmpPBgDNrxOEAT4N6OG7GQIhb9HEhXQCQ== example-key", //nolint
								"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCwpOyZjZ4gB0OTvmofH3llh6cBCWaEiEmHZWSkDXr8Bih6HcXVOtYMcFi/ZnUVGUBPw3ATNQBZUaVCYKeF+nDfKTJ9hmnlsyHxV2LeMsVg1o15Pb6f+QJuavEqtE6HI7mHyId4Z1quVTJXDWDW8OZEG7M3VktauqAn/e9UJvlL0bGmTFD1XkNcbRsWMRWkQgt2ozqlgrpPtvrg2/+bNucxX++VUjnsn+fGgAT07kbnrZwppGnAfjbYthxhv7GeSD0+Z0Lf1kiKy/bhUqXsZIuexOfF0YrRyUH1KBl8GCX2OLBYvXHyusByqsrOPiROqRdjX5PsK6HSAS0lk0niTt1p example-key-2",                                                                                                                                                                                                                                                                                                                                                       // nolint
							},
							NodeSSHPrivateKeys: sshPrivateKeySecretName,
						},
					},
				},
			},
			Status: airshipv1.SIPClusterStatus{},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sshPrivateKeySecretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"key": []byte(sshPrivateKeyBase64),
			},
			Type: corev1.SecretTypeOpaque,
		}
}

// CreateBMCAuthSecret creates a K8s Secret that matches the Metal3.io BaremetalHost credential format for use in test
// cases.
func CreateBMCAuthSecret(nodeName string, namespace string, username string, password string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-bmc-credentials", nodeName),
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
	}
}

func CompareLabels(expected labels.Selector, actual map[string]string) error {
	if !expected.Matches(labels.Set(actual)) {
		return fmt.Errorf("labels do not match expected selector %v. Has labels %v", expected, actual)
	}

	return nil
}
