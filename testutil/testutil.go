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
)

// CreateBMH initializes a BaremetalHost with specific parameteres for use in test cases.
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
				"networkData": []byte("ewoKICAgICJsaW5rcyI6IFsKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJlbm80IiwKICAgICAgICAgICAgIm5hbWUiOiAiZW5vNCIsCiAgICAgICAgICAgICJ0eXBlIjogInBoeSIsCiAgICAgICAgICAgICJtdHUiOiAxNTAwCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJlbnA1OXMwZjEiLAogICAgICAgICAgICAibmFtZSI6ICJlbnA1OXMwZjEiLAogICAgICAgICAgICAidHlwZSI6ICJwaHkiLAogICAgICAgICAgICAibXR1IjogOTEwMAogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAiZW5wMjE2czBmMCIsCiAgICAgICAgICAgICJuYW1lIjogImVucDIxNnMwZjAiLAogICAgICAgICAgICAidHlwZSI6ICJwaHkiLAogICAgICAgICAgICAibXR1IjogOTEwMAogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAiYm9uZDAiLAogICAgICAgICAgICAibmFtZSI6ICJib25kMCIsCiAgICAgICAgICAgICJ0eXBlIjogImJvbmQiLAogICAgICAgICAgICAiYm9uZF9saW5rcyI6IFsKICAgICAgICAgICAgICAgICJlbnA1OXMwZjEiLAogICAgICAgICAgICAgICAgImVucDIxNnMwZjAiCiAgICAgICAgICAgIF0sCiAgICAgICAgICAgICJib25kX21vZGUiOiAiODAyLjNhZCIsCiAgICAgICAgICAgICJib25kX3htaXRfaGFzaF9wb2xpY3kiOiAibGF5ZXIzKzQiLAogICAgICAgICAgICAiYm9uZF9taWltb24iOiAxMDAsCiAgICAgICAgICAgICJtdHUiOiA5MTAwCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40MSIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQxIiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQxLAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40MiIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQyIiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQyLAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40NCIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQ0IiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQ0LAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJib25kMC40NSIsCiAgICAgICAgICAgICJuYW1lIjogImJvbmQwLjQ1IiwKICAgICAgICAgICAgInR5cGUiOiAidmxhbiIsCiAgICAgICAgICAgICJ2bGFuX2xpbmsiOiAiYm9uZDAiLAogICAgICAgICAgICAidmxhbl9pZCI6IDQ1LAogICAgICAgICAgICAibXR1IjogOTEwMCwKICAgICAgICAgICAgInZsYW5fbWFjX2FkZHJlc3MiOiBudWxsCiAgICAgICAgfQogICAgXSwKICAgICJuZXR3b3JrcyI6IFsKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJvYW0taXB2NiIsCiAgICAgICAgICAgICJ0eXBlIjogImlwdjYiLAogICAgICAgICAgICAibGluayI6ICJib25kMC40MSIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjIwMDE6MTg5MDoxMDAxOjI5M2Q6OjE0MCIsCiAgICAgICAgICAgICJyb3V0ZXMiOiBbCiAgICAgICAgICAgICAgICB7CiAgICAgICAgICAgICAgICAgICAgIm5ldHdvcmsiOiAiOjovMCIsCiAgICAgICAgICAgICAgICAgICAgIm5ldG1hc2siOiAiOjovMCIsCiAgICAgICAgICAgICAgICAgICAgImdhdGV3YXkiOiAiMjAwMToxODkwOjEwMDE6MjkzZDo6MSIKICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAib2FtLWlwdjQiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY0IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDEiLAogICAgICAgICAgICAiaXBfYWRkcmVzcyI6ICIzMi42OC41MS4xNDAiLAogICAgICAgICAgICAibmV0bWFzayI6ICIyNTUuMjU1LjI1NS4xMjgiLAogICAgICAgICAgICAiZG5zX25hbWVzZXJ2ZXJzIjogWwogICAgICAgICAgICAgICAgIjEzNS4xODguMzQuMTI0IiwKICAgICAgICAgICAgICAgICIxMzUuMzguMjQ0LjE2IiwKICAgICAgICAgICAgICAgICIxMzUuMTg4LjM0Ljg0IgogICAgICAgICAgICBdLAogICAgICAgICAgICAicm91dGVzIjogWwogICAgICAgICAgICAgICAgewogICAgICAgICAgICAgICAgICAgICJuZXR3b3JrIjogIjAuMC4wLjAiLAogICAgICAgICAgICAgICAgICAgICJuZXRtYXNrIjogIjAuMC4wLjAiLAogICAgICAgICAgICAgICAgICAgICJnYXRld2F5IjogIjMyLjY4LjUxLjEyOSIKICAgICAgICAgICAgICAgIH0KICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAicHhlLWlwdjYiLAogICAgICAgICAgICAibGluayI6ICJlbm80IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NiIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogImZkMDA6OTAwOjEwMDoxMzg6OjEyIgogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAicHhlLWlwdjQiLAogICAgICAgICAgICAibGluayI6ICJlbm80IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NCIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjE3Mi4zMC4wLjEyIiwKICAgICAgICAgICAgIm5ldG1hc2siOiAiMjU1LjI1NS4yNTUuMTI4IgogICAgICAgIH0sCiAgICAgICAgewogICAgICAgICAgICAiaWQiOiAic3RvcmFnZS1pcHY2IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDIiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY2IiwKICAgICAgICAgICAgImlwX2FkZHJlc3MiOiAiZmQwMDo5MDA6MTAwOjEzOTo6MTYiCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJzdG9yYWdlLWlwdjQiLAogICAgICAgICAgICAibGluayI6ICJib25kMC40MiIsCiAgICAgICAgICAgICJ0eXBlIjogImlwdjQiLAogICAgICAgICAgICAiaXBfYWRkcmVzcyI6ICIxNzIuMzEuMC4xNiIsCiAgICAgICAgICAgICJuZXRtYXNrIjogIjI1NS4yNTUuMjU1LjEyOCIKICAgICAgICB9LAogICAgICAgIHsKICAgICAgICAgICAgImlkIjogImtzbi1pcHY2IiwKICAgICAgICAgICAgImxpbmsiOiAiYm9uZDAuNDQiLAogICAgICAgICAgICAidHlwZSI6ICJpcHY2IiwKICAgICAgICAgICAgImlwX2FkZHJlc3MiOiAiZmQwMDo5MDA6MTAwOjEzYTo6MTIiCiAgICAgICAgfSwKICAgICAgICB7CiAgICAgICAgICAgICJpZCI6ICJrc24taXB2NCIsCiAgICAgICAgICAgICJsaW5rIjogImJvbmQwLjQ0IiwKICAgICAgICAgICAgInR5cGUiOiAiaXB2NCIsCiAgICAgICAgICAgICJpcF9hZGRyZXNzIjogIjE3Mi4yOS4wLjEyIiwKICAgICAgICAgICAgIm5ldG1hc2siOiAiMjU1LjI1NS4yNTUuMTI4IgogICAgICAgIH0KICAgIF0KfQo="),
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
			Nodes: map[airshipv1.VmRoles]airshipv1.NodeSet{
				airshipv1.VmMaster: airshipv1.NodeSet{
					VmFlavor:   "airshipit.org/vino-flavor=master",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VmCount{
						Active:  masters,
						Standby: 0,
					},
				},
				airshipv1.VmWorker: airshipv1.NodeSet{
					VmFlavor:   "airshipit.org/vino-flavor=worker",
					Scheduling: airshipv1.ServerAntiAffinity,
					Count: &airshipv1.VmCount{
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
