# Kubernetes ConfigMaps and Secrets Tutorial

This tutorial demonstrates how to use ConfigMaps and Secrets in Kubernetes to manage configuration and sensitive data.

## Overview

In this lesson, we'll cover:
1. Creating and using ConfigMaps for configuration data
2. Creating and using Secrets for sensitive information
3. Mounting ConfigMaps as files
4. Using Secrets as environment variables

## ConfigMap Example

Our ConfigMap contains HTML content that will be served by an nginx pod:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-config
data:
  index.html: |
    <!DOCTYPE html>
    <html>
    <body>
      <h1>Hello from ConfigMap!</h1>
      <p>This content is loaded from a ConfigMap</p>
    </body>
    </html>
```

Key points about ConfigMaps:
- Used for non-sensitive configuration data
- Can store multiple key-value pairs
- Can be mounted as files or used as environment variables
- Updates to ConfigMaps can be picked up by pods (with some configuration)

## Secret Example

Our Secret contains a message that will be exposed as an environment variable:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nginx-secret
type: Opaque
stringData:  # Using stringData instead of data so we don't need to base64 encode
  message: "This message comes from a Secret!"
```

Key points about Secrets:
- Used for sensitive information (passwords, tokens, keys)
- Values are base64 encoded in the cluster
- Can be mounted as files or used as environment variables
- Kubernetes automatically handles encoding/decoding

## Deployment Configuration

Our deployment shows how to use both ConfigMap and Secret:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-configured
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-configured
  template:
    metadata:
      labels:
        app: nginx-configured
    spec:
      containers:
      - name: nginx
        image: nginx:1.14
        ports:
        - containerPort: 80
        volumeMounts:
        - name: config-volume    # Mount ConfigMap as files
          mountPath: /usr/share/nginx/html/
        env:
        - name: SECRET_MESSAGE   # Use Secret as environment variable
          valueFrom:
            secretKeyRef:
              name: nginx-secret
              key: message
      volumes:
      - name: config-volume      # Define ConfigMap volume
        configMap:
          name: nginx-config
```

Key points about the deployment:
1. ConfigMap mounted as files:
   - Creates files in the container for each key in the ConfigMap
   - Files are created at the specified mountPath
   - In this case, index.html will be served by nginx

2. Secret used as environment variable:
   - Creates environment variable SECRET_MESSAGE
   - Value is automatically decoded from base64
   - Available to the application as a normal environment variable

## Service Configuration

To access our configured nginx server:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-configured-service
spec:
  type: NodePort
  selector:
    app: nginx-configured
  ports:
    - port: 80
      targetPort: 80
      nodePort: 30091
```

## Testing the Configuration

You can verify the configuration is working by:

1. Checking the mounted ConfigMap content:
```bash
kubectl exec <pod-name> -- cat /usr/share/nginx/html/index.html
```

2. Verifying the Secret environment variable:
```bash
kubectl exec <pod-name> -- env | grep SECRET_MESSAGE
```

3. Accessing the web server:
```bash
kubectl port-forward svc/nginx-configured-service 8080:80
# Then open http://localhost:8080 in a browser
```

## Best Practices

1. ConfigMaps:
   - Use for non-sensitive configuration data
   - Keep configurations modular and specific to their use
   - Consider using them for configuration files, scripts, or static content

2. Secrets:
   - Never commit actual secrets to version control
   - Use for sensitive data only
   - Consider using external secret management solutions for production
   - Use stringData for better readability when creating Secrets

3. General:
   - Use meaningful names for ConfigMaps and Secrets
   - Document the purpose and usage of each configuration
   - Consider how updates will be handled
   - Use labels and annotations effectively