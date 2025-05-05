from kubernetes import client, config
import asyncio
import logging
from datetime import datetime

class TaskExecutor:
    def __init__(self, metrics_collector):
        self.metrics_collector = metrics_collector
        self.v1 = client.CoreV1Api()
        self.batch_v1 = client.BatchV1Api()

    async def execute_task(self, task, workflow_name, namespace):
        """Execute a single task in the workflow"""
        logging.info(f"Executing task: {task['name']} for workflow: {workflow_name}")
        
        start_time = datetime.now()
        
        try:
            if task['type'] == 'job':
                await self._create_job(task, workflow_name, namespace)
            elif task['type'] == 'service':
                await self._create_service(task, namespace)
            else:
                raise ValueError(f"Unsupported task type: {task['type']}")

            # Record task completion
            duration = (datetime.now() - start_time).total_seconds()
            self.metrics_collector.record_task_completion('success', duration)
            
            return True
            
        except Exception as e:
            logging.error(f"Error executing task {task['name']}: {str(e)}")
            duration = (datetime.now() - start_time).total_seconds()
            self.metrics_collector.record_task_completion('failure', duration)
            return False

    async def _create_job(self, task, workflow_name, namespace):
        """Create a Kubernetes Job"""
        job = client.V1Job(
            metadata=client.V1ObjectMeta(
                name=f"{workflow_name}-{task['name']}",
                namespace=namespace
            ),
            spec=client.V1JobSpec(
                template=client.V1PodTemplateSpec(
                    spec=client.V1PodSpec(
                        containers=[
                            client.V1Container(
                                name=task['name'],
                                image=task['image'],
                                command=task.get('command', []),
                                args=task.get('args', [])
                            )
                        ],
                        restart_policy='Never'
                    )
                )
            )
        )
        
        await self.batch_v1.create_namespaced_job(namespace, job)

    async def _create_service(self, task, namespace):
        """Create a Kubernetes Service"""
        service = client.V1Service(
            metadata=client.V1ObjectMeta(
                name=task['name'],
                namespace=namespace
            ),
            spec=client.V1ServiceSpec(
                selector=task['selector'],
                ports=[client.V1ServicePort(
                    port=port['port'],
                    target_port=port.get('targetPort', port['port']),
                    protocol=port.get('protocol', 'TCP')
                ) for port in task['ports']]
            )
        )
        
        await self.v1.create_namespaced_service(namespace, service)

    async def cleanup_task(self, task_name, workflow_name, namespace):
        """Clean up resources created by a task"""
        try:
            # Delete job if it exists
            await self.batch_v1.delete_namespaced_job(
                name=f"{workflow_name}-{task_name}",
                namespace=namespace,
                body=client.V1DeleteOptions()
            )
        except client.rest.ApiException as e:
            if e.status != 404:  # Ignore if job doesn't exist
                logging.error(f"Error cleaning up job {task_name}: {str(e)}")

        try:
            # Delete service if it exists
            await self.v1.delete_namespaced_service(
                name=task_name,
                namespace=namespace,
                body=client.V1DeleteOptions()
            )
        except client.rest.ApiException as e:
            if e.status != 404:  # Ignore if service doesn't exist
                logging.error(f"Error cleaning up service {task_name}: {str(e)}")