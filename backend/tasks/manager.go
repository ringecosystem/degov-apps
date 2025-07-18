package tasks

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type TaskManager struct {
	scheduler        gocron.Scheduler
	tasks            []Task
	metricsCollector *MetricsCollector
}

// Task interface for all background tasks
type Task interface {
	Name() string
	Execute() error
}

// NewTaskManager creates a new task manager with gocron scheduler
func NewTaskManager() (*TaskManager, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &TaskManager{
		scheduler:        scheduler,
		tasks:            make([]Task, 0),
		metricsCollector: NewMetricsCollector(),
	}, nil
}

// RegisterTask registers a new task with the scheduler
func (tm *TaskManager) RegisterTask(task Task, interval time.Duration) error {
	tm.tasks = append(tm.tasks, task)

	_, err := tm.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(
			func() {
				slog.Info("Executing scheduled task", "task", task.Name())

				startTime := time.Now()
				err := task.Execute()
				duration := time.Since(startTime)

				// Track metrics
				tm.metricsCollector.TrackExecution(task.Name(), duration, err)

				if err != nil {
					slog.Error("Task execution failed", "task", task.Name(), "error", err)
				} else {
					slog.Debug("Task execution completed", "task", task.Name(), "duration", duration.String())
				}
			},
		),
		gocron.WithName(task.Name()),
	)

	if err != nil {
		return err
	}

	slog.Info("Task registered successfully", "task", task.Name(), "interval", interval.String())
	return nil
}

// Start starts the task scheduler
func (tm *TaskManager) Start(ctx context.Context) {
	slog.Info("Starting task manager", "registered_tasks", len(tm.tasks))

	// Execute all tasks immediately on startup
	for _, task := range tm.tasks {
		go func(t Task) {
			slog.Info("Running initial execution", "task", t.Name())

			startTime := time.Now()
			err := t.Execute()
			duration := time.Since(startTime)

			// Track metrics for initial execution
			tm.metricsCollector.TrackExecution(t.Name(), duration, err)

			if err != nil {
				slog.Error("Initial task execution failed", "task", t.Name(), "error", err)
			}
		}(task)
	}

	// Start the scheduler
	tm.scheduler.Start()

	// Wait for context cancellation
	<-ctx.Done()
	slog.Info("Stopping task manager")

	// Log final metrics summary
	tm.metricsCollector.LogSummary()

	// Shutdown the scheduler gracefully
	if err := tm.scheduler.Shutdown(); err != nil {
		slog.Error("Error shutting down scheduler", "error", err)
	}
}

// GetTaskCount returns the number of registered tasks
func (tm *TaskManager) GetTaskCount() int {
	return len(tm.tasks)
}

// ListTasks returns a list of task names
func (tm *TaskManager) ListTasks() []string {
	names := make([]string, len(tm.tasks))
	for i, task := range tm.tasks {
		names[i] = task.Name()
	}
	return names
}

// GetMetrics returns the metrics collector
func (tm *TaskManager) GetMetrics() *MetricsCollector {
	return tm.metricsCollector
}
