#!/bin/bash

set -e

function cleanup {
  echo "Cleaning up resources..."
#  helm delete --purge iam-role-annotator
#  kubectl delete deployment nginx-deployment --namespace ${NAMESPACE}
#  kubectl delete namespace ${NAMESPACE}
  kind delete cluster
}
trap cleanup EXIT

NAMESPACE="ns-1"

GO111MODULE="on" go get -u sigs.k8s.io/kind@master
#docker inspect 'jet-app-serialsh.e2etest.sh'
#KIND_DOCKER_IP=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' 'jet-app-serialsh.e2etest.sh')
#echo "LA IP DEL CONTAINER ES ${KIND_DOCKER_IP}"
#sed -i "s/0.0.0.0/${KIND_DOCKER_IP}/g" kind-config.yaml
kind create cluster --wait 60s --config kind-config.yaml
#DOCKER_IP=$(echo ${DOCKER_HOST:-tcp://127.0.0.1:2376} | cut -d/ -f3 | cut -d: -f1)
export KUBECONFIG="$(kind get kubeconfig-path)"
#sed -i "s/localhost/127.0.0.1/g" ${KUBECONFIG}
sleep 10
env | grep DOCKER
cat ${KUBECONFIG}
cat kind-config.yaml
docker ps
#netstat -lntp
#docker info
#docker logs kind-control-plane
#docker exec -it kind-control-plane systemctl status kubelet.service
#docker exec -it kind-control-plane kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes
kubectl create namespace ${NAMESPACE}
kubectl create rolebinding default-admin --clusterrole=admin --serviceaccount=${NAMESPACE}:default --namespace=${NAMESPACE}
helm init --wait --tiller-namespace ${NAMESPACE}
helm upgrade --tiller-namespace ${NAMESPACE} --namespace "${NAMESPACE}" --wait --install "iam-role-annotator" "./charts/iam-role-annotator" --set image.tag="latest" --set awsAccountId="12345"

cat <<EOF | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
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

POD_NAME=$(kubectl get pods --namespace ${NAMESPACE} -l "app=nginx" -o jsonpath="{.items[0].metadata.name}")

if [[ $(kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq '.metadata.annotations' | jq 'contains({"iam.amazonaws.com/role"})') == 'true' ]]; then
  echo "SUCCESS!"
  exit 0
else
  echo "ERROR: the POD does not contain the expected annotation"
  kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq
  exit 1
fi
