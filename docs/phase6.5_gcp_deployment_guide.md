# Phase 6.5: GCP Deployment Guide for Go Queue System

This guide provides step-by-step instructions for deploying the Go Queue system on Google Cloud Platform (GCP), optimized for a demo project using GCP's free tier with the $300 credit allowance.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [IAM Setup](#iam-setup)
3. [Project and API Setup](#project-and-api-setup)
4. [Secret Management with Secret Manager](#secret-management-with-secret-manager)
5. [Artifact Registry Setup](#artifact-registry-setup)
6. [Building and Pushing Docker Images](#building-and-pushing-docker-images)
7. [GKE Cluster Setup](#gke-cluster-setup)
8. [Cloud Memorystore Redis Setup](#cloud-memorystore-redis-setup)
9. [Kubernetes Deployment](#kubernetes-deployment)
10. [Cloud Observability Integration](#cloud-observability-integration)
11. [CI/CD Pipeline with Cloud Build](#cicd-pipeline-with-cloud-build)
12. [Testing Your Deployment](#testing-your-deployment)
13. [Next Steps](#next-steps)
14. [Post-MVP: Adding External Connectivity](#post-mvp-adding-external-connectivity)
15. [Resource Cleanup](#resource-cleanup)

## Prerequisites

Before starting, make sure you have:

1. A Google Cloud account with access to the $300 free credits
2. [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) installed on your local machine
3. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
4. [Docker](https://docs.docker.com/get-docker/) installed
5. Your Go Queue codebase ready for deployment

## IAM Setup

IAM (Identity and Access Management) is crucial for securing your GCP resources. Here's a step-by-step guide:

### 1. Create a Service Account for Deployments

```bash
# Create a service account for deploying to GKE
gcloud iam service-accounts create gke-deployer \
  --display-name="GKE Deployment Service Account"

# Get your current project ID
PROJECT_ID=$(gcloud config get-value project)
```

### 2. Assign Required Roles to the Service Account

```bash
# Assign the necessary roles for GKE deployment
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/container.developer"

# Assign the necessary roles for Artifact Registry
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# Assign roles for Secret Manager
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-deployer@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.admin"
```

### 3. Create a Service Account Key

```bash
# Create and download a key for the service account
gcloud iam service-accounts keys create ~/gke-deployer-key.json \
  --iam-account=gke-deployer@$PROJECT_ID.iam.gserviceaccount.com

# Set the GOOGLE_APPLICATION_CREDENTIALS environment variable
export GOOGLE_APPLICATION_CREDENTIALS=~/gke-deployer-key.json
```

### 4. Create a Service Account for the GKE Nodes

```bash
# Create a service account for GKE nodes
gcloud iam service-accounts create gke-node-sa \
  --display-name="GKE Node Service Account"

# Assign the necessary roles for GKE nodes
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/monitoring.metricWriter"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/monitoring.viewer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/logging.logWriter"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.reader"

# Grant access to Secret Manager
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## Project and API Setup

### 1. Create a New GCP Project (optional)

```bash
# Create a new project (optional if you already have one)
gcloud projects create go-queue-demo --name="Go Queue Demo"

# Set the project as your default
gcloud config set project go-queue-demo
```

### 2. Enable Required APIs

```bash
# Enable the necessary APIs
gcloud services enable container.googleapis.com  # GKE
gcloud services enable artifactregistry.googleapis.com  # Artifact Registry
gcloud services enable monitoring.googleapis.com  # Cloud Monitoring
gcloud services enable logging.googleapis.com  # Cloud Logging
gcloud services enable redis.googleapis.com  # Cloud Memorystore for Redis
gcloud services enable cloudtrace.googleapis.com  # Cloud Trace
gcloud services enable secretmanager.googleapis.com  # Secret Manager
gcloud services enable cloudbuild.googleapis.com  # Cloud Build
```

## Secret Management with Secret Manager

Google Secret Manager provides a secure way to store and manage your application secrets.

### 1. Create Secrets for Redis Credentials

```bash
# Create a secret for Redis password (we'll set this up later)
echo -n "your-redis-password" | \
  gcloud secrets create redis-password --data-file=-

# Create other application secrets as needed
echo -n "your-api-key" | \
  gcloud secrets create api-key --data-file=-
```

### 2. Grant Access to Your GKE Service Account

```bash
# Allow the GKE nodes to access these secrets
gcloud secrets add-iam-policy-binding redis-password \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding api-key \
  --member="serviceAccount:gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## Artifact Registry Setup

### 1. Create a Docker Repository

```bash
# Create a repository for your Docker images
gcloud artifacts repositories create go-queue-repo \
  --repository-format=docker \
  --location=us-central1 \
  --description="Docker repository for Go Queue project"
```

### 2. Configure Docker Authentication

```bash
# Configure Docker to use Google Cloud as a credential helper
gcloud auth configure-docker us-central1-docker.pkg.dev
```

## Building and Pushing Docker Images

### 1. Build and Tag Your Docker Images

```bash
# Navigate to your project directory
cd /path/to/go-queue

# Build and tag the API image
docker build -t us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:v1.0.0 -f docker/Dockerfile.api .

# Build and tag the Worker image
docker build -t us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:v1.0.0 -f docker/Dockerfile.worker .
```

### 2. Push the Images to Artifact Registry

```bash
# Push the API image
docker push us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:v1.0.0

# Push the Worker image
docker push us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:v1.0.0
```

## GKE Cluster Setup

### 1. Create a Cost-Effective GKE Cluster

For a demo using free tier credits, we'll create a small cluster in the free tier region:

```bash
# Create a minimal GKE cluster
gcloud container clusters create go-queue-cluster \
  --zone us-central1-a \
  --num-nodes=2 \
  --machine-type=e2-small \
  --disk-size=10GB \
  --service-account=gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com \
  --enable-autoscaling \
  --min-nodes=1 \
  --max-nodes=3 \
  --workload-pool=$PROJECT_ID.svc.id.goog  # Enables GKE Workload Identity
```

### 2. Configure kubectl to Use the GKE Cluster

```bash
# Get credentials to connect kubectl to your GKE cluster
gcloud container clusters get-credentials go-queue-cluster --zone us-central1-a
```

### 3. Enable Workload Identity for Secret Access

```bash
# Create a namespace for your application
kubectl create namespace go-queue

# Create a Kubernetes service account
kubectl create serviceaccount go-queue-ksa --namespace=go-queue

# Bind the Kubernetes service account to the GCP service account
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:$PROJECT_ID.svc.id.goog[go-queue/go-queue-ksa]" \
  gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com

# Annotate the Kubernetes service account
kubectl annotate serviceaccount go-queue-ksa \
  --namespace=go-queue \
  iam.gke.io/gcp-service-account=gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com
```

## Cloud Memorystore Redis Setup

Instead of deploying Redis in the cluster, we'll use Cloud Memorystore for a managed Redis instance:

```bash
# Create a Redis instance (basic tier, 1GB)
gcloud redis instances create go-queue-redis \
  --size=1 \
  --region=us-central1 \
  --redis-version=redis_6_x \
  --tier=basic \
  --auth-enabled
```

Once created, you need to get the connection information:

```bash
# Get the Redis instance IP
REDIS_IP=$(gcloud redis instances describe go-queue-redis --region=us-central1 --format='get(host)')

# Get the Redis auth string (this is the default, but Secret Manager is more secure)
REDIS_AUTH_STRING=$(gcloud redis instances get-auth-string go-queue-redis --region=us-central1)

# Store the auth string in Secret Manager
echo -n "$REDIS_AUTH_STRING" | gcloud secrets create redis-auth --data-file=-
```

## Kubernetes Deployment

Now we need to adapt your Kubernetes manifests to work with GCP services:

### 1. Create a ConfigMap for Redis Connection

```bash
# Create ConfigMap with Redis connection information
kubectl create configmap redis-config -n go-queue \
  --from-literal=redis_host=$REDIS_IP \
  --from-literal=redis_port=6379
```

### 2. Create Kubernetes Manifests for GCP Deployment

Create a directory for GCP-specific manifests:

```bash
mkdir -p k8s/gcp
```

We need to create the following files:

1. `k8s/gcp/kustomization.yaml`
2. `k8s/gcp/redis-secret.yaml`
3. `k8s/gcp/api-deployment-patch.yaml`
4. `k8s/gcp/worker-deployment-patch.yaml`

Let's create them:

#### kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: go-queue
resources:
- ../base
- redis-secret.yaml
patchesStrategicMerge:
- api-deployment-patch.yaml
- worker-deployment-patch.yaml
```

#### redis-secret.yaml

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-secret
  namespace: go-queue
type: Opaque
stringData:
  auth: "${REDIS_AUTH_STRING}"  # This is a placeholder, use Secret Manager in production
```

#### api-deployment-patch.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-queue-api
  namespace: go-queue
spec:
  template:
    spec:
      serviceAccountName: go-queue-ksa
      containers:
      - name: api
        image: us-central1-docker.pkg.dev/PROJECT_ID/go-queue-repo/go-queue-api:v1.0.0
        env:
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: redis_host
        - name: REDIS_PORT
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: redis_port
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: auth
```

#### worker-deployment-patch.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-queue-worker
  namespace: go-queue
spec:
  template:
    spec:
      serviceAccountName: go-queue-ksa
      containers:
      - name: worker
        image: us-central1-docker.pkg.dev/PROJECT_ID/go-queue-repo/go-queue-worker:v1.0.0
        env:
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: redis_host
        - name: REDIS_PORT
          valueFrom:
            configMapKeyRef:
              name: redis-config
              key: redis_port
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: auth
```

### 3. Replace PROJECT_ID in the manifests

```bash
# Replace PROJECT_ID with your actual project ID
sed -i "s/PROJECT_ID/$PROJECT_ID/g" k8s/gcp/api-deployment-patch.yaml
sed -i "s/PROJECT_ID/$PROJECT_ID/g" k8s/gcp/worker-deployment-patch.yaml
```

### 4. Apply Kubernetes Manifests

```bash
# Apply the manifests with kustomize
kubectl apply -k k8s/gcp
```

## Cloud Observability Integration

The Go Queue project already includes observability components. Let's integrate them with Google Cloud's observability stack:

### 1. Integrate with Cloud Monitoring (for metrics)

To send Prometheus metrics to Cloud Monitoring:

1. Deploy the Google Cloud Managed Service for Prometheus

```bash
# Enable GMP
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/prometheus-engine/main/manifests/setup.yaml
```

2. Create a PodMonitoring resource to collect metrics from your application pods:

```yaml
# Save this as k8s/gcp/pod-monitoring.yaml
apiVersion: monitoring.googleapis.com/v1
kind: PodMonitoring
metadata:
  name: go-queue-monitoring
  namespace: go-queue
spec:
  selector:
    matchLabels:
      app: go-queue-api  # Match your pod labels
  endpoints:
  - port: metrics  # The port exposing Prometheus metrics in your container
    interval: 30s
```

3. Apply the PodMonitoring resource:

```bash
kubectl apply -f k8s/gcp/pod-monitoring.yaml
```

### 2. Integrate with Cloud Logging

Your application likely already outputs structured logs. To ensure they're properly captured by Cloud Logging:

1. Make sure your logs are in JSON format
2. Include appropriate severity levels

For more detailed logging, you can deploy Fluent Bit as a DaemonSet to customize log collection:

```bash
# Install Fluent Bit via Helm
helm repo add fluent https://fluent.github.io/helm-charts
helm repo update
helm install fluent-bit fluent/fluent-bit --namespace kube-system
```

### 3. Integrate with Cloud Trace

To integrate with Cloud Trace:

1. Update your application code to use the Google Cloud Trace exporter for OpenTelemetry

Example code snippet (you would integrate this into your existing tracing setup):

```go
import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/gcpexporter"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := gcpexporter.New()
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
```

## CI/CD Pipeline with Cloud Build

Let's set up a basic CI/CD pipeline using Cloud Build:

### 1. Create a Cloud Build Configuration File

Create a file named `cloudbuild.yaml` in the root of your project:

```yaml
steps:
# Build and tag API image
- name: 'gcr.io/cloud-builders/docker'
  args: ['build', '-t', 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:$TAG_NAME', '-f', 'docker/Dockerfile.api', '.']

# Build and tag Worker image
- name: 'gcr.io/cloud-builders/docker'
  args: ['build', '-t', 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:$TAG_NAME', '-f', 'docker/Dockerfile.worker', '.']

# Push API image to Artifact Registry
- name: 'gcr.io/cloud-builders/docker'
  args: ['push', 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:$TAG_NAME']

# Push Worker image to Artifact Registry
- name: 'gcr.io/cloud-builders/docker'
  args: ['push', 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:$TAG_NAME']

# Update deployment manifests with new image tags
- name: 'gcr.io/cloud-builders/kubectl'
  entrypoint: 'bash'
  args:
  - '-c'
  - |
    sed -i "s|us-central1-docker.pkg.dev/PROJECT_ID/go-queue-repo/go-queue-api:[^ ]*|us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:$TAG_NAME|" k8s/gcp/api-deployment-patch.yaml
    sed -i "s|us-central1-docker.pkg.dev/PROJECT_ID/go-queue-repo/go-queue-worker:[^ ]*|us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:$TAG_NAME|" k8s/gcp/worker-deployment-patch.yaml

# Apply changes using kubectl
- name: 'gcr.io/cloud-builders/kubectl'
  args: ['apply', '-k', 'k8s/gcp']
  env:
  - 'CLOUDSDK_COMPUTE_ZONE=us-central1-a'
  - 'CLOUDSDK_CONTAINER_CLUSTER=go-queue-cluster'

images:
- 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:$TAG_NAME'
- 'us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-worker:$TAG_NAME'
```

### 2. Connect Your GitHub Repository to Cloud Build

1. Navigate to Cloud Build in the GCP Console
2. Go to Triggers
3. Click "Connect Repository"
4. Select GitHub as the source
5. Authenticate and select your repository
6. Create a trigger with the following settings:
   - Name: go-queue-build-deploy
   - Event: Tag
   - Source: ^v\d+\.\d+\.\d+$ (matches version tags like v1.0.0)
   - Configuration: Cloud Build configuration file (cloudbuild.yaml)
   - Location: Repository

### 3. Grant Required Permissions to Cloud Build Service Account

```bash
# Get the Cloud Build service account
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
CLOUDBUILD_SA="$PROJECT_NUMBER@cloudbuild.gserviceaccount.com"

# Grant GKE developer permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$CLOUDBUILD_SA" \
  --role="roles/container.developer"
```

### 4. Trigger a Build

To trigger a build, create and push a new tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Testing Your Deployment

### 1. Check Deployment Status

```bash
# Check the status of your pods
kubectl get pods -n go-queue

# Check the status of your services
kubectl get services -n go-queue
```

### 2. Access the API Service Internally

```bash
# Create a port-forward to the API service
kubectl port-forward -n go-queue svc/go-queue-api 8080:80
```

Then access your API at http://localhost:8080

### 3. Test Job Processing

Send requests to your API and monitor the worker processing:

```bash
# Watch logs from worker pods
kubectl logs -n go-queue -l app=go-queue-worker -f
```

### 4. Check Monitoring Dashboards

1. Go to Google Cloud Console
2. Navigate to Monitoring > Dashboards
3. Create a custom dashboard for your Go Queue metrics

## Next Steps

1. **Scaling Strategy**: Based on your application needs and traffic patterns, you'll want to define appropriate scaling strategies. For this, consider:
   - Horizontal Pod Autoscaler for worker pods based on queue depth
   - Vertical Pod Autoscaler for optimizing resource allocation
   - Node autoscaling for cluster optimization

2. **High Availability Testing**: Implement chaos testing to validate your system's resilience:
   - Pod termination tests
   - Node failure simulation
   - Network partition tests

3. **Performance Optimization**: Analyze metrics to identify bottlenecks:
   - Redis connection pooling
   - Worker concurrency tuning
   - API response time optimization

Please refer to the separate cost management guide for detailed information on budget controls and cost optimization strategies for your GCP deployment.

## Post-MVP: Adding External Connectivity

While the initial deployment focuses on a secure, internal-only system, you might need to expose your API to external users after successful internal validation. This section provides guidance on securely expanding your deployment to support external connectivity.

### 1. External Access Architecture Options

#### Option A: Cloud Load Balancer with Ingress

```bash
# Create a static IP address for your ingress
gcloud compute addresses create go-queue-ip --global

# Get the reserved IP address
GO_QUEUE_IP=$(gcloud compute addresses describe go-queue-ip --global --format='get(address)')
echo "Your reserved IP: $GO_QUEUE_IP"

# Update your DNS records with this IP
```

Create a Kubernetes Ingress resource:

```yaml
# k8s/gcp/external/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-queue-api-ingress
  namespace: go-queue
  annotations:
    kubernetes.io/ingress.class: "gce"
    networking.gke.io/managed-certificates: "go-queue-cert"
    kubernetes.io/ingress.global-static-ip-name: "go-queue-ip"
spec:
  rules:
  - host: api.go-queue.example.com  # Replace with your domain
    http:
      paths:
      - path: /*
        pathType: ImplementationSpecific
        backend:
          service:
            name: go-queue-api
            port:
              number: 8080
```

Create a managed certificate:

```yaml
# k8s/gcp/external/managed-cert.yaml
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: go-queue-cert
  namespace: go-queue
spec:
  domains:
  - api.go-queue.example.com  # Replace with your domain
```

Apply the configurations:

```bash
kubectl apply -f k8s/gcp/external/managed-cert.yaml
kubectl apply -f k8s/gcp/external/ingress.yaml

# Check status
kubectl describe managedcertificate go-queue-cert -n go-queue
kubectl get ingress -n go-queue
```

#### Option B: Cloud Endpoints with API Gateway

1. Create an OpenAPI specification for your API:

```yaml
# openapi-spec.yaml
swagger: '2.0'
info:
  title: Go Queue API
  description: API for the Go Queue job processing system
  version: 1.0.0
host: api.go-queue.example.com
schemes:
  - https
paths:
  /jobs:
    post:
      summary: Create a new job
      operationId: createJob
      # ... rest of your API specification ...
```

2. Deploy the API configuration to Cloud Endpoints:

```bash
gcloud endpoints services deploy openapi-spec.yaml
```

3. Create an API Gateway:

```bash
# Create API Config
gcloud api-gateway api-configs create go-queue-config \
  --api=go-queue-api \
  --openapi-spec=openapi-spec.yaml \
  --project=$PROJECT_ID

# Create Gateway
gcloud api-gateway gateways create go-queue-gateway \
  --api=go-queue-api \
  --api-config=go-queue-config \
  --location=us-central1 \
  --project=$PROJECT_ID
```

#### Option C: VPN or Cloud Interconnect

For enterprise clients requiring secure, private access:

```bash
# Create a Cloud VPN gateway
gcloud compute vpn-gateways create go-queue-vpn-gateway \
  --network=default \
  --region=us-central1

# Continue with VPN tunnel configuration based on client requirements
```

### 2. Security Hardening for External Access

#### Implement Cloud Armor Protection

```bash
# Create a Cloud Armor security policy
gcloud compute security-policies create go-queue-security-policy \
  --description "Security policy for Go Queue API"

# Add a rule to block common web attacks
gcloud compute security-policies rules create 1000 \
  --security-policy go-queue-security-policy \
  --expression "evaluatePreconfiguredExpr('xss-stable')" \
  --action "deny-403"

# Add geographical restrictions if needed (example: only allow US and Canada)
gcloud compute security-policies rules create 1100 \
  --security-policy go-queue-security-policy \
  --expression "origin.region_code != 'US' && origin.region_code != 'CA'" \
  --action "deny-403" \
  --description "Geo-restriction: Allow US and Canada only"

# Apply the policy to your backend service (after Ingress is created)
# Note: You'll need to get the backend service name from the created ingress
BACKEND_NAME=$(gcloud compute backend-services list --filter="description~go-queue-api-ingress" --format="value(name)")
gcloud compute backend-services update $BACKEND_NAME \
  --security-policy go-queue-security-policy
```

#### Implement API Authentication

1. Update your Go API implementation to validate JWT tokens:

```go
// Example JWT validation middleware (add to your API code)
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Strip "Bearer " prefix if present
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		
		// Validate token (implement with your JWT library of choice)
		// ...validation code...
		
		// Continue to next handler if token is valid
		next.ServeHTTP(w, r)
	})
}
```

2. Create API keys for clients:

```bash
# Create an API key in Secret Manager
echo -n "$(openssl rand -base64 32)" | \
  gcloud secrets create api-key-client1 --data-file=-

# Get the API key (for distribution to client)
gcloud secrets versions access latest --secret=api-key-client1
```

### 3. Enhanced Observability for External Traffic

#### Set up request logging and monitoring

1. Deploy a custom monitoring dashboard for external traffic:

```bash
# Create a Cloud Monitoring dashboard (using the API or Console)
cat <<EOF > external_dashboard.json
{
  "displayName": "Go Queue External Traffic",
  "gridLayout": {
    "columns": "2",
    "widgets": [
      {
        "title": "HTTP Request Count",
        "xyChart": {
          "dataSets": [
            {
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"kubernetes.io/container/network/received_bytes_count\" resource.type=\"k8s_container\" resource.label.\"pod_name\"=monitoring.regex.full_match(\"go-queue-api.*\")"
                }
              }
            }
          ]
        }
      },
      {
        "title": "Error Rate",
        "xyChart": {
          "dataSets": [
            {
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"logging.googleapis.com/log_entry_count\" resource.type=\"k8s_container\" resource.label.\"pod_name\"=monitoring.regex.full_match(\"go-queue-api.*\") severity>=ERROR"
                }
              }
            }
          ]
        }
      }
    ]
  }
}
EOF

# Deploy the dashboard
gcloud monitoring dashboards create --config-from-file=external_dashboard.json
```

2. Set up alerting for suspicious traffic patterns:

```bash
# Create an alerting policy for unusual traffic spikes
gcloud alpha monitoring policies create \
  --notification-channels="projects/$PROJECT_ID/notificationChannels/CHANNEL_ID" \
  --display-name="Unusual API Traffic" \
  --condition-filter="resource.type = \"k8s_container\" AND resource.labels.pod_name =~ \"go-queue-api.*\" AND metric.type = \"kubernetes.io/container/network/received_bytes_count\" AND metric.label.response_code >= 400" \
  --condition-threshold="{\"comparison\":\"COMPARISON_GT\", \"thresholdValue\":100, \"duration\":\"60s\"}"
```

### 4. Gradual Rollout Strategy

Follow these steps for a safe, gradual rollout to external users:

1. Initial limited access:

```bash
# Create a restricted Cloud Armor rule during initial rollout
gcloud compute security-policies rules create 900 \
  --security-policy go-queue-security-policy \
  --expression "inIpRange(origin.ip, '203.0.113.0/24')" \
  --action "allow" \
  --description "Allow only trusted IP range during initial rollout"

# Add the default deny rule at the end
gcloud compute security-policies rules create 2147483647 \
  --security-policy go-queue-security-policy \
  --expression "true" \
  --action "deny-403" \
  --description "Default deny during limited rollout"
```

2. Expand access gradually:

```bash
# As you gain confidence, update the policy to allow more traffic
# First remove the default deny
gcloud compute security-policies rules delete 2147483647 \
  --security-policy go-queue-security-policy

# Then add more IP ranges or remove IP restrictions entirely
gcloud compute security-policies rules update 900 \
  --security-policy go-queue-security-policy \
  --expression "true" \
  --action "allow" \
  --description "Allow all traffic after successful validation"
```

### 5. External Connectivity Testing

Before full public rollout, conduct these essential tests:

```bash
# Test with curl from an external machine
curl -v -H "Authorization: Bearer $YOUR_TOKEN" https://api.go-queue.example.com/health

# Load testing with hey tool
go install github.com/rakyll/hey@latest
hey -n 1000 -c 50 -H "Authorization: Bearer $YOUR_TOKEN" https://api.go-queue.example.com/jobs
```

### 6. Documentation for API Consumers

Create comprehensive documentation for your newly exposed API:

1. Install and deploy API documentation:

```bash
# Example using Redoc for OpenAPI docs
kubectl create configmap api-docs --from-file=openapi-spec.yaml -n go-queue

# Deploy a documentation service
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-docs
  namespace: go-queue
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api-docs
  template:
    metadata:
      labels:
        app: api-docs
    spec:
      containers:
      - name: redoc
        image: redocly/redoc
        ports:
        - containerPort: 80
        volumeMounts:
        - name: api-spec
          mountPath: /usr/share/nginx/html/openapi.yaml
          subPath: openapi-spec.yaml
      volumes:
      - name: api-spec
        configMap:
          name: api-docs
---
apiVersion: v1
kind: Service
metadata:
  name: api-docs
  namespace: go-queue
spec:
  selector:
    app: api-docs
  ports:
  - port: 80
    targetPort: 80
EOF

# Add the docs service to your ingress
# Update your existing ingress or create a new path
```

By following this guide, you can safely extend your initially private Go Queue deployment to support external connectivity, with appropriate security measures, observability, and gradual rollout strategies.

## Resource Cleanup

When you're done with your deployment or want to avoid incurring charges after testing, follow these steps to clean up all GCP resources created during this deployment process.

### 1. Delete Kubernetes Resources

First, remove all Kubernetes resources deployed to your GKE cluster:

```bash
# Delete all resources in the go-queue namespace
kubectl delete namespace go-queue

# If you've deployed monitoring resources
kubectl delete -f k8s/gcp/pod-monitoring.yaml

# If you deployed Fluent Bit
helm uninstall fluent-bit --namespace kube-system
```

### 2. Delete GKE Cluster

Remove the GKE cluster (this will delete all nodes and workloads):

```bash
# Delete the GKE cluster
gcloud container clusters delete go-queue-cluster --zone us-central1-a --quiet
```

### 3. Delete Cloud Memorystore Redis Instance

```bash
# Delete the Redis instance
gcloud redis instances delete go-queue-redis --region=us-central1 --quiet
```

### 4. Remove Secrets from Secret Manager

```bash
# Delete all secrets
gcloud secrets delete redis-password --quiet
gcloud secrets delete api-key --quiet
gcloud secrets delete redis-auth --quiet
```

### 5. Clean Up Artifact Registry

```bash
# Delete the artifacts repository with all images
gcloud artifacts repositories delete go-queue-repo --location=us-central1 --quiet
```

### 6. Delete Service Accounts and Keys

```bash
# Get your project ID
PROJECT_ID=$(gcloud config get-value project)

# Delete service accounts
gcloud iam service-accounts delete gke-deployer@$PROJECT_ID.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com --quiet

# Remove the service account key file from your local machine
rm ~/gke-deployer-key.json
```

### 7. Disable APIs (Optional)

If you enabled APIs specifically for this project and don't need them anymore:

```bash
gcloud services disable container.googleapis.com
gcloud services disable artifactregistry.googleapis.com
gcloud services disable monitoring.googleapis.com
gcloud services disable logging.googleapis.com
gcloud services disable redis.googleapis.com
gcloud services disable cloudtrace.googleapis.com
gcloud services disable secretmanager.googleapis.com
gcloud services disable cloudbuild.googleapis.com
```

### 8. Delete the Project (Optional)

If you created a project specifically for this deployment and want to remove everything at once:

```bash
# Replace with your project ID
gcloud projects delete go-queue-demo --quiet
```

> **Important:** Deleting a project is irreversible and removes ALL resources within that project. Only do this if you're certain you don't need anything in the project.

### 9. Verify Cleanup

After cleanup, verify that all billable resources have been removed by checking your GCP console's billing section:

1. Go to the [GCP Billing page](https://console.cloud.google.com/billing)
2. Select your billing account
3. Go to "Reports" and check that there are no unexpected costs

It may take some time for the billing system to reflect all deletions. Check again after 24 hours to ensure all resources have been properly removed.
