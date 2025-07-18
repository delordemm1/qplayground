package automation

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/delordemm1/qplayground/internal/platform"
)

// Scheduler handles background job scheduling for automation runs
type Scheduler struct {
	automationRepo    AutomationRepository
	automationService AutomationService
	runCache          RunCache
	runner            *Runner
	sseManager        *SSEManager
	maxConcurrentRuns int
	ticker            *time.Ticker
	stopCh            chan struct{}
	mu                sync.Mutex
	runContexts       map[string]context.CancelFunc
}

// NewScheduler creates a new automation scheduler
func NewScheduler(
	automationRepo AutomationRepository,
	automationService AutomationService,
	runCache RunCache,
	runner *Runner,
	sseManager *SSEManager,
) *Scheduler {
	return &Scheduler{
		automationRepo:    automationRepo,
		automationService: automationService,
		runCache:          runCache,
		runner:            runner,
		sseManager:        sseManager,
		maxConcurrentRuns: platform.ENV_MAX_CONCURRENT_RUNS,
		stopCh:            make(chan struct{}),
		runContexts:       make(map[string]context.CancelFunc),
	}
}

// Start begins the scheduler's background processing
func (s *Scheduler) Start(ctx context.Context) {
	s.ticker = time.NewTicker(10 * time.Second)

	slog.Info("Automation scheduler started", "interval", "10s", "max_concurrent_runs", s.maxConcurrentRuns)

	go func() {
		defer s.ticker.Stop()

		for {
			select {
			case <-s.ticker.C:
				s.processPendingRuns(ctx)
			case <-s.stopCh:
				slog.Info("Automation scheduler stopped")
				return
			case <-ctx.Done():
				slog.Info("Automation scheduler context cancelled")
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)

	// Cancel all running contexts
	s.mu.Lock()
	for runID, cancel := range s.runContexts {
		slog.Info("Cancelling run due to scheduler shutdown", "run_id", runID)
		cancel()
	}
	s.runContexts = make(map[string]context.CancelFunc)
	s.mu.Unlock()
}

// processPendingRuns checks for pending runs and starts them if capacity allows
func (s *Scheduler) processPendingRuns(ctx context.Context) {
	// Check current running count
	runningCount, err := s.runCache.GetRunningRunCount(ctx)
	if err != nil {
		slog.Error("Failed to get running run count", "error", err)
		return
	}

	// If at capacity, skip processing
	if runningCount >= int64(s.maxConcurrentRuns) {
		slog.Debug("At max concurrent runs capacity", "running", runningCount, "max", s.maxConcurrentRuns)
		return
	}

	// Get pending runs from Redis
	pendingRuns, err := s.runCache.GetPendingRuns(ctx)
	if err != nil {
		slog.Error("Failed to get pending runs", "error", err)
		return
	}

	if len(pendingRuns) == 0 {
		return // No pending runs
	}

	slog.Debug("Processing pending runs", "pending_count", len(pendingRuns), "running_count", runningCount)

	// Process pending runs up to capacity
	availableSlots := int(int64(s.maxConcurrentRuns) - runningCount)
	for i, runID := range pendingRuns {
		if i >= availableSlots {
			break // No more capacity
		}

		// Get run details from database
		run, err := s.automationRepo.GetRunByID(ctx, runID)
		if err != nil {
			slog.Error("Failed to get run details", "run_id", runID, "error", err)
			continue
		}
		// Double-check status in case it changed
		if run.Status != "pending" && run.Status != "queued" {
			continue
		}

		automation, err := s.automationRepo.GetAutomationByID(ctx, run.AutomationID)
		if err != nil {
			slog.Error("Failed to get automation details", "automation_id", run.AutomationID, "error", err)
			continue
		}
		// Start the run
		s.startRun(ctx, automation.ProjectID, run)
	}
}

// startRun starts a single automation run
func (s *Scheduler) startRun(ctx context.Context, projectID string, run *AutomationRun) {
	slog.Info("Starting automation run", "run_id", run.ID, "automation_id", run.AutomationID)

	// Update status to running in DB and Redis
	run.Status = "running"
	startTime := time.Now()
	run.StartTime = &startTime

	err := s.automationRepo.UpdateRun(ctx, run)
	if err != nil {
		slog.Error("Failed to update run status to running", "run_id", run.ID, "error", err)
		return
	}

	err = s.runCache.SetRunStatus(ctx, run.ID, "running")
	if err != nil {
		slog.Error("Failed to update run status in cache", "run_id", run.ID, "error", err)
	}

	err = s.runCache.AddRunningRun(ctx, run.ID)
	if err != nil {
		slog.Error("Failed to add run to running set", "run_id", run.ID, "error", err)
	}

	// Send status update via SSE
	if s.sseManager != nil {
		s.sseManager.SendRunStatusUpdate(projectID, run.AutomationID, run.ID, "running")
	}

	// Create cancellable context for this run
	runCtx, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	s.runContexts[run.ID] = cancel
	s.mu.Unlock()

	// Start the automation in a goroutine
	go func() {
		defer func() {
			// Cleanup on completion
			s.mu.Lock()
			delete(s.runContexts, run.ID)
			s.mu.Unlock()

			cancel() // Release context resources

			// Remove from running set
			if err := s.runCache.RemoveRunningRun(context.Background(), run.ID); err != nil {
				slog.Error("Failed to remove run from running set", "run_id", run.ID, "error", err)
			}
		}()

		// Execute the automation
		err := s.runner.RunAutomation(runCtx, projectID, run)

		// Update final status
		endTime := time.Now()
		run.EndTime = &endTime

		if err != nil {
			run.Status = "failed"
			run.ErrorMessage = err.Error()
			slog.Error("Automation run failed", "run_id", run.ID, "error", err)
		} else {
			run.Status = "completed"
			slog.Info("Automation run completed", "run_id", run.ID)
		}

		// Update in database
		if updateErr := s.automationRepo.UpdateRun(context.Background(), run); updateErr != nil {
			slog.Error("Failed to update final run status", "run_id", run.ID, "error", updateErr)
		}

		// Update in Redis with expiry (final state)
		if cacheErr := s.runCache.SetRunStatusWithExpiry(context.Background(), run.ID, run.Status, 1*time.Minute); cacheErr != nil {
			slog.Error("Failed to update final run status in cache", "run_id", run.ID, "error", cacheErr)
		}

		// Send final status update via SSE
		if s.sseManager != nil {
			s.sseManager.SendRunStatusUpdate(projectID, run.AutomationID, run.ID, run.Status)
		}
	}()
}

// CancelRun cancels a specific automation run
func (s *Scheduler) CancelRun(ctx context.Context, projectID, runID string) error {
	s.mu.Lock()
	cancel, exists := s.runContexts[runID]
	if exists {
		delete(s.runContexts, runID)
	}
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("run not found or not running")
	}

	// Cancel the context
	cancel()

	// Update status in database
	run, err := s.automationRepo.GetRunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	run.Status = "cancelled"
	endTime := time.Now()
	run.EndTime = &endTime

	err = s.automationRepo.UpdateRun(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	// Update status in Redis with expiry
	err = s.runCache.SetRunStatusWithExpiry(ctx, runID, "cancelled", 1*time.Minute)
	if err != nil {
		slog.Error("Failed to update cancelled status in cache", "run_id", runID, "error", err)
	}

	// Send cancellation update via SSE
	if s.sseManager != nil {
		s.sseManager.SendRunStatusUpdate(projectID, run.AutomationID, runID, "cancelled")
	}

	slog.Info("Automation run cancelled", "run_id", runID)
	return nil
}
