# Minikube Local Deployment Quick Start Guide for go-queue

This guide provides step-by-step instructions to quickly deploy the go-queue system locally using Minikube.

## Deployment Summary

- Single-node Minikube cluster
- API: 2 replicas (stateless, auto-scaled by Kubernetes)
- Worker: 2 replicas (stateless, auto-scaled by Kubernetes)
- Redis: Single instance (no replication or Sentinel HA)
- Suitable for learning Kubernetes basics, simple scaling, and validating manifest configurations before cloud deployment

## Prerequisites

- [Minikube](https://minikube.sigs.k8s.io/docs/start/) installed
- [kubectl](https://kubernetes.io/docs/tasks/tools/) CLI installed
- Docker installed
- go-queue project cloned locally

## 1. Start Minikube Cluster

```bash
# Start minikube with appropriate resources
minikube start --cpus=2 --memory=4g

# Enable the ingress addon (optional, but recommended)
minikube addons enable ingress

# Verify cluster is running
kubectl get nodes
```

## 2. Build Docker Images

Build the API and Worker Docker images:

```bash
# If you're in the docs directory, first navigate back to the project root
cd ..

# Now build the Docker images
docker build -t go-queue-api:latest -f ./docker/Dockerfile.api .
docker build -t go-queue-worker:latest -f ./docker/Dockerfile.worker .

```

> **Note**: If you're running these commands using the green "Run" button in your IDE or documentation viewer, make sure to first navigate to the project root directory. Alternatively, you can use the commands with absolute paths shown above.

Note: We don't need to build a Redis image as the deployment will automatically pull the official Redis image (`redis:6.2-alpine`) from Docker Hub.

## 3. Load Docker Images into Minikube

```bash
# Load API image into Minikube
minikube image load go-queue-api:latest

# Load Worker image into Minikube  
minikube image load go-queue-worker:latest
```

## 4. Deploy to Kubernetes

Apply the Kubernetes resources:

```bash
kubectl apply -k k8s/base
```

Verify the deployment:

```bash
kubectl get all -n go-queue
```

## 5. Access the API

### Option A: Using Port Forward (Recommended for Quick Testing)

```bash
kubectl port-forward -n go-queue service/api 8080:8080
```

Then access the API at http://localhost:8080

### Option B: Using Ingress (if configured)

Add an entry to your `/etc/hosts` file:

```
127.0.0.1 go-queue.local
```

Get the Minikube IP:

```bash
minikube ip
```

Then access the API at http://go-queue.local

## 6. Testing the Deployment

Submit a job to the API:

```bash
curl -X POST http://localhost:8080/jobs -H "Content-Type: application/json" -d '{"payload": "test_job"}'
```

## 7. Troubleshooting

### Redis Connection Issues

If your services cannot connect to Redis:

1. Check Redis logs:
   ```bash
   kubectl logs -n go-queue deployment/redis
   ```

2. Test connectivity from API/Worker pods:
   ```bash
   kubectl exec -it -n go-queue <pod-name> -- sh
   # Inside the pod
   ping redis.go-queue.svc.cluster.local
   ```

3. Verify API/Worker logs:
   ```bash
   kubectl logs -n go-queue deployment/api
   kubectl logs -n go-queue deployment/go-queue-worker
   ```

### Image Pull Issues

If pods are stuck in "ImagePullBackOff" state, ensure images are properly loaded:

```bash
minikube image load go-queue-api:latest
minikube image load go-queue-worker:latest
```

## 8. Cleaning Up Resources

When you're done:

```bash
# Delete everything in the namespace
kubectl delete namespace go-queue

# Or delete using Kustomize
kubectl delete -k k8s/base

# Stop Minikube (optional)
minikube stop

# Delete Minikube cluster (optional)
minikube delete
```
