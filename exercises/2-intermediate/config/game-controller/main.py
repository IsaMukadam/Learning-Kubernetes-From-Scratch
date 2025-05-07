from kubernetes import client, config
from kubernetes.client import V1Deployment, V1Service
from kubernetes.client.rest import ApiException
from dataclasses import dataclass
import os
import argparse
from pathlib import Path

@dataclass
class GameSpec:
    game_name: str
    players: int
    port: int

@dataclass
class Game:
    name: str
    namespace: str
    spec: GameSpec

def create_game_deployment(game: Game) -> V1Deployment:
    return client.V1Deployment(
        metadata=client.V1ObjectMeta(
            name=game.name,
            namespace=game.namespace
        ),
        spec=client.V1DeploymentSpec(
            replicas=1,
            selector=client.V1LabelSelector(
                match_labels={"game": game.name}
            ),
            template=client.V1PodTemplateSpec(
                metadata=client.V1ObjectMeta(
                    labels={"game": game.name}
                ),
                spec=client.V1PodSpec(
                    containers=[
                        client.V1Container(
                            name=game.spec.game_name,
                            image=f"{game.spec.game_name}:latest",
                            ports=[
                                client.V1ContainerPort(
                                    container_port=game.spec.port
                                )
                            ],
                            env=[
                                client.V1EnvVar(
                                    name="MAX_PLAYERS",
                                    value=str(game.spec.players)
                                )
                            ]
                        )
                    ]
                )
            )
        )
    )

def create_game_service(game: Game) -> V1Service:
    return client.V1Service(
        metadata=client.V1ObjectMeta(
            name=game.name,
            namespace=game.namespace
        ),
        spec=client.V1ServiceSpec(
            type="NodePort",
            selector={"game": game.name},
            ports=[
                client.V1ServicePort(
                    port=game.spec.port,
                    target_port=game.spec.port
                )
            ]
        )
    )

def main():
    # Parse arguments
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--kubeconfig",
        default=os.path.join(str(Path.home()), ".kube", "config"),
        help="Path to kubeconfig file"
    )
    args = parser.parse_args()

    # Load kubernetes configuration
    try:
        config.load_kube_config(args.kubeconfig)
    except Exception as e:
        print(f"Error loading kubeconfig: {e}")
        exit(1)

    # Create API clients
    apps_v1 = client.AppsV1Api()
    core_v1 = client.CoreV1Api()

    # Create game instance
    game = Game(
        name="minecraft-server",
        namespace="default",
        spec=GameSpec(
            game_name="minecraft",
            players=20,
            port=25565
        )
    )

    # Create deployment
    try:
        deployment = create_game_deployment(game)
        apps_v1.create_namespaced_deployment(
            namespace=game.namespace,
            body=deployment
        )
        print(f"Created deployment: {game.name}")
    except ApiException as e:
        print(f"Error creating deployment: {e}")

    # Create service
    try:
        service = create_game_service(game)
        core_v1.create_namespaced_service(
            namespace=game.namespace,
            body=service
        )
        print(f"Created service: {game.name}")
    except ApiException as e:
        print(f"Error creating service: {e}")

if __name__ == "__main__":
    main()