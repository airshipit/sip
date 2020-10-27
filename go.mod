module sipcluster

go 1.13

require (
	github.com/fluxcd/helm-controller/api v0.1.3
	github.com/go-logr/logr v0.2.1
	github.com/metal3-io/baremetal-operator v0.0.0-20201014161845-a6d4f1fc3228
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/prometheus/common v0.10.0
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)
