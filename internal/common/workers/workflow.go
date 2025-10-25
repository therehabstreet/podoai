package workers

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// WorkflowRun represents a single workflow run
type WorkflowRun interface {
	GetID() string
	GetType() string
	GetSteps() []WorkflowStep
	GetInput() map[string]any
}

// WorkflowStep represents a single step in a workflow run
type WorkflowStep interface {
	Execute(ctx context.Context, input map[string]any) error
	GetName() string
}

// WorkflowEngine handles async processing of workflow runs with a worker pool
type WorkflowEngine struct {
	name       string
	workerPool chan struct{}
	queue      chan WorkflowRun
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// WorkflowConfig configuration for workflow engine
type WorkflowConfig struct {
	Name       string
	MaxWorkers int
	QueueSize  int
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(config *WorkflowConfig) *WorkflowEngine {
	ctx, cancel := context.WithCancel(context.Background())
	engine := &WorkflowEngine{
		name:       config.Name,
		workerPool: make(chan struct{}, config.MaxWorkers),
		queue:      make(chan WorkflowRun, config.QueueSize),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start worker pool
	go engine.startWorkers()

	log.Printf("Workflow engine '%s' initialized (workers: %d, queue: %d)",
		config.Name, config.MaxWorkers, config.QueueSize)

	return engine
}

// startWorkers starts the worker pool that processes workflow runs
func (we *WorkflowEngine) startWorkers() {
	for {
		select {
		case <-we.ctx.Done():
			log.Printf("Workflow '%s': shutting down workers", we.name)
			return
		case run := <-we.queue:
			// Acquire a worker slot
			we.workerPool <- struct{}{}

			we.wg.Add(1)
			go func(run WorkflowRun) {
				defer we.wg.Done()
				defer func() { <-we.workerPool }() // Release worker slot
				we.executeRun(run)
			}(run)
		}
	}
}

// SubmitRun queues a task for async processing
func (we *WorkflowEngine) SubmitRun(ctx context.Context, run WorkflowRun) error {
	select {
	case we.queue <- run:
		log.Printf("Workflow '%s': Run %s (%s) queued for processing", we.name, run.GetID(), run.GetType())
		return nil
	case <-ctx.Done():
		return fmt.Errorf("workflow '%s': context cancelled while queuing run %s: %w", we.name, run.GetID(), ctx.Err())
	case <-we.ctx.Done():
		return fmt.Errorf("workflow '%s' is shutting down", we.name)
	}
}

// executeRun executes a single run through all registered steps
func (we *WorkflowEngine) executeRun(run WorkflowRun) {
	log.Printf("Workflow '%s': Processing run %s (%s)", we.name, run.GetID(), run.GetType())

	// Get steps for execution
	steps := run.GetSteps()
	if len(steps) == 0 {
		log.Printf("Workflow '%s': No steps to execute for run %s", we.name, run.GetID())
		return
	}

	// Create a context for this run execution
	ctx := context.Background()

	// Execute all steps sequentially
	for i, step := range steps {
		log.Printf("Workflow '%s': Run %s - Executing step %d/%d: %s",
			we.name, run.GetID(), i+1, len(steps), step.GetName())

		if err := step.Execute(ctx, run.GetInput()); err != nil {
			log.Printf("Workflow '%s': Error processing run %s at step '%s': %v",
				we.name, run.GetID(), step.GetName(), err)
			return
		}
	}

	log.Printf("Workflow '%s': Successfully executed run %s", we.name, run.GetID())
}

// Shutdown gracefully shuts down the workflow engine
func (we *WorkflowEngine) Shutdown() {
	log.Printf("Workflow '%s': Initiating shutdown...", we.name)
	we.cancel()

	// Close task queue to prevent new tasks
	close(we.queue)

	// Wait for all workers to complete
	we.wg.Wait()

	log.Printf("Workflow '%s': Shutdown complete", we.name)
}
