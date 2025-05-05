package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

// Game represents our custom resource
type Game struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec             GameSpec `json:"spec"`
}

// GameSpec defines our game server configuration
type GameSpec struct {
    GameName string `json:"gameName"`
    Players  int32  `json:"players"`
    Port     int32  `json:"port"`
}

func main() {
    // Setup kubernetes config
    kubeconfig := flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "")
    flag.Parse()

    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
        fmt.Printf("Error building config: %v\n", err)
        os.Exit(1)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Printf("Error creating clientset: %v\n", err)
        os.Exit(1)
    }

    // Example game server
    game := &Game{
        ObjectMeta: metav1.ObjectMeta{
            Name: "minecraft-server",
            Namespace: "default",
        },
        Spec: GameSpec{
            GameName: "minecraft",
            Players: 20,
            Port: 25565,
        },
    }

    // Create deployment for the game server
    deployment := createGameDeployment(game)
    _, err = clientset.AppsV1().Deployments(game.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
    if err != nil {
        fmt.Printf("Error creating deployment: %v\n", err)
    }

    // Create service to expose the game server
    service := createGameService(game)
    _, err = clientset.CoreV1().Services(game.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
    if err != nil {
        fmt.Printf("Error creating service: %v\n", err)
    }
}

func createGameDeployment(game *Game) *appsv1.Deployment {
    replicas := int32(1)
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: game.Name,
            Namespace: game.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "game": game.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "game": game.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  game.Spec.GameName,
                            Image: fmt.Sprintf("%s:latest", game.Spec.GameName),
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: game.Spec.Port,
                                },
                            },
                            Env: []corev1.EnvVar{
                                {
                                    Name: "MAX_PLAYERS",
                                    Value: fmt.Sprintf("%d", game.Spec.Players),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

func createGameService(game *Game) *corev1.Service {
    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name: game.Name,
            Namespace: game.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Type: corev1.ServiceTypeNodePort,
            Selector: map[string]string{
                "game": game.Name,
            },
            Ports: []corev1.ServicePort{
                {
                    Port: game.Spec.Port,
                    TargetPort: intstr.FromInt(int(game.Spec.Port)),
                },
            },
        },
    }
}