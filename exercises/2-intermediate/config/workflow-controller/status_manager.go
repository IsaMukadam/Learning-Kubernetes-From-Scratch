package main

import (
    "fmt"
    "time"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
    PhaseInitializing = "Initializing"
    PhaseRunning      = "Running"
    PhaseCompleted    = "Completed"
    PhaseFailed      = "Failed"
    PhaseTimedOut    = "TimedOut"

    ConditionTypeStarted    = "Started"
    ConditionTypeCompleted  = "Completed"
    ConditionTypeFailed    = "Failed"
)

type StatusManager struct {
    workflow *Workflow
}

func NewStatusManager(workflow *Workflow) *StatusManager {
    return &StatusManager{
        workflow: workflow,
    }
}

func (sm *StatusManager) InitializeWorkflow() {
    now := metav1.Now()
    sm.workflow.Status = WorkflowStatus{
        Phase:     PhaseInitializing,
        StartTime: now,
        Tasks:     make([]TaskStatus, 0),
        Conditions: []Condition{
            {
                Type:               ConditionTypeStarted,
                Status:            "True",
                LastTransitionTime: now,
                Reason:            "WorkflowStarted",
                Message:           "Workflow initialization started",
            },
        },
    }
}

func (sm *StatusManager) StartTask(task Task) {
    now := metav1.Now()
    taskStatus := TaskStatus{
        Name:      task.Name,
        Phase:     "Running",
        StartTime: now,
    }
    sm.workflow.Status.Tasks = append(sm.workflow.Status.Tasks, taskStatus)
    sm.workflow.Status.Phase = PhaseRunning
}

func (sm *StatusManager) CompleteTask(taskName string, err error) {
    now := metav1.Now()
    for i, task := range sm.workflow.Status.Tasks {
        if task.Name == taskName {
            sm.workflow.Status.Tasks[i].FinishTime = now
            if err != nil {
                sm.workflow.Status.Tasks[i].Phase = "Failed"
                sm.workflow.Status.Tasks[i].Error = err.Error()
            } else {
                sm.workflow.Status.Tasks[i].Phase = "Completed"
            }
            break
        }
    }

    // Update workflow status based on task completion
    sm.updateWorkflowStatus()
}

func (sm *StatusManager) updateWorkflowStatus() {
    // Check if all tasks are completed
    allCompleted := true
    anyFailed := false

    for _, task := range sm.workflow.Status.Tasks {
        if task.Phase == "Failed" {
            anyFailed = true
            break
        }
        if task.Phase != "Completed" {
            allCompleted = false
        }
    }

    now := metav1.Now()
    if anyFailed {
        sm.workflow.Status.Phase = PhaseFailed
        sm.workflow.Status.Conditions = append(sm.workflow.Status.Conditions, Condition{
            Type:               ConditionTypeFailed,
            Status:            "True",
            LastTransitionTime: now,
            Reason:            "TaskFailed",
            Message:           "One or more tasks failed",
        })
    } else if allCompleted {
        sm.workflow.Status.Phase = PhaseCompleted
        sm.workflow.Status.Conditions = append(sm.workflow.Status.Conditions, Condition{
            Type:               ConditionTypeCompleted,
            Status:            "True",
            LastTransitionTime: now,
            Reason:            "WorkflowCompleted",
            Message:           "All tasks completed successfully",
        })
    }
}

func (sm *StatusManager) CheckTimeout() bool {
    if sm.workflow.Spec.TimeoutSeconds == 0 {
        return false
    }

    elapsed := time.Since(sm.workflow.Status.StartTime.Time)
    if int(elapsed.Seconds()) > sm.workflow.Spec.TimeoutSeconds {
        now := metav1.Now()
        sm.workflow.Status.Phase = PhaseTimedOut
        sm.workflow.Status.Conditions = append(sm.workflow.Status.Conditions, Condition{
            Type:               ConditionTypeFailed,
            Status:            "True",
            LastTransitionTime: now,
            Reason:            "WorkflowTimeout",
            Message:           fmt.Sprintf("Workflow exceeded timeout of %d seconds", sm.workflow.Spec.TimeoutSeconds),
        })
        return true
    }
    return false
}

func (sm *StatusManager) HandleTimeout() error {
    switch sm.workflow.Spec.TimeoutPolicy {
    case "TIME_OUT":
        return fmt.Errorf("workflow timed out after %d seconds", sm.workflow.Spec.TimeoutSeconds)
    case "ALERT_ONLY":
        // In a real implementation, this would send alerts to monitoring systems
        return nil
    case "RETRY":
        sm.InitializeWorkflow() // Reset the workflow for retry
        return nil
    default:
        return fmt.Errorf("unknown timeout policy: %s", sm.workflow.Spec.TimeoutPolicy)
    }
}