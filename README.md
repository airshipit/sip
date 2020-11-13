# SIP Cluster Operator

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
    - If master 
        -  collect into list of bmh's to label
    - If worker
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

### Kind kubernetes cluster
Fastest way to set up a k8s cluster for development env is to use kind to set it up

#### Install kind on linux (amd64 arch)

```
# curl -Lo kind https://kind.sigs.k8s.io/dl/v0.9.0/kind-linux-amd64
# sudo install  -m 755 --owner=root --group=root kind /usr/local/bin
# rm kind
```

More information on how to install kind binary can be found be found [here](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

#### Create k8s cluster with kind

```
# make kind-create
# kubectl get nodes
```

### Deploy sip operator on top of kind cluster
kind-load-image target will build docker image from the current state of your local
git repository and upload it to kind cluster to be available for kubelet.

```
# make kind-load-image
# make deploy
```

Now you have a working k8s cluster with sip installed on it with your changes to SIP operator

### Deliver sip CRs to kubernetes

Use kubectl apply to deliver SIP CRs and BaremetalHost CRDs to kubernetes cluster

```
# kubectl apply -f config/samples/airship_v1beta1_sipcluster.yaml
# kubectl apply -f config/samples/bmh/baremetalhosts.metal3.io.yaml
```
Now you are ready to craft and add BaremetalHost CRs into cluster, check samples directory
to find BaremetalHost examples there.
