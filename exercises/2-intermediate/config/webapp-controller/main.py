from dataclasses import dataclass
from typing import List, Optional
from kubernetes import client, config
from pathlib import Path
import argparse
import os

@dataclass
class Resources:
    cpu: Optional[str] = None
    memory: Optional[str] = None

@dataclass
class ResourceRequests:
    limits: Optional[Resources] = None
    requests: Optional[Resources] = None

@dataclass
class SSLConfig:
    enabled: bool = False
    secret_name: Optional[str] = None

@dataclass
class WebAppSpec:
    image: str
    port: int
    replicas: int
    domains: Optional[List[str]] = None
    ssl: Optional[SSLConfig] = None
    resources: Optional[ResourceRequests] = None

@dataclass
class WebApp:
    name: str
    namespace: str
    spec: WebAppSpec

def create_deployment(webapp: WebApp) -> client.V1Deployment:
    container = client.V1Container(
        name=webapp.name,
        image=webapp.spec.image,
        ports=[client.V1ContainerPort(container_port=webapp.spec.port)]
    )

    # Add resource requirements if specified
    if webapp.spec.resources:
        limits = {}
        requests = {}
        
        if webapp.spec.resources.limits:
            if webapp.spec.resources.limits.cpu:
                limits['cpu'] = webapp.spec.resources.limits.cpu
            if webapp.spec.resources.limits.memory:
                limits['memory'] = webapp.spec.resources.limits.memory
                
        if webapp.spec.resources.requests:
            if webapp.spec.resources.requests.cpu:
                requests['cpu'] = webapp.spec.resources.requests.cpu
            if webapp.spec.resources.requests.memory:
                requests['memory'] = webapp.spec.resources.requests.memory
                
        container.resources = client.V1ResourceRequirements(
            limits=limits,
            requests=requests
        )

    return client.V1Deployment(
        metadata=client.V1ObjectMeta(
            name=webapp.name,
            namespace=webapp.namespace
        ),
        spec=client.V1DeploymentSpec(
            replicas=webapp.spec.replicas,
            selector=client.V1LabelSelector(
                match_labels={"app": webapp.name}
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels={"app": webapp.name}
                ),
                spec=client.V1PodSpec(
                    containers=[container]
                )
            )
        )
    )

def create_service(webapp: WebApp) -> client.V1Service:
    return client.V1Service(
        metadata=client.V1ObjectMeta(
            name=webapp.name,
            namespace=webapp.namespace
        ),
        spec=client.V1ServiceSpec(
            selector={"app": webapp.name},
            ports=[client.V1ServicePort(
                port=webapp.spec.port,
                target_port=webapp.spec.port
            )]
        )
    )

def create_ingress(webapp: WebApp) -> client.V1Ingress:
    path_type = "Prefix"
    rules = []
    
    for domain in webapp.spec.domains:
        rules.append(client.V1IngressRule(
            host=domain,
            http=client.V1HTTPIngressRuleValue(
                paths=[client.V1HTTPIngressPath(
                    path="/",
                    path_type=path_type,
                    backend=client.V1IngressBackend(
                        service=client.V1IngressServiceBackend(
                            name=webapp.name,
                            port=client.V1ServiceBackendPort(
                                number=webapp.spec.port
                            )
                        )
                    )
                )]
            )
        ))

    tls = None
    if webapp.spec.ssl and webapp.spec.ssl.enabled:
        tls = [client.V1IngressTLS(
            hosts=webapp.spec.domains,
            secret_name=webapp.spec.ssl.secret_name
        )]

    return client.V1Ingress(
        metadata=client.V1ObjectMeta(
            name=webapp.name,
            namespace=webapp.namespace
        ),
        spec=client.V1IngressSpec(
            tls=tls,
            rules=rules
        )
    )

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--kubeconfig",
        default=os.path.join(str(Path.home()), ".kube", "config"),
        help="Path to kubeconfig file"
    )
    args = parser.parse_args()

    # Load kubernetes configuration
    config.load_kube_config(args.kubeconfig)

    # Create API clients
    apps_v1 = client.AppsV1Api()
    core_v1 = client.CoreV1Api()
    networking_v1 = client.NetworkingV1Api()

    # Example WebApp
    webapp = WebApp(
        name="example-webapp",
        namespace="default",
        spec=WebAppSpec(
            image="nginx:1.14",
            port=80,
            replicas=3,
            ssl=SSLConfig(
                enabled=True,
                secret_name="webapp-tls"
            ),
            resources=ResourceRequests(
                limits=Resources(
                    cpu="500m",
                    memory="512Mi"
                ),
                requests=Resources(
                    cpu="250m",
                    memory="256Mi"
                )
            )
        )
    )

    try:
        # Create deployment
        deployment = create_deployment(webapp)
        apps_v1.create_namespaced_deployment(
            namespace=webapp.namespace,
            body=deployment
        )
        print(f"Created deployment: {webapp.name}")

        # Create service
        service = create_service(webapp)
        core_v1.create_namespaced_service(
            namespace=webapp.namespace,
            body=service
        )
        print(f"Created service: {webapp.name}")

        # Create ingress if domains are specified and SSL is enabled
        if webapp.spec.domains and webapp.spec.ssl and webapp.spec.ssl.enabled:
            ingress = create_ingress(webapp)
            networking_v1.create_namespaced_ingress(
                namespace=webapp.namespace,
                body=ingress
            )
            print(f"Created ingress: {webapp.name}")

    except client.rest.ApiException as e:
        print(f"Kubernetes API error: {e}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()