package tasks

import (
	"log/slog"
	"time"
)

// TaskMetrics holds metrics for task execution
type TaskMetrics struct {
	Name              string
	LastExecution     time.Time
	LastExecutionTime time.Duration
	ExecutionCount    int64
	ErrorCount        int64
	LastError         error
}

// MetricsCollector collects and tracks task execution metrics
type MetricsCollector struct {
	metrics map[string]*TaskMetrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*TaskMetrics),
	}
}

// TrackExecution records a task execution
func (mc *MetricsCollector) TrackExecution(taskName string, duration time.Duration, err error) {
	if mc.metrics[taskName] == nil {
		mc.metrics[taskName] = &TaskMetrics{
			Name: taskName,
		}
	}

	metric := mc.metrics[taskName]
	metric.LastExecution = time.Now()
	metric.LastExecutionTime = duration
	metric.ExecutionCount++

	if err != nil {
		metric.ErrorCount++
		metric.LastError = err
		slog.Error("Task execution failed",
			"task", taskName,
			"error", err,
			"execution_count", metric.ExecutionCount,
			"error_count", metric.ErrorCount)
	} else {
		metric.LastError = nil
		slog.Debug("Task execution successful",
			"task", taskName,
			"duration", duration.String(),
			"execution_count", metric.ExecutionCount)
	}
}

// GetMetrics returns metrics for a specific task
func (mc *MetricsCollector) GetMetrics(taskName string) *TaskMetrics {
	return mc.metrics[taskName]
}

// GetAllMetrics returns all task metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]*TaskMetrics {
	return mc.metrics
}

// LogSummary logs a summary of all task metrics
func (mc *MetricsCollector) LogSummary() {
	if len(mc.metrics) == 0 {
		slog.Info("No task metrics available")
		return
	}

	slog.Info("Task execution summary", "total_tasks", len(mc.metrics))

	for _, metric := range mc.metrics {
		slog.Info("Task metrics",
			"task", metric.Name,
			"executions", metric.ExecutionCount,
			"errors", metric.ErrorCount,
			"last_execution", metric.LastExecution.Format(time.RFC3339),
			"last_duration", metric.LastExecutionTime.String(),
			"success_rate", float64(metric.ExecutionCount-metric.ErrorCount)/float64(metric.ExecutionCount)*100)
	}
}
