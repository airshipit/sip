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
