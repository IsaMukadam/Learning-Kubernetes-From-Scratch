# Final Basic Kubernetes Concepts

## Resource Management
Resource management in Kubernetes helps ensure efficient resource utilization and prevents resource starvation.

### Key Concepts:
1. **Resource Requests**
   - Minimum resources guaranteed to the container
   - Used for scheduling decisions
   - Helps Kubernetes place pods on nodes with sufficient resources

2. **Resource Limits**
   - Maximum resources a container can use
   - Container will be throttled if it exceeds CPU limits
   - Container will be terminated if it exceeds memory limits

### Resource Units:
- **CPU**: 
  - Measured in cores or millicores
  - 1000m (millicores) = 1 CPU core
  - Example: "250m" = 0.25 CPU cores

- **Memory**:
  - Measured in bytes
  - Common units: Mi (Mebibytes), Gi (Gibibytes)
  - Example: "64Mi" = 64 Mebibytes

## Health Checks
Health checks help Kubernetes monitor container health and ensure application availability.

### Types of Probes:

1. **Liveness Probe**
   - Determines if container is alive
   - Failed probe causes container restart
   - Use for detecting deadlocks or infinite loops

2. **Readiness Probe**
   - Determines if container can serve traffic
   - Failed probe removes pod from service endpoints
   - Use for startup checks and dependency readiness

### Probe Configuration:
- **initialDelaySeconds**: Wait before first probe
- **periodSeconds**: Time between probes
- **timeoutSeconds**: Probe timeout duration
- **failureThreshold**: Max consecutive failures
- **successThreshold**: Min consecutive successes

### Probe Types:
1. HTTP GET
2. TCP Socket
3. Exec (run command in container)

## Testing Commands

Test Resource Management:
```bash
kubectl apply -f 05-resource-management.yaml
kubectl top pods -l app=nginx-resources
kubectl describe pods -l app=nginx-resources
```

Test Health Checks:
```bash
kubectl apply -f 06-health-checks.yaml
kubectl get pods -l app=nginx-health -w
kubectl describe pods -l app=nginx-health
```