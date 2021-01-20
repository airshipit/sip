#!/bin/bash
set -xe
sudo snap install kustomize && sudo snap install go --classic
make docker-build
kubectl wait --for=condition=Ready pods --all -A --timeout=180s
make deploy
#Wait for sip controller manager Pod.
kubectl wait -n sipcluster-system pod -l control-plane=controller-manager --for=condition=ready --timeout=240s
kubectl get po -A