#!/bin/sh

function cleanup {
  echo "Cleaning up resources..."
  helm delete --purge iam-role-annotator
  kubectl delete deployment nginx-deployment --namespace ${NAMESPACE}
}
trap cleanup EXIT

NAMESPACE="default"
helm init > /dev/null
helm upgrade --namespace "${NAMESPACE}" --install "iam-role-annotator" "./charts/iam-role-annotator" --set image.tag="latest" --set awsAccountId="12345" > /dev/null

echo "Controller iam-role-annotator deployed on the cluster."

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

echo "Let's wait some seconds for the controller to detect the annotation in the Pod and re-create it"
sleep 20

POD_NAME=$(kubectl get pods --namespace ${NAMESPACE} --field-selector=status.phase=Running -l "app=nginx" -o jsonpath="{.items[0].metadata.name}")

if [[ $(kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq '.metadata.annotations' | jq 'contains({"iam.amazonaws.com/role"})') == 'true' ]]; then
  echo "SUCCESS!"
  exit 0
else
  echo "ERROR: the POD does not contain the expected annotation"
  kubectl get pod --namespace ${NAMESPACE} ${POD_NAME} -o json | jq
  exit 1
fi
