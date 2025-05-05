package main

import (
    "fmt"
    "sync"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkflowTemplate defines the structure for workflow templates
type WorkflowTemplate struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec             WorkflowTemplateSpec `json:"spec"`
}

type WorkflowTemplateSpec struct {
    Version     int          `json:"version"`
    Description string       `json:"description,omitempty"`
    Parameters  []Parameter  `json:"parameters,omitempty"`
    Tasks       []TaskTemplate `json:"tasks"`
}

type Parameter struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Default     string `json:"default,omitempty"`
}

type TaskTemplate struct {
    Name            string                 `json:"name"`
    TaskType        string                 `json:"taskType"`
    RetryCount      int                   `json:"retryCount,omitempty"`
    RetryLogic      string                `json:"retryLogic,omitempty"`
    TimeoutSeconds  int                   `json:"timeoutSeconds,omitempty"`
    InputTemplate   map[string]interface{} `json:"inputTemplate,omitempty"`
    Optional        bool                  `json:"optional,omitempty"`
    DependsOn      []string              `json:"dependsOn,omitempty"`
}

// TemplateManager handles workflow template operations
type TemplateManager struct {
    templates     map[string]map[int]*WorkflowTemplate // name -> version -> template
    mutex         sync.RWMutex
}

func NewTemplateManager() *TemplateManager {
    return &TemplateManager{
        templates: make(map[string]map[int]*WorkflowTemplate),
    }
}

func (tm *TemplateManager) AddTemplate(template *WorkflowTemplate) error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()

    name := template.Name
    version := template.Spec.Version

    // Initialize version map if needed
    if _, exists := tm.templates[name]; !exists {
        tm.templates[name] = make(map[int]*WorkflowTemplate)
    }

    // Check if version already exists
    if _, exists := tm.templates[name][version]; exists {
        return fmt.Errorf("template %s version %d already exists", name, version)
    }

    tm.templates[name][version] = template
    return nil
}

func (tm *TemplateManager) GetTemplate(name string, version int) (*WorkflowTemplate, error) {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()

    if versions, exists := tm.templates[name]; exists {
        if template, exists := versions[version]; exists {
            return template, nil
        }
        return nil, fmt.Errorf("version %d not found for template %s", version, name)
    }
    return nil, fmt.Errorf("template %s not found", name)
}

func (tm *TemplateManager) GetLatestVersion(name string) (*WorkflowTemplate, error) {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()

    if versions, exists := tm.templates[name]; exists {
        var latestVersion int
        for v := range versions {
            if v > latestVersion {
                latestVersion = v
            }
        }
        return versions[latestVersion], nil
    }
    return nil, fmt.Errorf("template %s not found", name)
}

func (tm *TemplateManager) CreateWorkflowFromTemplate(template *WorkflowTemplate, params map[string]string) (*Workflow, error) {
    // Validate required parameters
    for _, param := range template.Spec.Parameters {
        if param.Required {
            if _, exists := params[param.Name]; !exists {
                if param.Default == "" {
                    return nil, fmt.Errorf("required parameter %s not provided", param.Name)
                }
                params[param.Name] = param.Default
            }
        }
    }

    // Create new workflow from template
    workflow := &Workflow{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "netflix.com/v1",
            Kind:       "Workflow",
        },
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: template.Name + "-",
            Labels: map[string]string{
                "template":        template.Name,
                "templateVersion": fmt.Sprintf("%d", template.Spec.Version),
            },
        },
        Spec: WorkflowSpec{
            Name:           template.Name,
            Version:        template.Spec.Version,
            TimeoutPolicy:  "TIME_OUT", // Default timeout policy
            TimeoutSeconds: 3600,       // Default timeout
            Tasks:         make([]Task, len(template.Spec.Tasks)),
        },
    }

    // Convert task templates to tasks
    for i, taskTemplate := range template.Spec.Tasks {
        workflow.Spec.Tasks[i] = Task{
            Name:           taskTemplate.Name,
            TaskType:       taskTemplate.TaskType,
            RetryCount:    taskTemplate.RetryCount,
            RetryLogic:    taskTemplate.RetryLogic,
            TimeoutSeconds: taskTemplate.TimeoutSeconds,
            Optional:      taskTemplate.Optional,
            InputParameters: tm.resolveInputParameters(taskTemplate.InputTemplate, params),
        }
    }

    return workflow, nil
}

func (tm *TemplateManager) resolveInputParameters(inputTemplate map[string]interface{}, params map[string]string) map[string]interface{} {
    resolved := make(map[string]interface{})
    for k, v := range inputTemplate {
        switch val := v.(type) {
        case string:
            if param, exists := params[val]; exists {
                resolved[k] = param
            } else {
                resolved[k] = val
            }
        default:
            resolved[k] = v
        }
    }
    return resolved
}

func (tm *TemplateManager) ListTemplates() []WorkflowTemplate {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()

    var templates []WorkflowTemplate
    for _, versions := range tm.templates {
        for _, template := range versions {
            templates = append(templates, *template)
        }
    }
    return templates
}

func (tm *TemplateManager) DeleteTemplate(name string, version int) error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()

    if versions, exists := tm.templates[name]; exists {
        if _, exists := versions[version]; exists {
            delete(versions, version)
            if len(versions) == 0 {
                delete(tm.templates, name)
            }
            return nil
        }
        return fmt.Errorf("version %d not found for template %s", version, name)
    }
    return fmt.Errorf("template %s not found", name)
}