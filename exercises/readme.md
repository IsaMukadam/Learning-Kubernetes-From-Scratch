# Creating directory structure for learning

```sh
mkdir exercises
```

```sh
New-Item -ItemType Directory -Path "exercises\2-intermediate","exercises\2-intermediate\pods","exercises\2-intermediate\deployments","exercises\2-intermediate\services","exercises\2-intermediate\config","exercises\3-advanced","exercises\3-advanced\pods","exercises\3-advanced\deployments","exercises\3-advanced\services","exercises\3-advanced\config"
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