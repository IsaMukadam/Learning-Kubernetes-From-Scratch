import kopf
from datetime import datetime

class StatusManager:
    def __init__(self, metrics):
        self.metrics = metrics
        self._workflow_start_time = None

    def initialize_workflow(self, workflow_name):
        """Initialize workflow status"""
        self._workflow_start_time = datetime.now()
        return {
            'phase': 'Initializing',
            'startTime': self._workflow_start_time.isoformat(),
            'tasks': {}
        }

    def update_phase(self, phase):
        """Update workflow phase"""
        return {'phase': phase}

    def start_task(self, task_name):
        """Mark task as started"""
        return {
            'tasks': {
                task_name: {
                    'status': 'Running',
                    'startTime': datetime.now().isoformat()
                }
            }
        }

    def complete_task(self, task_name):
        """Mark task as completed"""
        return {
            'tasks': {
                task_name: {
                    'status': 'Completed',
                    'completionTime': datetime.now().isoformat()
                }
            }
        }

    def fail_task(self, task_name, error_message):
        """Mark task as failed"""
        return {
            'tasks': {
                task_name: {
                    'status': 'Failed',
                    'completionTime': datetime.now().isoformat(),
                    'error': error_message
                }
            }
        }

    def skip_task(self, task_name):
        """Mark task as skipped (for optional tasks)"""
        return {
            'tasks': {
                task_name: {
                    'status': 'Skipped',
                    'completionTime': datetime.now().isoformat()
                }
            }
        }

    def finalize_workflow(self, status='Completed'):
        """Finalize workflow status"""
        end_time = datetime.now()
        duration = (end_time - self._workflow_start_time).total_seconds()
        
        self.metrics.record_workflow_completion(
            status=status,
            duration=duration
        )
        
        return {
            'phase': status,
            'completionTime': end_time.isoformat()
        }