# Kubernetes Deployment Guide (Phase 5) - Single Instance Deployment

This guide provides step-by-step instructions to deploy a single instance of each component (API, Worker, and Redis) of the job queue system to Kubernetes in Phase 5. Scaling to multiple instances will be addressed in Phase 7.

## Prerequisites

- Docker images for API and Worker services built (Phase 4 completed)
- Minikube or kind installed locally
- kubectl CLI installed
- Basic understanding of Kubernetes concepts

## 1. Set Up Local Kubernetes Cluster

### Option A: Using Minikube

```bash
# Start minikube with appropriate resources
minikube start --cpus=2 --memory=4g

# Enable the ingress addon (optional, but useful)
minikube addons enable ingress

# Verify cluster is running
kubectl get nodes
```

### Option B: Using kind (Kubernetes IN Docker)

```bash
# Create a kind cluster
kind create cluster --name go-queue

# Verify cluster is running
kubectl get nodes
```

## 2. Create Kubernetes Resource Manifests

Create a `k8s` directory in the project root to store all Kubernetes manifests:

```bash
mkdir -p k8s/base
```

### 2.1 Create Namespace

Create `k8s/base/namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: go-queue
```

### 2.2 Create ConfigMap for Application Configuration

Create `k8s/base/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: go-queue-config
  namespace: go-queue
data:
  API_PORT: "8080"
  WORKER_CONCURRENCY: "5"
  REDIS_DB: "0"
  REDIS_QUEUE_KEY: "job_queue"
  DLQ_QUEUE_KEY: "job_queue_dlq"
  RETRY_MAX_ATTEMPTS: "3"
  RETRY_MIN_BACKOFF_MS: "100"
  RETRY_MAX_BACKOFF_MS: "1000"
  LOG_LEVEL: "info"
```

### 2.3 Create Secret for Sensitive Information

Create `k8s/base/secrets.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: go-queue-secrets
  namespace: go-queue
type: Opaque
data:
  # Base64 encoded values
  # You can generate with: echo -n "your-value" | base64
  REDIS_PASSWORD: ""  # Add your base64 encoded password if needed
```

### 2.4 Create Redis Deployment and Service (Single Instance)

Create `k8s/base/redis.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: go-queue
spec:
  replicas: 1  # Single instance for Phase 5
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:6.2-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: go-queue
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
  type: ClusterIP
```

### 2.5 Create Worker Deployment

Create `k8s/base/worker.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-queue-worker
  namespace: go-queue
spec:
  replicas: 1  # Single instance for Phase 5
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: worker
        image: go-queue-worker:latest
        imagePullPolicy: IfNotPresent
        envFrom:
        - configMapRef:
            name: go-queue-config
        - secretRef:
            name: go-queue-secrets
        env:
        - name: REDIS_HOST
          value: redis.go-queue.svc.cluster.local  # Using FQDN for reliable service discovery
        - name: REDIS_PORT
          value: "6379"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
```

### 2.6 Create API Deployment and Service

Create `k8s/base/api.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: go-queue
spec:
  replicas: 1  # Single instance for Phase 5
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: go-queue-api:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: go-queue-config
        - secretRef:
            name: go-queue-secrets
        env:
        - name: REDIS_ADDR
          value: "redis.go-queue.svc.cluster.local:6379"  # Explicitly set FQDN
        - name: REDIS_HOST
          value: "redis.go-queue.svc.cluster.local"  # Using FQDN instead of just 'redis'
        - name: REDIS_PORT
          value: "6379"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: go-queue
spec:
  selector:
    app: api
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 2.7 (Optional) Create Ingress for API

Create `k8s/base/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-queue-ingress
  namespace: go-queue
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: go-queue.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 8080
```

## 3. Create Kustomization File

Create `k8s/base/kustomization.yaml` to manage all the Kubernetes resources:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- namespace.yaml
- configmap.yaml
- secrets.yaml
- redis.yaml
- worker.yaml
- api.yaml
- ingress.yaml
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

## 5. Load Docker Images into Minikube (if using Minikube)

If you're using Minikube and local Docker images:

```bash
# Load API image into Minikube
minikube image load go-queue-api:latest

# Load Worker image into Minikube  
minikube image load go-queue-worker:latest
```

## 6. Access the API

### Option A: Using Port Forward

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

## 7. Testing the Deployment

Submit a job to the API:

```bash
curl -X POST http://localhost:8080/jobs -H "Content-Type: application/json" -d '{"payload": "test_job"}'
```

## 8. Troubleshooting

### Common Issues and Solutions

#### Redis Connection Issues

If your services cannot connect to Redis, try the following:

1. **Use Fully-Qualified Domain Names (FQDNs)**: In Kubernetes, use the format `service.namespace.svc.cluster.local` for reliable service discovery:
   ```yaml
   env:
   - name: REDIS_ADDR
     value: "redis.go-queue.svc.cluster.local:6379"
   ```

2. **Check Redis logs**:
   ```bash
   kubectl logs -n go-queue deployment/redis
   ```

3. **Test connectivity from API/Worker pods**:
   ```bash
   kubectl exec -it -n go-queue <pod-name> -- sh
   # Inside the pod
   ping redis.go-queue.svc.cluster.local
   ```

4. **Verify API/Worker logs**:
   ```bash
   kubectl logs -n go-queue deployment/api
   kubectl logs -n go-queue deployment/go-queue-worker
   ```

#### Image Pull Issues

If pods are stuck in "ImagePullBackOff" state:

```bash
# For Minikube, ensure images are loaded into Minikube's Docker daemon:
minikube image load go-queue-api:latest
minikube image load go-queue-worker:latest
```

## 9. Cleaning Up Resources

When you need to delete the go-queue resources from your Kubernetes cluster, you have several options:

### Option A: Delete Everything in the Namespace

The simplest approach is to delete the entire namespace, which will remove all resources within it:

```bash
kubectl delete namespace go-queue
```

### Option B: Delete Specific Resources

To selectively delete resources while keeping the namespace:

```bash
# Delete deployments
kubectl delete deployment -n go-queue api redis go-queue-worker

# Delete services
kubectl delete service -n go-queue api redis

# Delete ConfigMap and Secrets
kubectl delete configmap -n go-queue go-queue-config
kubectl delete secret -n go-queue go-queue-secrets

# Delete Ingress (if created)
kubectl delete ingress -n go-queue go-queue-ingress
```

### Option C: Using Kustomize

If you used Kustomize to deploy the resources, you can delete everything defined in your kustomization file:

```bash
kubectl delete -k k8s/base
```

### Cleaning Up Minikube (Optional)

If you're done with development and want to free up resources on your machine:

```bash
# Stop Minikube cluster
minikube stop

# Delete Minikube cluster
minikube delete
```
