#!/bin/bash

set -ex

: ${KUBE_VERSION:="v1.19.2"}
: ${MINIKUBE_VERSION:="v1.16.0"}
: ${UPSTREAM_DNS_SERVER:="8.8.4.4"}
: ${DNS_DOMAIN:="cluster.local"}
: ${CALICO_VERSION:="v3.17"}
: ${CNI_MANIFEST_PATH:="/tmp/calico.yaml"}

export DEBCONF_NONINTERACTIVE_SEEN=true
export DEBIAN_FRONTEND=noninteractive

sudo -E apt-get update

sudo -E apt-get install -y \
  socat \
  jq \
  util-linux \
  nfs-common \
  bridge-utils \
  iptables \
  conntrack \
  libffi-dev

# Prepare tmpfs for etcd
sudo mkdir -p /data
sudo mount -t tmpfs -o size=512m tmpfs /data

# Download calico manifest
if [ ! -f "$CNI_MANIFEST_PATH" ]; then
  curl -Ss https://docs.projectcalico.org/"${CALICO_VERSION}"/manifests/calico.yaml -o ${CNI_MANIFEST_PATH}
fi

# Install minikube and kubectl
URL="https://storage.googleapis.com"
sudo -E curl -sSLo /usr/local/bin/minikube "${URL}"/minikube/releases/"${MINIKUBE_VERSION}"/minikube-linux-amd64
sudo -E curl -sSLo /usr/local/bin/kubectl "${URL}"/kubernetes-release/release/"${KUBE_VERSION}"/bin/linux/amd64/kubectl
sudo -E chmod +x /usr/local/bin/minikube
sudo -E chmod +x /usr/local/bin/kubectl

export CHANGE_MINIKUBE_NONE_USER=true
export MINIKUBE_IN_STYLE=false

sudo -E minikube start \
  --kubernetes-version="${KUBE_VERSION}" \
  --embed-certs=true \
  --interactive=false \
  --driver=none \
  --wait=apiserver,system_pods,node_ready \
  --wait-timeout=6m0s \
  --network-plugin=cni \
  --cni=${CNI_MANIFEST_PATH} \
  --extra-config=kube-proxy.mode=ipvs \
  --extra-config=controller-manager.allocate-node-cidrs=true \
  --extra-config=controller-manager.cluster-cidr=192.168.0.0/16 \
  --extra-config=kubeadm.pod-network-cidr=192.168.0.0/16 \
  --extra-config=kubelet.resolv-conf=/run/systemd/resolve/resolv.conf

kubectl get nodes -o wide
kubectl get pod -A

cat <<EOF | kubectl replace -f -
apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes ${DNS_DOMAIN} in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . ${UPSTREAM_DNS_SERVER} {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
EOF

kubectl wait --for=condition=Ready pods --all -A --timeout=180s