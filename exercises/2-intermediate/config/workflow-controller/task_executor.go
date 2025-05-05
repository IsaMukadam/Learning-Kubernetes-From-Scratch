package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/lambda"
    "k8s.io/client-go/kubernetes"
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/utils/pointer"
)

type DefaultTaskExecutor struct {
    httpClient    *http.Client
    lambdaClient  *lambda.Client
    kubeClient    kubernetes.Interface
    metricsCollector *MetricsCollector
}

func NewDefaultTaskExecutor(kubeClient kubernetes.Interface) (*DefaultTaskExecutor, error) {
    // Configure AWS Lambda client
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("unable to load AWS config: %v", err)
    }

    return &DefaultTaskExecutor{
        httpClient: &http.Client{
            Timeout: time.Second * 30,
        },
        lambdaClient: lambda.NewFromConfig(cfg),
        kubeClient: kubeClient,
        metricsCollector: NewMetricsCollector(),
    }, nil
}

func (e *DefaultTaskExecutor) ExecuteTask(task Task, workflow *Workflow) error {
    startTime := time.Now()
    var err error

    switch task.TaskType {
    case "KUBERNETES_JOB":
        err = e.executeKubernetesJob(task, workflow)
    case "HTTP":
        err = e.executeHTTPTask(task)
    case "LAMBDA":
        err = e.executeLambdaTask(task)
    case "SIMPLE":
        err = e.executeSimpleTask(task)
    case "FORK_JOIN":
        err = e.executeForkJoinTask(task, workflow)
    default:
        err = fmt.Errorf("unsupported task type: %s", task.TaskType)
    }

    duration := time.Since(startTime).Seconds()
    status := "success"
    if err != nil {
        status = "failed"
        e.metricsCollector.RecordError("task_execution_error")
    }
    e.metricsCollector.RecordTaskExecution(task.TaskType, status, duration)
    return err
}

func (e *DefaultTaskExecutor) executeHTTPTask(task Task) error {
    params, ok := task.InputParameters["http"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid HTTP parameters for task %s", task.Name)
    }

    // Extract HTTP parameters
    uri := params["uri"].(string)
    method := params["method"].(string)
    contentType := params["contentType"].(string)

    var body []byte
    if params["body"] != nil {
        body, _ = json.Marshal(params["body"])
    }

    // Create and execute request
    req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request: %v", err)
    }

    req.Header.Set("Content-Type", contentType)

    // Execute with retry logic
    var resp *http.Response
    for retries := 0; retries <= task.RetryCount; retries++ {
        resp, err = e.httpClient.Do(req)
        if err == nil && resp.StatusCode < 500 {
            break
        }
        
        if retries < task.RetryCount {
            delay := e.calculateRetryDelay(task, retries)
            time.Sleep(delay)
        }
    }

    if err != nil {
        return fmt.Errorf("HTTP request failed after %d retries: %v", task.RetryCount, err)
    }

    if resp.StatusCode >= 400 {
        return fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
    }

    return nil
}

func (e *DefaultTaskExecutor) executeLambdaTask(task Task) error {
    params, err := json.Marshal(task.InputParameters)
    if err != nil {
        return fmt.Errorf("failed to marshal Lambda parameters: %v", err)
    }

    input := &lambda.InvokeInput{
        FunctionName: aws.String(task.Name),
        Payload:      params,
    }

    // Execute with retry logic
    var output *lambda.InvokeOutput
    for retries := 0; retries <= task.RetryCount; retries++ {
        output, err = e.lambdaClient.Invoke(context.Background(), input)
        if err == nil && output.FunctionError == nil {
            break
        }
        
        if retries < task.RetryCount {
            delay := e.calculateRetryDelay(task, retries)
            time.Sleep(delay)
        }
    }

    if err != nil {
        return fmt.Errorf("Lambda invocation failed after %d retries: %v", task.RetryCount, err)
    }

    if output.FunctionError != nil {
        return fmt.Errorf("Lambda function returned error: %s", *output.FunctionError)
    }

    return nil
}

func (e *DefaultTaskExecutor) executeSimpleTask(task Task) error {
    // Simple tasks are implemented by the user and registered with the controller
    // Here we'll just log and succeed
    log.Printf("Executing simple task: %s with parameters: %v", task.Name, task.InputParameters)
    return nil
}

