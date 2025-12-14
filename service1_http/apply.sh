kubectl apply -f k8s/namespace.yaml
kubectl -n service1-lab apply -f k8s/deployment.yaml
kubectl -n service1-lab apply -f k8s/service.yaml