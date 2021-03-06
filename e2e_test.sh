#!/bin/bash

set -e

function cleanup {
    if [[ "${CREATE_CLUSTER}" != true ]]; then
        kubectl delete namespace ${NAMESPACE}
    fi
}
trap cleanup EXIT

CREATE_CLUSTER=${CREATE_CLUSTER:-true}
NAMESPACE=${NAMESPACE:-ns-1}
DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-nginx-deployment}
AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID:-12345}
K8S_VERSION=${K8S_VERSION:-v1.13.6}

if [[ "${CREATE_CLUSTER}" == true ]]; then
    sed -i -E "s/(kubernetesVersion:\s*)(.+)/\1 ${K8S_VERSION}/" kind-config.yaml
    kind create cluster --wait 60s --config kind-config.yaml
    export KUBECONFIG="$(kind get kubeconfig-path)"
fi

# Create namespace for test
kubectl create namespace ${NAMESPACE}

# Install Helm
kubectl create rolebinding default-admin --clusterrole=admin --serviceaccount=${NAMESPACE}:default --namespace=${NAMESPACE}
helm init --wait --tiller-namespace ${NAMESPACE}

#  Install controller
helm upgrade --tiller-namespace ${NAMESPACE} --namespace "${NAMESPACE}" --wait --install "iam-role-annotator" "./charts/iam-role-annotator" --set image.tag="${TRAVIS_COMMIT:-latest}" --set awsAccountId="${AWS_ACCOUNT_ID}"

# Create Deployment that needs annotation
cat <<EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${DEPLOYMENT_NAME}
  namespace: ${NAMESPACE}
  labels:
    app: nginx
  annotations:
    armesto.net/iam-role-annotator: "true"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
      annotations:
        prometheus.io/scheme: http
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
EOF

sleep 20

# Test if deployment has annotation
kubectl get pods --namespace ${NAMESPACE}
kubectl logs --namespace ${NAMESPACE} deploy/iam-role-annotator iam-role-annotator
POD_NAME=$(kubectl get pods --namespace ${NAMESPACE} --field-selector=status.phase=Running -l "app=nginx" -o jsonpath="{.items[0].metadata.name}")

if [[ $(kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq '.metadata.annotations' | jq 'contains({"iam.amazonaws.com/role"})') == 'true' ]]; then
  if [[ $(kubectl get pods --namespace ${NAMESPACE} ${POD_NAME} -o json | jq -r '.metadata.annotations."iam.amazonaws.com/role"') == "arn:aws:iam::${AWS_ACCOUNT_ID}:role/${DEPLOYMENT_NAME}" ]]; then
    echo "SUCCESS!"
    exit 0
  else
    echo "ERROR: the annotation contains the wrong value"
    kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq '.'
    exit 1
  fi
else
  echo "ERROR: the POD does not contain the expected annotation"
  kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq '.'
  exit 1
fi