func (e *DefaultTaskExecutor) executeForkJoinTask(task Task, workflow *Workflow) error {
    forkTasks, ok := task.InputParameters["forkTasks"].([]interface{})
    if !ok {
        return fmt.Errorf("invalid fork tasks configuration")
    }

    var wg sync.WaitGroup
    errors := make(chan error, len(forkTasks))

    for _, t := range forkTasks {
        wg.Add(1)
        go func(taskData interface{}) {
            defer wg.Done()

            // Convert task data to Task struct
            taskBytes, _ := json.Marshal(taskData)
            var subTask Task
            if err := json.Unmarshal(taskBytes, &subTask); err != nil {
                errors <- fmt.Errorf("failed to parse fork task: %v", err)
                return
            }

            // Execute the subtask
            if err := e.ExecuteTask(subTask, workflow); err != nil {
                errors <- fmt.Errorf("fork task %s failed: %v", subTask.Name, err)
            }
        }(t)
    }

    // Wait for all tasks to complete
    wg.Wait()
    close(errors)

    // Collect any errors
    var errs []string
    for err := range errors {
        errs = append(errs, err.Error())
    }

    if len(errs) > 0 {
        return fmt.Errorf("fork-join task errors: %v", errs)
    }

    return nil
}

func (e *DefaultTaskExecutor) executeKubernetesJob(task Task, workflow *Workflow) error {
    params, ok := task.InputParameters["job"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid Kubernetes Job parameters for task %s", task.Name)
    }

    // Create the Job object
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-%s", workflow.Name, task.Name),
            Namespace: workflow.Namespace,
            Labels: map[string]string{
                "workflow": workflow.Name,
                "task":    task.Name,
            },
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyOnFailure,
                    Containers: []corev1.Container{
                        {
                            Name:    task.Name,
                            Image:   params["image"].(string),
                            Command: interfaceSliceToStringSlice(params["command"].([]interface{})),
                            Env:     createEnvVarsFromMap(params["env"].(map[string]interface{})),
                        },
                    },
                },
            },
            BackoffLimit: pointer.Int32Ptr(task.RetryCount),
        },
    }

    // Create the job
    createdJob, err := e.kubeClient.BatchV1().Jobs(workflow.Namespace).Create(
        context.Background(),
        job,
        metav1.CreateOptions{},
    )
    if err != nil {
        return fmt.Errorf("failed to create job: %v", err)
    }

    // Watch job completion
    return e.waitForJobCompletion(createdJob.Namespace, createdJob.Name, task)
}

func (e *DefaultTaskExecutor) waitForJobCompletion(namespace, name string, task Task) error {
    watch, err := e.kubeClient.BatchV1().Jobs(namespace).Watch(
        context.Background(),
        metav1.ListOptions{
            FieldSelector: fmt.Sprintf("metadata.name=%s", name),
        },
    )
    if err != nil {
        return fmt.Errorf("failed to watch job: %v", err)
    }
    defer watch.Stop()

    timeout := time.After(time.Duration(task.TimeoutSeconds) * time.Second)
    for {
        select {
        case event := <-watch.ResultChan():
            job, ok := event.Object.(*batchv1.Job)
            if !ok {
                continue
            }

            if job.Status.Succeeded > 0 {
                return nil
            }
            if job.Status.Failed > 0 {
                return fmt.Errorf("job failed")
            }

        case <-timeout:
            return fmt.Errorf("job timed out")
        }
    }
}

func interfaceSliceToStringSlice(slice []interface{}) []string {
    result := make([]string, len(slice))
    for i, v := range slice {
        result[i] = v.(string)
    }
    return result
}

func createEnvVarsFromMap(envMap map[string]interface{}) []corev1.EnvVar {
    envVars := make([]corev1.EnvVar, 0)
    for k, v := range envMap {
        envVars = append(envVars, corev1.EnvVar{
            Name:  k,
            Value: fmt.Sprintf("%v", v),
        })
    }
    return envVars
}

func (e *DefaultTaskExecutor) calculateRetryDelay(task Task, attempt int) time.Duration {
    baseDelay := time.Second

    switch task.RetryLogic {
    case "EXPONENTIAL_BACKOFF":
        // Exponential backoff with jitter
        delay := baseDelay * time.Duration(1<<uint(attempt))
        jitter := time.Duration(rand.Float64() * float64(delay/2))
        return delay + jitter
    default: // "FIXED"
        return baseDelay
    }
}