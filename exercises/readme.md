# Creating directory structure for learning

```sh
mkdir exercises
```

```sh
# Create deployment, config, services and pods directories across all levels
New-Item -ItemType Directory -Path "exercises\2-intermediate","exercises\2-intermediate\pods","exercises\2-intermediate\deployments","exercises\2-intermediate\services","exercises\2-intermediate\config","exercises\3-advanced","exercises\3-advanced\pods","exercises\3-advanced\deployments","exercises\3-advanced\services","exercises\3-advanced\config"

# Create storage, security, and jobs directories across all levels
New-Item -ItemType Directory -Path `
    "exercises\1-basics\storage","exercises\1-basics\security","exercises\1-basics\jobs",`
    "exercises\2-intermediate\storage","exercises\2-intermediate\security","exercises\2-intermediate\jobs",`
    "exercises\3-advanced\storage","exercises\3-advanced\security","exercises\3-advanced\jobs"
```

# Installing minikube

```sh
# Install Minikube if not already installed
winget install Minikube
# or


# Checking minikube is installed
minikube version
minikube status
```