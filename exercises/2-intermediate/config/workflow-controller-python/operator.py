import kopf
import kubernetes.client as k8s
from datetime import datetime
import yaml
import asyncio
from metrics import MetricsCollector
from task_executor import TaskExecutor
from status_manager import StatusManager
from template_manager import TemplateManager

# Initialize global components
metrics = MetricsCollector()
template_mgr = TemplateManager()
task_executor = TaskExecutor(metrics)

@kopf.on.create('netflix.com', 'v1', 'workflows')
async def create_workflow(spec, meta, status, **kwargs):
    """Handler for workflow creation"""
    workflow_name = meta.get('name')
    metrics.record_workflow_start(workflow_name)
    
    status_mgr = StatusManager(metrics)
    status_mgr.initialize_workflow(workflow_name)
    
    try:
        # Process tasks sequentially or in parallel based on dependencies
        tasks = spec.get('tasks', [])
        dependent_tasks = {}
        
        # Group tasks by their dependencies
        for task in tasks:
            deps = task.get('dependsOn', [])
            if not deps:
                dependent_tasks.setdefault(None, []).append(task)
            else:
                for dep in deps:
                    dependent_tasks.setdefault(dep, []).append(task)
        
        # Execute tasks with proper dependency handling
        completed_tasks = set()
        while len(completed_tasks) < len(tasks):
            # Find tasks that can be executed (all dependencies completed)
            executable_tasks = []
            for task in tasks:
                task_deps = set(task.get('dependsOn', []))
                if task['name'] not in completed_tasks and task_deps.issubset(completed_tasks):
                    executable_tasks.append(task)
            
            # Execute current batch of tasks in parallel
            if executable_tasks:
                status_mgr.update_phase("Running")
                await asyncio.gather(*(
                    execute_task(task, workflow_name, status_mgr)
                    for task in executable_tasks
                ))
                completed_tasks.update(task['name'] for task in executable_tasks)
            else:
                # If no tasks can be executed but we haven't completed all tasks,
                # there might be a circular dependency
                if len(completed_tasks) < len(tasks):
                    raise kopf.PermanentError("Detected circular dependency in workflow tasks")
        
        status_mgr.update_phase("Completed")
        metrics.record_workflow_completion(workflow_name, "success")
        return {"status": "Completed"}
        
    except Exception as e:
        status_mgr.update_phase("Failed")
        metrics.record_workflow_completion(workflow_name, "failed")
        raise kopf.PermanentError(f"Workflow failed: {str(e)}")

async def execute_task(task, workflow_name, status_mgr):
    """Execute a single task within the workflow"""
    task_name = task['name']
    status_mgr.start_task(task_name)
    
    try:
        start_time = datetime.now()
        await task_executor.execute_task(task, workflow_name)
        duration = (datetime.now() - start_time).total_seconds()
        
        metrics.record_task_execution(
            task_type=task['taskType'],
            status="success",
            duration=duration,
            workflow_name=workflow_name,
            task_name=task_name
        )
        status_mgr.complete_task(task_name)
        
    except Exception as e:
        if not task.get('optional', False):
            status_mgr.fail_task(task_name, str(e))
            raise
        else:
            status_mgr.skip_task(task_name)

@kopf.on.delete('netflix.com', 'v1', 'workflows')
def delete_workflow(spec, meta, **kwargs):
    """Handler for workflow deletion"""
    workflow_name = meta.get('name')
    metrics.record_workflow_deletion(workflow_name)
    return {"status": "Deleted"}

@kopf.on.resume('netflix.com', 'v1', 'workflows')
async def resume_workflow(spec, meta, status, **kwargs):
    """Handler for workflow resumption after operator restart"""
    workflow_name = meta.get('name')
    current_phase = status.get('phase')
    
    if current_phase not in ["Completed", "Failed"]:
        # Resume workflow from where it left off
        await create_workflow(spec, meta, status, **kwargs)