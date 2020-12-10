package testutil

import (
	"fmt"

	metal3 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	airshipv1 "sipcluster/pkg/api/v1"
)

// NOTE(aw442m): These constants have been redefined from the vbmh package in order to avoid an import cycle.
const (
	sipRackLabel     = "sip.airshipit.org/rack"
	sipScheduleLabel = "sip.airshipit.org/sip-scheduled"
	sipServerLabel   = "sip.airshipit.org/server"

	VinoFlavorLabel = "airshipit.org/vino-flavor"

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
func CreateBMH(node int, namespace string, role string, rack int) (*metal3.BareMetalHost, *corev1.Secret) {
	rackLabel := fmt.Sprintf("r%d", rack)
	networkDataName := fmt.Sprintf("node%d-network-data", node)
	return &metal3.BareMetalHost{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("node0%d", node),
				Namespace: namespace,
				Labels: map[string]string{
					"airshipit.org/vino-flavor": role,
					sipScheduleLabel:            "false",
					sipRackLabel:                rackLabel,
					sipServerLabel:              fmt.Sprintf("stl2%so%d", rackLabel, node),
				},
			},
			Spec: metal3.BareMetalHostSpec{
				NetworkData: &corev1.SecretReference{
					Namespace: namespace,
					Name:      networkDataName,
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
func CreateSIPCluster(name string, namespace string, masters int, workers int) *airshipv1.SIPCluster {
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
			Config: &airshipv1.SipConfig{
				ClusterName: name,
			},
			Nodes: map[airshipv1.VMRoles]airshipv1.NodeSet{
				airshipv1.VMMaster: {
					VMFlavor:   "airshipit.org/vino-flavor=master",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VMCount{
						Active:  masters,
						Standby: 0,
					},
				},
				airshipv1.VMWorker: {
					VMFlavor:   "airshipit.org/vino-flavor=worker",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VMCount{
						Active:  workers,
						Standby: 0,
					},
				},
			},
			InfraServices: map[airshipv1.InfraService]airshipv1.InfraConfig{},
		},
		Status: airshipv1.SIPClusterStatus{},
	}
}
