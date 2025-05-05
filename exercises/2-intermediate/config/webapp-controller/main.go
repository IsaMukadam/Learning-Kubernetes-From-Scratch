package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "path/filepath"
    
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/util/intstr"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

type WebApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec             WebAppSpec   `json:"spec"`
}

type WebAppSpec struct {
    Image     string             `json:"image"`
    Port      int32             `json:"port"`
    Replicas  int32             `json:"replicas"`
    Domains   []string          `json:"domains,omitempty"`
    SSL       *SSLConfig        `json:"ssl,omitempty"`
    Resources *ResourceRequests `json:"resources,omitempty"`
}

type SSLConfig struct {
    Enabled    bool   `json:"enabled"`
    SecretName string `json:"secretName,omitempty"`
}

type ResourceRequests struct {
    Limits   *Resources `json:"limits,omitempty"`
    Requests *Resources `json:"requests,omitempty"`
}

type Resources struct {
    CPU    string `json:"cpu,omitempty"`
    Memory string `json:"memory,omitempty"`
}

func main() {
    kubeconfig := flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "kubeconfig file")
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

    // Example WebApp
    webapp := &WebApp{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "example-webapp",
            Namespace: "default",
        },
        Spec: WebAppSpec{
            Image:    "nginx:1.14",
            Port:     80,
            Replicas: 3,
            SSL: &SSLConfig{
                Enabled:    true,
                SecretName: "webapp-tls",
            },
            Resources: &ResourceRequests{
                Limits: &Resources{
                    CPU:    "500m",
                    Memory: "512Mi",
                },
                Requests: &Resources{
                    CPU:    "250m",
                    Memory: "256Mi",
                },
            },
        },
    }

    // Create deployment
    deployment := createDeployment(webapp)
    _, err = clientset.AppsV1().Deployments(webapp.Namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
    if err != nil {
        fmt.Printf("Error creating deployment: %v\n", err)
    }

    // Create service
    service := createService(webapp)
    _, err = clientset.CoreV1().Services(webapp.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
    if err != nil {
        fmt.Printf("Error creating service: %v\n", err)
    }

    if len(webapp.Spec.Domains) > 0 && webapp.Spec.SSL != nil && webapp.Spec.SSL.Enabled {
        // Create Ingress with TLS
        ingress := createIngress(webapp)
        _, err = clientset.NetworkingV1().Ingresses(webapp.Namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
        if err != nil {
            fmt.Printf("Error creating ingress: %v\n", err)
        }
    }
}

func createDeployment(webapp *WebApp) *appsv1.Deployment {
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      webapp.Name,
            Namespace: webapp.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &webapp.Spec.Replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": webapp.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": webapp.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  webapp.Name,
                            Image: webapp.Spec.Image,
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: webapp.Spec.Port,
                                },
                            },
                            Resources: createResourceRequirements(webapp.Spec.Resources),
                        },
                    },
                },
            },
        },
    }
}

func createService(webapp *WebApp) *corev1.Service {
    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      webapp.Name,
            Namespace: webapp.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{
                "app": webapp.Name,
            },
            Ports: []corev1.ServicePort{
                {
                    Port:       webapp.Spec.Port,
                    TargetPort: intstr.FromInt(int(webapp.Spec.Port)),
                },
            },
        },
    }
}

func createIngress(webapp *WebApp) *networkingv1.Ingress {
    pathType := networkingv1.PathTypePrefix
    ingressRules := make([]networkingv1.IngressRule, 0, len(webapp.Spec.Domains))
    
    for _, domain := range webapp.Spec.Domains {
        ingressRules = append(ingressRules, networkingv1.IngressRule{
            Host: domain,
            IngressRuleValue: networkingv1.IngressRuleValue{
                HTTP: &networkingv1.HTTPIngressRuleValue{
                    Paths: []networkingv1.HTTPIngressPath{
                        {
                            Path:     "/",
                            PathType: &pathType,
                            Backend: networkingv1.IngressBackend{
                                Service: &networkingv1.IngressServiceBackend{
                                    Name: webapp.Name,
                                    Port: networkingv1.ServiceBackendPort{
                                        Number: webapp.Spec.Port,
                                    },
                                },
                            },
                        },
                    },
                },
            },
        })
    }

    var tls []networkingv1.IngressTLS
    if webapp.Spec.SSL != nil && webapp.Spec.SSL.Enabled {
        tls = []networkingv1.IngressTLS{
            {
                Hosts:      webapp.Spec.Domains,
                SecretName: webapp.Spec.SSL.SecretName,
            },
        }
    }

    return &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:      webapp.Name,
            Namespace: webapp.Namespace,
        },
        Spec: networkingv1.IngressSpec{
            TLS:   tls,
            Rules: ingressRules,
        },
    }
}

func createResourceRequirements(resources *ResourceRequests) corev1.ResourceRequirements {
    if resources == nil {
        return corev1.ResourceRequirements{}
    }

    reqs := corev1.ResourceRequirements{}
    
    if resources.Limits != nil {
        reqs.Limits = corev1.ResourceList{}
        if resources.Limits.CPU != "" {
            reqs.Limits[corev1.ResourceCPU] = resource.MustParse(resources.Limits.CPU)
        }
        if resources.Limits.Memory != "" {
            reqs.Limits[corev1.ResourceMemory] = resource.MustParse(resources.Limits.Memory)
        }
    }

    if resources.Requests != nil {
        reqs.Requests = corev1.ResourceList{}
        if resources.Requests.CPU != "" {
            reqs.Requests[corev1.ResourceCPU] = resource.MustParse(resources.Requests.CPU)
        }
        if resources.Requests.Memory != "" {
            reqs.Requests[corev1.ResourceMemory] = resource.MustParse(resources.Requests.Memory)
        }
    }

    return reqs
}