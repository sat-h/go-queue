<hr></hr> Generate an infographic-style Kubernetes architecture diagram for a local Minikube deployment of the go-queue system. The diagram should include:  
A single-node cluster (Minikube VM)
Two API pods (replicas), each labeled with its own podip
Two Worker pods (replicas), each labeled with its own podip
One Redis pod, labeled with its podip
Show the internal networking, including ClusterIP Service for API
Use clear infographic design elements: icons for pods, node, and services, color coding, and labeled connections
Indicate that all pods are running on the same node
Visualize pod-to-pod and service-to-pod communication paths
The diagram should be clean, modern, and easy to understand for someone learning Kubernetes basics.