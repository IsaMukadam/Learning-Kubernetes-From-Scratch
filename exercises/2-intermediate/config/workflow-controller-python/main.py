import kopf
import asyncio
from metrics import MetricsCollector
from status_manager import StatusManager
from task_executor import TaskExecutor
from template_manager import TemplateManager

# Initialize components
metrics = MetricsCollector()
status_manager = StatusManager(metrics)
task_executor = TaskExecutor(metrics)
template_manager = TemplateManager()

@kopf.on.create('netflix.com', 'v1', 'workflows')
async def workflow_created(spec, meta, status, **kwargs):
    """Handle workflow creation"""
    workflow_name = meta.get('name')
    metrics.record_workflow_start(workflow_name)
    
    # Initialize workflow status
    status_update = status_manager.initialize_workflow(workflow_name)
    await kopf.patch_status(status_update)
    
    try:
        # Execute tasks in sequence
        for task in spec.get('tasks', []):
            # Update status to show task is starting
            await kopf.patch_status(status_manager.start_task(task['name']))
            
            try:
                # Execute the task
                await task_executor.execute_task(task, workflow_name)
                # Update status to show task completed
                await kopf.patch_status(status_manager.complete_task(task['name']))
                
            except Exception as e:
                # Handle task failure
                await kopf.patch_status(status_manager.fail_task(task['name'], str(e)))
                if not task.get('optional', False):
                    raise
                
        # Mark workflow as completed
        await kopf.patch_status(status_manager.finalize_workflow('Completed'))
        
    except Exception as e:
        # Mark workflow as failed
        await kopf.patch_status(status_manager.finalize_workflow('Failed'))
        raise kopf.PermanentError(f"Workflow failed: {str(e)}")

@kopf.on.delete('netflix.com', 'v1', 'workflows')
def workflow_deleted(meta, **kwargs):
    """Handle workflow deletion"""
    workflow_name = meta.get('name')
    metrics.record_workflow_deletion(workflow_name)

@kopf.on.resume('netflix.com', 'v1', 'workflows')
async def workflow_resumed(spec, meta, status, **kwargs):
    """Handle operator restart - resume workflows"""
    # Implementation would depend on how you want to handle
    # workflows that were in progress when the operator was restarted
    pass

def main():
    """Main entry point"""
    # Start the Kopf operator
    kopf.run()

if __name__ == "__main__":
    main()