package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/robfig/cron/v3"
)

// JobFunc is the function signature for cron jobs.
type JobFunc func(ctx context.Context) error

// RegisteredJob holds metadata about a registered cron job.
type RegisteredJob struct {
	Name    string
	Spec    string
	EntryID cron.EntryID
}

// Register adds a named job to the scheduler with automatic:
// - Panic recovery
// - Structured logging (start/end/duration/error)
// - Distributed lock via Redis SETNX with TTL (skips if another instance holds the lock)
// - Context with timeout (5 minutes default)
func (cs *CronScheduler) Register(spec, name string, job JobFunc) error {
	wrapped := cs.wrapJob(name, job)
	id, err := cs.c.AddFunc(spec, wrapped)
	if err != nil {
		return fmt.Errorf("failed to register cron job %q: %w", name, err)
	}

	cs.jobs = append(cs.jobs, RegisteredJob{
		Name:    name,
		Spec:    spec,
		EntryID: id,
	})
	cs.logger.Info("Registered cron job", "name", name, "spec", spec)
	return nil
}

// Jobs returns a copy of all registered jobs.
func (cs *CronScheduler) Jobs() []RegisteredJob {
	result := make([]RegisteredJob, len(cs.jobs))
	copy(result, cs.jobs)
	return result
}

func (cs *CronScheduler) wrapJob(name string, job JobFunc) func() {
	return func() {
		// Panic recovery
		defer func() {
			if r := recover(); r != nil {
				cs.logger.Error("Cron job panicked", "job", name, "panic", fmt.Sprintf("%v", r))
			}
		}()

		start := time.Now()
		cs.logger.Info("Cron job starting", "job", name)

		// Distributed lock via Redis SETNX with unique value per invocation
		lockKey := fmt.Sprintf("cron_lock:%s", name)
		lockValue := uuid.New().String()
		lockTTL := 10 * time.Minute

		if cs.redisClient != nil {
			acquired, err := cs.acquireLock(lockKey, lockValue, lockTTL)
			if err != nil {
				cs.logger.Warn("Failed to acquire distributed lock, running anyway",
					"job", name, "error", err)
			} else if !acquired {
				cs.logger.Info("Cron job skipped (another instance holds the lock)",
					"job", name)
				return
			}
			defer cs.releaseLock(lockKey, lockValue)
		}

		// Context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := job(ctx); err != nil {
			cs.logger.Error("Cron job failed",
				"job", name, "error", err, "duration", time.Since(start))
			return
		}

		cs.logger.Info("Cron job completed",
			"job", name, "duration", time.Since(start))
	}
}

func (cs *CronScheduler) acquireLock(key, value string, ttl time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := cs.redisClient.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	return result == "OK", nil
}

// releaseLockScript atomically deletes the key only if it holds the expected value.
// This prevents releasing a lock acquired by another instance after TTL expiry.
var releaseLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`)

func (cs *CronScheduler) releaseLock(key, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := releaseLockScript.Run(ctx, cs.redisClient, []string{key}, value).Err(); err != nil && err != redis.Nil {
		cs.logger.Warn("Failed to release distributed lock", "key", key, "error", err)
	}
}

// newCronWithTimezone creates a robfig/cron instance with the configured timezone.
func newCronWithTimezone(timezone string, l *logger.Logger) *cron.Cron {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		l.Warn("Invalid cron timezone, falling back to UTC", "timezone", timezone, "error", err)
		loc = time.UTC
	}
	return cron.New(cron.WithLocation(loc))
}
