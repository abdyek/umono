package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestJobServiceEnqueueCreatesPendingJob(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	now := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	job, err := svc.Enqueue(context.Background(), EnqueueJobInput{
		Type:    "media.variant.generate",
		Payload: []byte(`{"media_id":"1"}`),
	})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	if job.ID == "" {
		t.Fatal("expected id")
	}
	if job.Status != models.JobStatusPending {
		t.Fatalf("expected pending status, got %q", job.Status)
	}
	if job.MaxRetry != DefaultJobMaxRetry {
		t.Fatalf("expected default max retry, got %d", job.MaxRetry)
	}
	if !job.RunAt.Equal(now) {
		t.Fatalf("expected run_at %s, got %s", now, job.RunAt)
	}
}

func TestJobServiceProcessReadyJobMarksDone(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	var handled bool

	if err := svc.RegisterHandler("test.done", func(_ context.Context, job models.Job) error {
		handled = true
		if job.Attempts != 1 {
			t.Fatalf("expected first attempt, got %d", job.Attempts)
		}
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	created, err := svc.Enqueue(ctx, EnqueueJobInput{Type: "test.done"})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	processed, err := svc.processNext(ctx)
	if err != nil {
		t.Fatalf("process next: %v", err)
	}
	if !processed {
		t.Fatal("expected job to be processed")
	}
	if !handled {
		t.Fatal("expected handler to run")
	}

	job, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Status != models.JobStatusDone {
		t.Fatalf("expected done status, got %q", job.Status)
	}
	if job.FinishedAt == nil {
		t.Fatal("expected finished_at")
	}
}

func TestJobServiceProcessFailedJobRetriesThenFails(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	handlerErr := errors.New("boom")
	svc.retryDelay = func(int) time.Duration { return 0 }

	if err := svc.RegisterHandler("test.fail", func(context.Context, models.Job) error {
		return handlerErr
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	created, err := svc.Enqueue(ctx, EnqueueJobInput{Type: "test.fail", MaxRetry: 2})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	processed, err := svc.processNext(ctx)
	if err != nil {
		t.Fatalf("process first attempt: %v", err)
	}
	if !processed {
		t.Fatal("expected first attempt to be processed")
	}

	job, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get job after retry: %v", err)
	}
	if job.Status != models.JobStatusPending {
		t.Fatalf("expected pending retry status, got %q", job.Status)
	}
	if job.Attempts != 1 {
		t.Fatalf("expected one attempt, got %d", job.Attempts)
	}
	if job.LastError == nil || *job.LastError != handlerErr.Error() {
		t.Fatalf("expected last error %q, got %v", handlerErr.Error(), job.LastError)
	}

	processed, err = svc.processNext(ctx)
	if err != nil {
		t.Fatalf("process second attempt: %v", err)
	}
	if !processed {
		t.Fatal("expected second attempt to be processed")
	}

	job, err = svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get job after failure: %v", err)
	}
	if job.Status != models.JobStatusFailed {
		t.Fatalf("expected failed status, got %q", job.Status)
	}
	if job.Attempts != 2 {
		t.Fatalf("expected two attempts, got %d", job.Attempts)
	}
	if job.FinishedAt == nil {
		t.Fatal("expected finished_at")
	}
}

func TestJobServiceCleanupDoneDeletesOnlyExpiredDoneJobs(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	oldFinishedAt := now.Add(-DefaultJobDoneRetention - time.Hour)
	recentFinishedAt := now.Add(-DefaultJobDoneRetention + time.Hour)
	jobs := []models.Job{
		{
			ID:         "old-done",
			Type:       "test.cleanup",
			Payload:    []byte("{}"),
			Status:     models.JobStatusDone,
			MaxRetry:   DefaultJobMaxRetry,
			RunAt:      now,
			FinishedAt: &oldFinishedAt,
		},
		{
			ID:         "recent-done",
			Type:       "test.cleanup",
			Payload:    []byte("{}"),
			Status:     models.JobStatusDone,
			MaxRetry:   DefaultJobMaxRetry,
			RunAt:      now,
			FinishedAt: &recentFinishedAt,
		},
		{
			ID:       "old-pending",
			Type:     "test.cleanup",
			Payload:  []byte("{}"),
			Status:   models.JobStatusPending,
			MaxRetry: DefaultJobMaxRetry,
			RunAt:    now,
		},
	}
	if err := db.Create(&jobs).Error; err != nil {
		t.Fatalf("seed jobs: %v", err)
	}

	deleted, err := svc.CleanupDone(ctx)
	if err != nil {
		t.Fatalf("cleanup done: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected one deleted job, got %d", deleted)
	}

	var count int64
	if err := db.Model(&models.Job{}).Where("id = ?", "old-done").Count(&count).Error; err != nil {
		t.Fatalf("count old done: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected old done job to be deleted, count %d", count)
	}
	if err := db.Model(&models.Job{}).Where("id IN ?", []string{"recent-done", "old-pending"}).Count(&count).Error; err != nil {
		t.Fatalf("count retained jobs: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected retained jobs, count %d", count)
	}
}

func newJobTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return db
}

func newJobTestService(db *gorm.DB) *JobService {
	return NewJobService(repository.NewJobRepository(db))
}
