# Advanced Kubernetes Topics Guide

## Table of Contents
1. [StatefulSets](#statefulsets)
2. [DaemonSets](#daemonsets)
3. [Ingress Controllers](#ingress-controllers)
4. [RBAC and Security](#rbac-and-security)
5. [Helm Package Manager](#helm-package-manager)

## 1. StatefulSets

**Description**: Manage stateful applications requiring stable network identities and persistent storage.

**Required Files**:
```
statefulsets/
├── 01-statefulset.yaml           # Basic StatefulSet with persistent storage
├── 02-database-statefulset.yaml  # Example MySQL/PostgreSQL cluster
├── 03-statefulset-update.yaml    # Rolling updates for StatefulSets
└── 04-backup-restore.yaml        # Backup and recovery procedures
```

**Key Concepts**:
* Ordered pod creation/deletion
* Stable network identities
* Persistent storage management
* Scaling operations

## 2. DaemonSets

**Description**: Ensure specific pods run on all (or some) nodes in the cluster.

**Required Files**:
```
daemonsets/
├── 01-basic-daemonset.yaml         # Simple monitoring agent
├── 02-node-selector-daemonset.yaml # Node-specific deployments
├── 03-logging-daemonset.yaml       # Cluster-wide logging setup
└── 04-update-strategy.yaml         # DaemonSet update configurations
```

**Key Concepts**:
* Node-level operations
* System daemon management
* Rolling updates
* Node affinity/taints

## 3. Ingress Controllers

**Description**: Manage external access to services in a cluster.

**Required Files**:
```
ingress/
├── 01-ingress-controller.yaml   # NGINX Ingress Controller setup
├── 02-ingress-rules.yaml       # Path-based routing
├── 03-tls-ingress.yaml        # SSL/TLS configuration
└── 04-custom-annotations.yaml  # Controller-specific features
```

**Key Concepts**:
* Layer 7 load balancing
* TLS termination
* Path-based routing
* Host-based routing

## 4. RBAC and Security

**Description**: Configure Role-Based Access Control and security settings.

**Required Files**:
```
rbac/
├── 01-role-rolebinding.yaml      # Basic RBAC setup
├── 02-clusterrole-binding.yaml   # Cluster-wide permissions
├── 03-service-accounts.yaml      # Service account configuration
├── 04-pod-security-policies.yaml # Security constraints
└── 05-network-policies.yaml      # Network security rules
```

**Key Concepts**:
* Roles and ClusterRoles
* ServiceAccounts
* Security Contexts
* Network Policies

## 5. Helm Package Manager

**Description**: Package management for Kubernetes applications.

**Required Files**:
```
helm/mychart/
├── Chart.yaml          # Chart metadata
├── values.yaml        # Default configuration values
├── custom-values.yaml # Environment-specific values
└── templates/         # Kubernetes manifest templates
```

**Key Concepts**:
* Chart creation
* Template functions
* Values and variables
* Chart dependencies
* Release management

## Project Structure
```
exercises/3-advanced/
├── statefulsets/
├── daemonsets/
├── ingress/
├── rbac/
├── helm/
└── README.md
```

## Prerequisites
* Working Kubernetes cluster
* `kubectl` configured
* Basic understanding of Kubernetes networking
* Familiarity with YAML syntax
* Helm CLI installed (for Helm section)


## 1. StatefulSets

```sh
kubectl apply -f exercises/3-advanced/statefulsets/

kubectl create secret generic mysql-secret --from-literal=password=yourpassword
```

## 2. DaemonSets

```sh


```