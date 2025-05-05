# Intermediate Kubernetes Concepts

This section covers more advanced Kubernetes concepts and patterns that build upon the basic fundamentals. Each topic includes practical examples and explanations.

## 1. Horizontal Pod Autoscaling (HPA)

HPA automatically adjusts the number of pods in a deployment based on observed metrics (typically CPU utilization).

### Key Components:
- `02-php-apache.yaml`: Base deployment with resource limits
- `03-hpa.yaml`: HPA configuration
  
### Features:
- Scales between 1-10 replicas
- Targets 50% CPU utilization
- Custom scaling behaviors:
  - Rapid scale up (0s stabilization window)
  - Conservative scale down (300s stabilization window)

### Testing:
```bash
kubectl apply -f deployments/02-php-apache.yaml
kubectl apply -f deployments/03-hpa.yaml
kubectl get hpa
```

## 2. Pod Disruption Budget (PDB)

PDBs protect applications from voluntary disruptions by ensuring a minimum number of pods remain available.

### Key Components:
- `04-pdb.yaml`: PDB configuration ensuring minimum availability

### Features:
- Ensures at least 1 pod is always available
- Protects against voluntary disruptions (node drains, upgrades)
- Works with the php-apache deployment

## 3. Init Containers and Multi-container Patterns

Demonstrates various patterns for running multiple containers in a pod and initialization containers.

### Key Components:
- `01-multi-container-patterns.yaml`: Pod with init and sidecar containers

### Features:
- Init container for content preparation
- Main container running nginx
- Sidecar container for log monitoring
- Shared volumes between containers
- Common patterns:
  - Sidecar: Enhances main container
  - Init: Performs setup tasks
  - Shared storage: Communication between containers

## 4. Pod Affinity and Anti-affinity

Controls pod scheduling based on node topology and existing pod placements.

### Key Components:
- `02-pod-affinity.yaml`: Deployment with affinity rules

### Features:
- Pod Affinity: Requires co-location with cache pods
- Pod Anti-affinity: Spreads replicas across nodes
- Uses both hard (required) and soft (preferred) rules
- Topology key based on hostname

## 5. Custom Resource Definitions (CRDs)

Extends Kubernetes API with custom resources tailored to your needs.

### Key Components:
- `01-webapp-crd.yaml`: CRD definition
- `02-webapp-example.yaml`: Example custom resource

### Features:
- Defines a new WebApp resource type
- Validates required fields (image, port)
- Custom schema validation
- Shortname alias (wa)

### Usage:
```bash
kubectl apply -f config/01-webapp-crd.yaml
kubectl apply -f config/02-webapp-example.yaml
kubectl get webapps
```

## 6. Service Mesh Integration

Basic service mesh implementation using Istio patterns.

### Key Components:
- `01-service-mesh.yaml`: Deployment for `ratings-v1` with sidecar injection.
- `ratings-v2.yaml`: Deployment for `ratings-v2` (if traffic is split).
- `destination-rule.yaml`: Defines routing subsets.
- `02-virtual-service.yaml`: Traffic management rules (90/10 split between v1 and v2).

### Features:
- Automatic sidecar injection
- Traffic splitting (90/10)
- Version-based routing using `DestinationRule`
- Security context configuration (`runAsUser: 1000`)

### Prerequisites:
- Istio must be installed in the cluster
- Sidecar injection must be enabled in the namespace

## Best Practices

1. Resource Management:
   - Always set resource requests and limits
   - Use HPA for automated scaling
   - Configure PDBs for critical services

2. High Availability:
   - Use pod anti-affinity to spread replicas
   - Configure appropriate PDBs
   - Implement proper health checks

3. Service Mesh:
   - Enable mTLS for service-to-service communication
   - Use virtual services for traffic management
   - Implement proper monitoring and tracing

4. Custom Resources:
   - Keep CRDs focused and specific
   - Implement proper validation
   - Use shortnames for convenience

## Testing and Verification

Each concept can be tested using the following steps:

1. HPA Testing:
```bash
kubectl apply -f deployments/02-php-apache.yaml
kubectl apply -f deployments/03-hpa.yaml
# Generate load
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://php-apache; done"
```

2. Multi-container Testing:
```bash
kubectl apply -f pods/01-multi-container-patterns.yaml
kubectl logs web-app -c logger-sidecar
```

3. Affinity Testing:
```bash
kubectl apply -f pods/02-pod-affinity.yaml
kubectl get pods -o wide
```

4. CRD Testing:
```bash
kubectl apply -f config/01-webapp-crd.yaml
kubectl apply -f config/02-webapp-example.yaml
kubectl get webapps
kubectl describe webapp my-webapp
```

## Troubleshooting

Common issues and solutions:

1. HPA not scaling:
   - Check metrics-server installation
   - Verify resource requests/limits
   - Check HPA configuration

2. Pod scheduling issues:
   - Review node labels
   - Check affinity rules
   - Verify available resources

3. Service mesh problems:
   - Verify Istio installation
   - Check sidecar injection
   - Review virtual service configuration

4. CRD issues:
   - Validate CRD schema
   - Check API version compatibility
   - Verify RBAC permissions