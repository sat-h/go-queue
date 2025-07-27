# GCP Cost Management Guide for Go Queue Project

This guide focuses on managing costs for your Go Queue project deployment on Google Cloud Platform, helping you stay within the $300 free credits allowance.

## Table of Contents
1. [Understanding GCP Free Tier](#understanding-gcp-free-tier)
2. [Setting Up Budget Controls](#setting-up-budget-controls)
3. [Cost Optimization for GKE](#cost-optimization-for-gke)
4. [Cost Optimization for Cloud Memorystore](#cost-optimization-for-cloud-memorystore)
5. [Cost Optimization for Artifact Registry](#cost-optimization-for-artifact-registry)
6. [Monitoring Your Costs](#monitoring-your-costs)
7. [Cost-Effective Scaling](#cost-effective-scaling)
8. [Clean-Up Procedures](#clean-up-procedures)

## Understanding GCP Free Tier

Google Cloud offers a $300 credit for new accounts along with an "Always Free" tier for certain resources. To maximize your credits:

1. **Always Free Resources**: Prioritize using these when possible
   - 1 e2-micro VM instance per month
   - 5GB of regional storage in Cloud Storage
   - Small Artifact Registry storage (0.5GB)
   - Small Cloud Build minutes

2. **Limited Free Credits**: These consume your $300 credit
   - GKE cluster usage
   - Cloud Memorystore instances
   - Higher-tier VMs
   - Premium networking features

## Setting Up Budget Controls

### 1. Create a Budget Alert

```bash
# Get your billing account ID
BILLING_ACCOUNT=$(gcloud billing accounts list --format="value(name.basename())")

# Create a budget alert at 50%, 75%, and 90% of $200 (saving $100 buffer)
gcloud billing budgets create \
  --billing-account=$BILLING_ACCOUNT \
  --display-name="Go Queue Project Budget" \
  --budget-amount=200USD \
  --threshold-rules=threshold-percent=50,basis=current-spend \
  --threshold-rules=threshold-percent=75,basis=current-spend \
  --threshold-rules=threshold-percent=90,basis=current-spend \
  --all-updates-rule-pubsub-topic=projects/$PROJECT_ID/topics/budget-alerts
```

### 2. Create a Cloud Function to Handle Budget Alerts (Optional)

You can create a Cloud Function that receives budget notifications and takes actions (like sending an email or shutting down resources).

### 3. Set Up Export to BigQuery for Detailed Analysis

```bash
# Enable the BigQuery API
gcloud services enable bigquery.googleapis.com

# Set up billing export to BigQuery
gcloud billing accounts export billing-export-dataset billing-export-table \
  --billing-account=$BILLING_ACCOUNT
```

## Cost Optimization for GKE

GKE can be the most significant cost in your demo project. Optimize with these strategies:

### 1. Use the Smallest Viable Node Size

```bash
# For demo purposes, use e2-small instances
gcloud container clusters create go-queue-cluster \
  --machine-type=e2-small \
  --disk-size=10 \
  --num-nodes=2
```

### 2. Use Spot Instances for Non-Critical Workloads

For development or testing, consider using spot instances which are up to 60-91% cheaper:

```bash
# Create a spot node pool
gcloud container node-pools create spot-pool \
  --cluster=go-queue-cluster \
  --machine-type=e2-small \
  --spot \
  --num-nodes=1
```

### 3. Schedule Cluster Shutdown During Non-Working Hours

For a demo project, you don't need the cluster running 24/7:

```bash
# Create a Cloud Scheduler job to stop the cluster at night
gcloud scheduler jobs create http shutdown-cluster \
  --schedule="0 20 * * 1-5" \
  --uri="https://container.googleapis.com/v1/projects/$PROJECT_ID/zones/us-central1-a/clusters/go-queue-cluster:stop" \
  --oauth-service-account-email="$PROJECT_ID@appspot.gserviceaccount.com"

# Create a job to start the cluster in the morning
gcloud scheduler jobs create http start-cluster \
  --schedule="0 8 * * 1-5" \
  --uri="https://container.googleapis.com/v1/projects/$PROJECT_ID/zones/us-central1-a/clusters/go-queue-cluster:start" \
  --oauth-service-account-email="$PROJECT_ID@appspot.gserviceaccount.com"
```

### 4. Enable GKE Autopilot for Hands-Off Management (Alternative)

If you prefer a fully-managed option:

```bash
# Create an Autopilot cluster
gcloud container clusters create-auto go-queue-autopilot \
  --region=us-central1
```

## Cost Optimization for Cloud Memorystore

### 1. Choose the Smallest Instance Size

```bash
# Create a 1GB Basic tier Redis instance
gcloud redis instances create go-queue-redis \
  --size=1 \
  --region=us-central1 \
  --redis-version=redis_6_x \
  --tier=basic
```

### 2. Pause/Recreate the Instance When Not Needed

Cloud Memorystore doesn't have a native "pause" feature, but you can:

1. Export your data (if needed)
2. Delete the instance when not in use
3. Recreate it when needed

Create scripts to automate this:

```bash
# Delete Redis instance
gcloud redis instances delete go-queue-redis --region=us-central1 --quiet

# Recreate Redis instance
gcloud redis instances create go-queue-redis \
  --size=1 \
  --region=us-central1 \
  --redis-version=redis_6_x \
  --tier=basic
```

## Cost Optimization for Artifact Registry

Artifact Registry costs include storage and data transfer:

### 1. Clean Up Old Images Regularly

```bash
# List all image versions
gcloud artifacts docker images list us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo

# Delete old images
gcloud artifacts docker images delete us-central1-docker.pkg.dev/$PROJECT_ID/go-queue-repo/go-queue-api:v1.0.0
```

### 2. Set Up Cleanup Rules

```bash
# Create a cleanup policy to keep only recent images
gcloud artifacts repositories set-cleanup-policies go-queue-repo \
  --location=us-central1 \
  --policy='{
    "policies": [{
      "condition": {
        "tagState": "TAGGED",
        "tagPrefixes": ["v"],
        "olderThan": "30d"
      },
      "action": {
        "type": "DELETE"
      }
    }]
  }'
```

## Monitoring Your Costs

### 1. Use the GCP Billing Dashboard

Regularly check the Billing section in Google Cloud Console to monitor your spending.

### 2. Set Up Custom Cost Reports

```bash
# Create a custom report for your project
1. Go to Billing in the GCP Console
2. Select "Reports"
3. Create a new report filtered by your project ID
4. Group by service and SKU to see detailed breakdowns
```

### 3. Use Cost Insights API (for Programmatic Monitoring)

```python
from google.cloud import billing

def get_cost_insights():
    client = billing.CloudCatalogClient()
    # Get service details
    services = client.list_services()
    for service in services:
        print(f"Service: {service.display_name}")
```

## Cost-Effective Scaling

For your Go Queue application, consider these cost-effective scaling strategies:

### 1. Horizontal Pod Autoscaler Based on Custom Metrics

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: go-queue-worker-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: go-queue-worker
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: External
    external:
      metric:
        name: redis_list_length
      target:
        type: AverageValue
        averageValue: 10
```

### 2. Node Auto-Provisioning

Enable cluster autoscaler to automatically provision nodes based on pod demands:

```bash
gcloud container clusters update go-queue-cluster \
  --enable-autoprovisioning \
  --min-cpu=1 \
  --min-memory=1 \
  --max-cpu=4 \
  --max-memory=16
```

### 3. Implement Rate Limiting on API

To prevent unexpected scaling due to traffic spikes, implement rate limiting:

```go
// Example using go-redis rate limiter
import "github.com/go-redis/redis_rate/v9"

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        limiter := redis_rate.NewLimiter(redisClient)
        res, err := limiter.Allow(ctx, "project:api", redis_rate.PerSecond(10))
        // Handle rate limiting logic
    })
}
```

## Clean-Up Procedures

To ensure you don't incur unexpected costs, have clear procedures for cleaning up resources:

### 1. Create a Cleanup Script

Save this as `cleanup-gcp.sh`:

```bash
#!/bin/bash
# GCP Resource Cleanup Script

# Set your project ID
PROJECT_ID=$(gcloud config get-value project)

# Delete GKE cluster
echo "Deleting GKE cluster..."
gcloud container clusters delete go-queue-cluster --zone=us-central1-a --quiet

# Delete Cloud Memorystore instance
echo "Deleting Redis instance..."
gcloud redis instances delete go-queue-redis --region=us-central1 --quiet

# Delete Artifact Registry repository
echo "Deleting Artifact Registry repo..."
gcloud artifacts repositories delete go-queue-repo --location=us-central1 --quiet

# Delete secrets
echo "Deleting secrets..."
gcloud secrets delete redis-password --quiet
gcloud secrets delete api-key --quiet

# Delete service accounts
echo "Deleting service accounts..."
gcloud iam service-accounts delete gke-deployer@$PROJECT_ID.iam.gserviceaccount.com --quiet
gcloud iam service-accounts delete gke-node-sa@$PROJECT_ID.iam.gserviceaccount.com --quiet

echo "Cleanup complete!"
```

### 2. Set Up a Billing Budget Alert at 90% of Your Budget

```bash
# Create a critical alert at 90% of budget
gcloud billing budgets update budget-name \
  --threshold-rules=threshold-percent=90,basis=current-spend,spend-basis=forecasted-spend
```

### 3. Consider Setting an Expiration Date

For a demo project, consider setting an automatic expiration date:

```bash
# Schedule project cleanup after demo period (e.g., 30 days)
gcloud scheduler jobs create http cleanup-project \
  --schedule="0 0 * * *" \
  --time-zone="UTC" \
  --description="Check if project should be cleaned up" \
  --uri="https://us-central1-$PROJECT_ID.cloudfunctions.net/checkProjectExpiration" \
  --oauth-service-account-email="$PROJECT_ID@appspot.gserviceaccount.com"
```

## Conclusion

By following these guidelines, you can make the most of your $300 GCP credits while deploying a functional Go Queue system. Remember to regularly monitor your costs and be proactive about shutting down resources when not needed.
