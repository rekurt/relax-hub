package cron

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/nikitaaldaev/bani/config"
	"github.com/nikitaaldaev/bani/internal/logger"
	"github.com/nikitaaldaev/bani/internal/repository/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScheduler(redisClient *redis.Client) *CronScheduler {
	log := logger.New(logger.LevelInfo)
	cfg := &config.Config{}
	cfg.Cron.Enabled = true
	cfg.Cron.Timezone = "UTC"
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()
	return NewCronScheduler(cfg, log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, redisClient, nil, nil, nil, nil, nil)
}

func TestRegister_ValidSpec(t *testing.T) {
	cs := newTestScheduler(nil)

	called := false
	err := cs.Register("* * * * *", "test_job", func(ctx context.Context) error {
		called = true
		return nil
	})
	require.NoError(t, err)

	jobs := cs.Jobs()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test_job", jobs[0].Name)
	assert.Equal(t, "* * * * *", jobs[0].Spec)

	// Job hasn't run yet since we haven't started the scheduler
	assert.False(t, called)
}

func TestRegister_InvalidSpec(t *testing.T) {
	cs := newTestScheduler(nil)

	err := cs.Register("invalid cron spec", "bad_job", func(ctx context.Context) error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad_job")
}

func TestRegister_MultipleJobs(t *testing.T) {
	cs := newTestScheduler(nil)

	for i, name := range []string{"job_a", "job_b", "job_c"} {
		_ = i
		err := cs.Register("* * * * *", name, func(ctx context.Context) error { return nil })
		require.NoError(t, err)
	}

	assert.Len(t, cs.Jobs(), 3)
}

func TestWrapJob_PanicRecovery(t *testing.T) {
	cs := newTestScheduler(nil)

	wrapped := cs.wrapJob("panic_job", func(ctx context.Context) error {
		panic("something went wrong")
	})

	// Should not panic — the wrapper recovers
	assert.NotPanics(t, wrapped)
}

func TestWrapJob_ErrorLogging(t *testing.T) {
	cs := newTestScheduler(nil)

	wrapped := cs.wrapJob("error_job", func(ctx context.Context) error {
		return errors.New("job failed")
	})

	// Should not panic even when the job returns an error
	assert.NotPanics(t, wrapped)
}

func TestWrapJob_ContextProvided(t *testing.T) {
	cs := newTestScheduler(nil)

	var gotCtx context.Context
	wrapped := cs.wrapJob("ctx_job", func(ctx context.Context) error {
		gotCtx = ctx
		return nil
	})

	wrapped()
	require.NotNil(t, gotCtx)

	// Context should have a deadline (from the 5-minute timeout)
	deadline, ok := gotCtx.Deadline()
	assert.True(t, ok, "Context should have a deadline")
	assert.True(t, deadline.After(time.Now()), "Deadline should be in the future")
}

func TestWrapJob_DistributedLock_Acquired(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cs := newTestScheduler(redisClient)

	var callCount int32
	wrapped := cs.wrapJob("lock_job", func(ctx context.Context) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	wrapped()
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))

	// Lock should be released after job completes, so second run should also execute
	wrapped()
	assert.Equal(t, int32(2), atomic.LoadInt32(&callCount))
}

func TestWrapJob_DistributedLock_AlreadyHeld(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cs := newTestScheduler(redisClient)

	// Pre-acquire the lock
	lockKey := "cron_lock:held_job"
	err = redisClient.Set(context.Background(), lockKey, "1", 10*time.Minute).Err()
	require.NoError(t, err)

	var callCount int32
	wrapped := cs.wrapJob("held_job", func(ctx context.Context) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	wrapped()
	// Job should be skipped because the lock is already held
	assert.Equal(t, int32(0), atomic.LoadInt32(&callCount))
}

func TestWrapJob_NoRedis(t *testing.T) {
	cs := newTestScheduler(nil) // no Redis client

	var callCount int32
	wrapped := cs.wrapJob("no_redis_job", func(ctx context.Context) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	})

	wrapped()
	// Should still run without Redis
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount))
}

func TestStart_CronDisabled(t *testing.T) {
	log := logger.New(logger.LevelInfo)
	cfg := &config.Config{}
	cfg.Cron.Enabled = false
	mockSvc := &MockAnalyticsService{}
	mockRepo := mock.NewAnalyticsRepo()
	cs := NewCronScheduler(cfg, log, mockSvc, mockRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	err := cs.Start(context.Background())
	assert.NoError(t, err)

	// No jobs should be registered when disabled
	assert.Empty(t, cs.Jobs())
}

func TestNewCronWithTimezone(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	// Valid timezone
	c := newCronWithTimezone("Europe/Moscow", log)
	assert.NotNil(t, c)

	// Invalid timezone falls back to UTC
	c2 := newCronWithTimezone("Invalid/Zone", log)
	assert.NotNil(t, c2)
}
