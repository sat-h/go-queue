Q. How to change the number of replicas of a service in Kubernetes?
A. To change the number of replicas of a service in Kubernetes, you can update the `replicas` field in the deployment YAML file and then apply the changes using `kubectl apply -f <deployment-file.yaml>`. Alternatively, you can use the command:
```bash
```

Q. How do you change the number of nodes in a local Minikube Kubernetes cluster?
A. You change the number of nodes when starting Minikube by using the `--nodes` flag. For example:
```bash
minikube start --nodes=3
```

Q. Is the level of complexity when having two replicas of API and two replicas of Redis similar in Kubernetes? A. No, the complexity is not similar.  
API replicas: Scaling stateless services like API is straightforward—set replicas: 2 in the Deployment. Kubernetes handles load balancing and pod management automatically.
Redis replicas: Scaling Redis is more complex because it is stateful. You need to manage master-replica roles, data synchronization, and failover, typically using Redis Sentinel or Redis HA manifests (redis-ha.yaml). Simply increasing the replicas field is not enough.
Summary: Scaling stateless services (API) is simple; scaling stateful services (Redis) requires additional configuration and orchestration.

????A Kubernetes Service of type ClusterIP is the default service type that exposes a group of pods on a stable, internal IP address within the cluster. This IP (the ClusterIP) is only accessible from within the Kubernetes cluster, not from outside. It enables reliable communication and load balancing between pods.

