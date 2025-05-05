from prometheus_client import Counter, Histogram, Gauge, start_http_server
import time

class MetricsCollector:
    def __init__(self, port=8000):
        # Initialize Prometheus metrics
        self.workflow_starts = Counter(
            'workflow_starts_total',
            'Number of workflow starts'
        )
        self.workflow_completions = Counter(
            'workflow_completions_total',
            'Number of workflow completions',
            ['status']
        )
        self.workflow_duration = Histogram(
            'workflow_duration_seconds',
            'Workflow duration in seconds'
        )
        self.task_completions = Counter(
            'task_completions_total',
            'Number of task completions',
            ['status']
        )
        self.task_duration = Histogram(
            'task_duration_seconds',
            'Task duration in seconds'
        )

        # Start Prometheus HTTP server
        start_http_server(port)

        # Workflow metrics
        self.workflow_duration = Histogram(
            'workflow_duration_seconds',
            'Duration of workflow execution',
            ['workflow_name', 'status'],
            buckets=[60, 300, 600, 1800, 3600, 7200]  # 1m, 5m, 10m, 30m, 1h, 2h
        )
        
        self.active_workflows = Gauge(
            'active_workflows',
            'Number of currently active workflows'
        )
        
        # Task metrics
        self.task_duration = Histogram(
            'task_duration_seconds',
            'Duration of task execution',
            ['task_type', 'workflow_name', 'task_name'],
            buckets=[1, 5, 15, 30, 60, 300]  # 1s, 5s, 15s, 30s, 1m, 5m
        )
        
        self.task_counter = Counter(
            'tasks_total',
            'Total number of tasks by type and status',
            ['task_type', 'status', 'workflow_name']
        )
        
        self.error_counter = Counter(
            'workflow_errors_total',
            'Total number of workflow errors by type',
            ['error_type', 'workflow_name']
        )
        
        self.retry_counter = Counter(
            'task_retries_total',
            'Total number of task retries',
            ['task_type', 'workflow_name', 'task_name']
        )

    def record_workflow_start(self, workflow_name):
        """Record workflow start"""
        self.workflow_starts.inc()
        self.active_workflows.inc()

    def record_workflow_completion(self, status, duration=None):
        """Record workflow completion"""
        self.workflow_completions.labels(status=status).inc()
        self.workflow_duration.observe(duration)
        self.active_workflows.dec()
        if duration is not None:
            self.workflow_duration.labels(workflow_name=workflow_name, status=status).observe(duration)

    def record_workflow_deletion(self, workflow_name):
        """Record workflow deletion"""
        self.active_workflows.dec()

    def record_task_execution(self, task_type, status, duration, workflow_name, task_name):
        self.task_counter.labels(
            task_type=task_type,
            status=status,
            workflow_name=workflow_name
        ).inc()
        
        self.task_duration.labels(
            task_type=task_type,
            workflow_name=workflow_name,
            task_name=task_name
        ).observe(duration)

    def record_error(self, error_type, workflow_name):
        self.error_counter.labels(
            error_type=error_type,
            workflow_name=workflow_name
        ).inc()

    def record_retry(self, task_type, workflow_name, task_name):
        self.retry_counter.labels(
            task_type=task_type,
            workflow_name=workflow_name,
            task_name=task_name
        ).inc()

    def record_task_completion(self, status, duration):
        """Record task completion"""
        self.task_completions.labels(status=status).inc()
        self.task_duration.observe(duration)

    @staticmethod
    def record_duration(func):
        """Decorator to record function duration"""
        async def wrapper(*args, **kwargs):
            start_time = time.time()
            try:
                result = await func(*args, **kwargs)
                duration = time.time() - start_time
                # You might want to add the duration recording here
                return result
            except Exception as e:
                duration = time.time() - start_time
                # You might want to add the duration recording here
                raise
        return wrapper