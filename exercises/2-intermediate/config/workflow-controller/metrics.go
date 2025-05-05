package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsCollector struct {
    taskExecutionTime    *prometheus.HistogramVec
    taskExecutionCount   *prometheus.CounterVec
    workflowDuration    *prometheus.HistogramVec
    errorCount          *prometheus.CounterVec
    activeWorkflows     prometheus.Gauge
    taskQueueSize       *prometheus.GaugeVec
    taskRetries         *prometheus.CounterVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        taskExecutionTime: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "workflow_task_execution_duration_seconds",
                Help: "Time taken to execute workflow tasks",
                Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
            },
            []string{"task_type", "workflow_name", "task_name"},
        ),
        
        taskExecutionCount: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "workflow_task_execution_total",
                Help: "Total number of task executions",
            },
            []string{"task_type", "status", "workflow_name"},
        ),

        workflowDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "workflow_execution_duration_seconds",
                Help: "Time taken to execute complete workflows",
                Buckets: prometheus.ExponentialBuckets(1, 2, 12),
            },
            []string{"workflow_name", "status"},
        ),

        errorCount: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "workflow_errors_total",
                Help: "Total number of workflow errors",
            },
            []string{"error_type", "workflow_name"},
        ),

        activeWorkflows: promauto.NewGauge(
            prometheus.GaugeOpts{
                Name: "workflow_active_count",
                Help: "Number of currently active workflows",
            },
        ),

        taskQueueSize: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "workflow_task_queue_size",
                Help: "Current size of task queues",
            },
            []string{"task_type"},
        ),

        taskRetries: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "workflow_task_retries_total",
                Help: "Total number of task retries",
            },
            []string{"task_type", "workflow_name", "task_name"},
        ),
    }
}

func (mc *MetricsCollector) RecordTaskExecution(taskType, status string, duration float64, workflowName, taskName string) {
    mc.taskExecutionTime.WithLabelValues(taskType, workflowName, taskName).Observe(duration)
    mc.taskExecutionCount.WithLabelValues(taskType, status, workflowName).Inc()
}

func (mc *MetricsCollector) RecordWorkflowExecution(workflowName, status string, duration float64) {
    mc.workflowDuration.WithLabelValues(workflowName, status).Observe(duration)
}

func (mc *MetricsCollector) RecordError(errorType, workflowName string) {
    mc.errorCount.WithLabelValues(errorType, workflowName).Inc()
}

func (mc *MetricsCollector) UpdateActiveWorkflows(count float64) {
    mc.activeWorkflows.Set(count)
}

func (mc *MetricsCollector) UpdateTaskQueueSize(taskType string, size float64) {
    mc.taskQueueSize.WithLabelValues(taskType).Set(size)
}

func (mc *MetricsCollector) RecordTaskRetry(taskType, workflowName, taskName string) {
    mc.taskRetries.WithLabelValues(taskType, workflowName, taskName).Inc()
}