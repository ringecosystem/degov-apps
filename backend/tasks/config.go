package tasks

import (
	"time"

	"github.com/ringecosystem/degov-apps/internal/config"
)

// TaskConfig represents the configuration for a background task
type TaskConfig struct {
	Name     string
	Interval time.Duration
	Enabled  bool
}

// TaskDefinition combines configuration with constructor
type TaskDefinition struct {
	Config      TaskConfig
	Constructor func() Task
}

// GetTaskDefinitions returns all task definitions with their configurations
func GetTaskDefinitions() []TaskDefinition {
	cfg := config.GetConfig()

	return []TaskDefinition{
		{
			Config: TaskConfig{
				Name:     "dao-sync",
				Interval: cfg.GetTaskDAOSyncInterval(),
				Enabled:  cfg.GetTaskDAOSyncEnabled(),
			},
			Constructor: func() Task { return NewDaoSyncTask() },
		},
		{
			Config: TaskConfig{
				Name:     "tracking-vote",
				Interval: cfg.GetTaskVoteTrackingInterval(),
				Enabled:  cfg.GetTaskVoteTrackingEnabled(),
			},
			Constructor: func() Task { return NewTrackingVoteTask() },
		},
		{
			Config: TaskConfig{
				Name:     "tracking-proposal",
				Interval: cfg.GetTaskProposalTrackingInterval(),
				Enabled:  cfg.GetTaskProposalTrackingEnabled(),
			},
			Constructor: func() Task { return NewTrackingProposalTask() },
		},
		// Add more task definitions here
	}
}

// TaskRegistry holds all available task constructors (deprecated)
// Use GetTaskDefinitions() instead for better configuration management
type TaskRegistry struct {
	constructors map[string]func() Task
}

// NewTaskRegistry creates a new task registry (deprecated)
func NewTaskRegistry() *TaskRegistry {
	registry := &TaskRegistry{
		constructors: make(map[string]func() Task),
	}

	// Auto-register tasks from definitions
	for _, def := range GetTaskDefinitions() {
		registry.Register(def.Config.Name, def.Constructor)
	}

	return registry
}

// Register adds a task constructor to the registry
func (tr *TaskRegistry) Register(name string, constructor func() Task) {
	tr.constructors[name] = constructor
}

// Create creates a task instance by name
func (tr *TaskRegistry) Create(name string) Task {
	if constructor, exists := tr.constructors[name]; exists {
		return constructor()
	}
	return nil
}

// GetAvailableTasks returns a list of all registered task names
func (tr *TaskRegistry) GetAvailableTasks() []string {
	names := make([]string, 0, len(tr.constructors))
	for name := range tr.constructors {
		names = append(names, name)
	}
	return names
}
