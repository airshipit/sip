#!/bin/bash
set -xe
sudo snap install kustomize && sudo snap install go --classic
make images
kubectl wait --for=condition=Ready pods --all -A --timeout=180s
make deploy
#Wait for sip controller manager Pod
count=0
until [[ $(kubectl -n sipcluster-system get pod -l control-plane=controller-manager 2>/dev/null) ]]; do
  count=$((count + 1))
  if [[ ${count} -eq "120" ]]; then
    echo ' Timed out waiting for sip controller manager pod to exist' >&3
    return 1
  fi
  sleep 2
done
kubectl wait -n sipcluster-system pod -l control-plane=controller-manager --for=condition=ready --timeout=240s
kubectl get po -A
# Install Cert-Manager
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.1/cert-manager.yaml
kubectl wait --timeout=180s --for=condition=Established crd/clusterissuers.cert-manager.io \
  crd/issuers.cert-manager.io \
  crd/certificaterequests.cert-manager.io \
  crd/certificates.cert-manager.io
kubectl rollout status --timeout=180s -n cert-manager deployment/cert-manager
