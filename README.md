# SIP Cluster Operator

[![Docker Repository on Quay](https://quay.io/repository/airshipit/sip/status "Docker Repository on Quay")](https://quay.io/repository/airshipit/sip)

## Overview

The lifecycle of the VM's and their relationship to Cluster will be managed using two operators: vNode-Operator(ViNO) and the Support Infra Provider Operator (SIP) .


## Description

The SIP Cluster Operator helps identity appropriate `BareMetalHost` objects to fulfill a tenant cluster, including initial creation as well as expanding and contracting it over time.  It also helps create supporting *per-cluster* supporting infrastructure such as LoadBalancers, Jump Hosts, and so on as value added cluster services for each cluster.

While ViNO is responsible for setting up VM infrastructure, such as:

- per-node vino pod:
    * libvirt init, e.g.
        * setup vm-infra bridge
        * provisioning tftp/dhcp definition
    * libvirt launch
    * sushi pod
- libvirt domains
- networking
- bmh objects, with labels:
    * location - i.e. `rack: 8` and `node: rdm8r008c002` - should follow k8s semi-standard
    * vm role - i.e. `node-type: worker`
    * vm flavor - i.e `node-flavor: foobar`
    * networks - i.e. `networks: [foo, bar]`
and the details for ViNO can be found [here](https://hackmd.io/KSu8p4QeTc2kXIjlrso2eA)

The Cluster Support Infrastructure Provider, or SIP, is responsible for the lifecycle of:
- identifying the correct `BareMetalHost` resources to label (or unlabel) based on scheduling constraints.
- extract IP address information from `BareMetalHost` objects to use in the creation of supporting infrastructure.
- creating support infra for the tenant k8s cluster:
    * load balancers (for tenant k8s api)
    * jump pod to access the cluster and nodes via ssh
    * an OIDC provider for the tenant cluster, i.e. Dex
    * potentially more in the future

## SIP Operator High level Algorithm

::::info
The expectation is that the operator will only deal with one `SIPCluster` object at a time -- in other words serially. There will be absolutely no concurrency support. This is critical to avoid race conditions. There is an expectation that all of the operations below are idempotent.
::::

Pseudo Algorithm at a high level after reading the `SIPCluster` CR:

### Gather Phase

#### Identity BMH VM's
- Gather BMH's that meet the criteria expected for the groups
- Check for existing labeled BMH's
- Complete the expected scheduling contraints :
    - If ControlPlane
        -  collect into list of bmh's to label
    - If Worker
        - collect into list of bmh's to label
#### Extract Info from Identified BMH
-  identify and extract  the IP address ands other info as needed (***)
    -  Use it as part of the service infrastucture configuration
- At this point I have a list of BMH's, and I have the extrapolated data I need for configuring services.

### Service Infrastructure Deploy Phase
- Create or Updated the [LB|admin pod] with the appropriate configuration

### Label Phase
- Label the collected hosts.
- At this point SIPCluster is done processing a given CR, and can move on the next.


SIPCluster CR will exists within the Control phase for a Tenant cluster.

## Development environment

### Pre-requisites

#### Install Golang 1.15+

SIP is a project written in Go, and the make targets used to deploy SIP leverage both Go and
Kustomize commands which require Golang be installed.

For detailed installation instructions, please see the [Golang installation guide](https://golang.org/doc/install).

#### Install Kustomize v3.2.3+

In order to apply manifests to your cluster via Make targets we suggest the use of Kustomize.

For detailed installation instructions, please see the [Kustomize installation guide](https://kubectl.docs.kubernetes.io/installation/kustomize/).

#### Proxy Setup

If your organization requires development behind a proxy server, you will need to define the
following environment variables with your organization's information:

```
HTTP_PROXY=http://username:password@host:port
HTTPS_PROXY=http://username:password@host:port
NO_PROXY="localhost,127.0.0.1,10.96.0.0/12"
PROXY=http://username:password@host:port
USE_PROXY=true
```

10.96.0.0/12 is the Kubernetes service CIDR.

#### Deploy kubernetes using minikube and create k8s cluster

```
# ./tools/deployment/install-k8s.sh
```

### Deploy SIP

```
# make docker-build
# kubectl get nodes
# make deploy
```

By now, you should have a working cluster with ViNO deployed on top of it.

```
kubectl get pods -A
NAMESPACE           NAME                                            READY   STATUS    RESTARTS   AGE
kube-system         calico-kube-controllers-744cfdf676-428vp        1/1     Running   0          4h30m
kube-system         calico-node-pgr4k                               1/1     Running   0          4h30m
kube-system         coredns-f9fd979d6-qk2dc                         1/1     Running   0          4h30m
kube-system         etcd-govino                                     1/1     Running   0          4h30m
kube-system         kube-apiserver-govino                           1/1     Running   0          4h30m
kube-system         kube-controller-manager-govino                  1/1     Running   0          4h30m
kube-system         kube-proxy-6wx46                                1/1     Running   0          4h30m
kube-system         kube-scheduler-govino                           1/1     Running   0          4h30m
kube-system         storage-provisioner                             1/1     Running   0          4h30m
sipcluster-system   sipcluster-controller-manager-59c7dddcb-65lcb   2/2     Running   0          3h47m
```



### Deliver SIP CRs to kubernetes

Now you are ready to craft and add BaremetalHost CRs into cluster, check samples directory
to find BaremetalHost examples there.

Use kubectl apply to deliver SIP CRs and BaremetalHost CRDs to kubernetes cluster

```
# kustomize build config/samples | kubectl apply -f -
```

## Testing

Need kubebuilder installed to run tests.

#### Installation of kubebuilder:

```
# os=$(go env GOOS)
# arch=$(go env GOARCH)

download kubebuilder and extract it to /tmp
# curl -L https://go.kubebuilder.io/dl/2.3.1/${os}/${arch} | tar -xz -C /tmp/

move to a long-term location and put it on your path
(you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
# sudo mv /tmp/kubebuilder_2.3.1_${os}_${arch} /usr/local/kubebuilder
# export PATH=$PATH:/usr/local/kubebuilder/bin
```
Run the tests:
Run `make test` to execute a suite of unit and integration tests against the SIP
operator.
