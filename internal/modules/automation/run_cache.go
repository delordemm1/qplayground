package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RunCache defines the interface for managing automation run state in cache
type RunCache interface {
	// SetRunStatus sets run status without expiry (for active states: pending, running, queued)
	SetRunStatus(ctx context.Context, runID, status string) error
	
	// SetRunStatusWithExpiry sets run status with TTL (for final states: completed, failed, cancelled)
	SetRunStatusWithExpiry(ctx context.Context, runID, status string, ttl time.Duration) error
	
	// GetRunStatus retrieves the current status of a run
	GetRunStatus(ctx context.Context, runID string) (string, error)
	
	// AddRunningRun adds a run to the set of currently running runs
	AddRunningRun(ctx context.Context, runID string) error
	
	// RemoveRunningRun removes a run from the set of currently running runs
	RemoveRunningRun(ctx context.Context, runID string) error
	
	// GetRunningRunCount returns the number of currently running runs
	GetRunningRunCount(ctx context.Context) (int64, error)
	
	// GetAllRunningRuns returns all currently running run IDs
	GetAllRunningRuns(ctx context.Context) ([]string, error)
	
	// GetPendingRuns returns all pending run IDs from Redis
	GetPendingRuns(ctx context.Context) ([]string, error)
	
	// UpsertAllRuns syncs all runs from database to Redis
	UpsertAllRuns(ctx context.Context, runs []*AutomationRun) error
}

// RedisRunCache implements RunCache using Redis
type RedisRunCache struct {
	client *redis.Client
}

// NewRedisRunCache creates a new Redis-based run cache
func NewRedisRunCache(client *redis.Client) RunCache {
	return &RedisRunCache{
		client: client,
	}
}

// SetRunStatus sets run status without expiry
func (r *RedisRunCache) SetRunStatus(ctx context.Context, runID, status string) error {
	key := fmt.Sprintf("run:%s:status", runID)
	return r.client.Set(ctx, key, status, 0).Err()
}

// SetRunStatusWithExpiry sets run status with TTL
func (r *RedisRunCache) SetRunStatusWithExpiry(ctx context.Context, runID, status string, ttl time.Duration) error {
	key := fmt.Sprintf("run:%s:status", runID)
	return r.client.Set(ctx, key, status, ttl).Err()
}

// GetRunStatus retrieves the current status of a run
func (r *RedisRunCache) GetRunStatus(ctx context.Context, runID string) (string, error) {
	key := fmt.Sprintf("run:%s:status", runID)
	result := r.client.Get(ctx, key)
	if result.Err() == redis.Nil {
		return "", fmt.Errorf("run status not found")
	}
	return result.Val(), result.Err()
}

// AddRunningRun adds a run to the set of currently running runs
func (r *RedisRunCache) AddRunningRun(ctx context.Context, runID string) error {
	return r.client.SAdd(ctx, "running_automation_ids", runID).Err()
}

// RemoveRunningRun removes a run from the set of currently running runs
func (r *RedisRunCache) RemoveRunningRun(ctx context.Context, runID string) error {
	return r.client.SRem(ctx, "running_automation_ids", runID).Err()
}

// GetRunningRunCount returns the number of currently running runs
func (r *RedisRunCache) GetRunningRunCount(ctx context.Context) (int64, error) {
	return r.client.SCard(ctx, "running_automation_ids").Result()
}

// GetAllRunningRuns returns all currently running run IDs
func (r *RedisRunCache) GetAllRunningRuns(ctx context.Context) ([]string, error) {
	return r.client.SMembers(ctx, "running_automation_ids").Result()
}

// GetPendingRuns returns all pending run IDs from Redis
func (r *RedisRunCache) GetPendingRuns(ctx context.Context) ([]string, error) {
	// Use SCAN to find all run status keys
	var pendingRuns []string
	iter := r.client.Scan(ctx, 0, "run:*:status", 0).Iterator()
	
	for iter.Next(ctx) {
		key := iter.Val()
		status, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip if error reading status
		}
		
		if status == "pending" || status == "queued" {
			// Extract run ID from key (format: "run:runID:status")
			runID := key[4 : len(key)-7] // Remove "run:" prefix and ":status" suffix
			pendingRuns = append(pendingRuns, runID)
		}
	}
	
	return pendingRuns, iter.Err()
}

// UpsertAllRuns syncs all runs from database to Redis
func (r *RedisRunCache) UpsertAllRuns(ctx context.Context, runs []*AutomationRun) error {
	pipe := r.client.Pipeline()
	
	for _, run := range runs {
		key := fmt.Sprintf("run:%s:status", run.ID)
		
		// Set status based on whether it's a final state or not
		if run.Status == "completed" || run.Status == "failed" || run.Status == "cancelled" {
			pipe.Set(ctx, key, run.Status, 1*time.Minute)
		} else {
			pipe.Set(ctx, key, run.Status, 0) // No expiry for active states
		}
		
		// Add to running set if status is running
		if run.Status == "running" {
			pipe.SAdd(ctx, "running_automation_ids", run.ID)
		}
	}
	
	_, err := pipe.Exec(ctx)
	return err
}