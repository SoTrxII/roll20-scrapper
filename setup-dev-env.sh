 # Setup a k8s cluster to deploy openfass in
kind create cluster
arkade install openfaas --basic-auth=false
kubectl rollout status -n openfaas deploy/gateway
kubectl port-forward -n openfaas svc/gateway 8080:8080  > /dev/null 2>&1 &
# Start local docker registry
docker run -d -p 5000:5000 --name registry registry:2