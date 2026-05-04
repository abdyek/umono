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
	if job.UniqueKey != nil {
		t.Fatalf("expected nil unique key, got %q", *job.UniqueKey)
	}
	if !job.RunAt.Equal(now) {
		t.Fatalf("expected run_at %s, got %s", now, job.RunAt)
	}
}

func TestJobServiceEnqueueStoresUniqueKey(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)

	job, err := svc.Enqueue(context.Background(), EnqueueJobInput{
		Type:      "media.variant.generate",
		UniqueKey: " media:1:variant:large ",
	})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	if job.UniqueKey == nil {
		t.Fatal("expected unique key")
	}
	if *job.UniqueKey != "media:1:variant:large" {
		t.Fatalf("expected trimmed unique key, got %q", *job.UniqueKey)
	}

	got, err := svc.GetByID(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if got.UniqueKey == nil {
		t.Fatal("expected persisted unique key")
	}
	if *got.UniqueKey != "media:1:variant:large" {
		t.Fatalf("expected persisted unique key, got %q", *got.UniqueKey)
	}
}

func TestJobServiceEnqueueReturnsExistingJobForUniqueKey(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()

	first, err := svc.Enqueue(ctx, EnqueueJobInput{
		Type:      "media.variant.plan",
		UniqueKey: "media.variant.plan:media-id",
		Payload:   []byte(`{"media_id":"media-id"}`),
	})
	if err != nil {
		t.Fatalf("enqueue first job: %v", err)
	}

	second, err := svc.Enqueue(ctx, EnqueueJobInput{
		Type:      "media.variant.plan",
		UniqueKey: "media.variant.plan:media-id",
		Payload:   []byte(`{"media_id":"media-id"}`),
	})
	if err != nil {
		t.Fatalf("enqueue second job: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("expected existing job, got %q want %q", second.ID, first.ID)
	}

	jobs, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one job, got %d", len(jobs))
	}
}

func TestJobServiceProcessReadyJobMarksDone(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	var handled bool

	if err := svc.RegisterHandler("test.done", func(_ context.Context, job models.Job, _ JobEnqueuer) error {
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

func TestJobServiceHandlerCanEnqueueJob(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()

	if err := svc.RegisterHandler("test.parent", func(ctx context.Context, _ models.Job, enqueuer JobEnqueuer) error {
		_, err := enqueuer.Enqueue(ctx, EnqueueJobInput{
			Type:    "test.child",
			Payload: []byte(`{"source":"parent"}`),
		})
		return err
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	created, err := svc.Enqueue(ctx, EnqueueJobInput{Type: "test.parent"})
	if err != nil {
		t.Fatalf("enqueue parent job: %v", err)
	}

	processed, err := svc.processNext(ctx)
	if err != nil {
		t.Fatalf("process parent job: %v", err)
	}
	if !processed {
		t.Fatal("expected parent job to be processed")
	}

	parent, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get parent job: %v", err)
	}
	if parent.Status != models.JobStatusDone {
		t.Fatalf("expected parent job done, got %q", parent.Status)
	}

	jobs, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	var child models.Job
	for _, job := range jobs {
		if job.Type == "test.child" {
			child = job
			break
		}
	}
	if child.ID == "" {
		t.Fatal("expected child job to be enqueued")
	}
	if child.Status != models.JobStatusPending {
		t.Fatalf("expected child job pending, got %q", child.Status)
	}
	if string(child.Payload) != `{"source":"parent"}` {
		t.Fatalf("expected child payload from parent, got %s", child.Payload)
	}
}

func TestJobServiceProcessFailedJobRetriesThenFails(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	handlerErr := errors.New("boom")
	svc.retryDelay = func(int) time.Duration { return 0 }

	if err := svc.RegisterHandler("test.fail", func(context.Context, models.Job, JobEnqueuer) error {
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

func TestJobServiceListAllUsesAdminOrder(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)

	jobs := []models.Job{
		{ID: "done", Type: "test", Payload: []byte("{}"), Status: models.JobStatusDone, MaxRetry: 3, RunAt: now},
		{ID: "pending", Type: "test", Payload: []byte("{}"), Status: models.JobStatusPending, MaxRetry: 3, RunAt: now},
		{ID: "failed", Type: "test", Payload: []byte("{}"), Status: models.JobStatusFailed, MaxRetry: 3, RunAt: now},
		{ID: "processing", Type: "test", Payload: []byte("{}"), Status: models.JobStatusProcessing, MaxRetry: 3, RunAt: now},
	}
	if err := db.Create(&jobs).Error; err != nil {
		t.Fatalf("seed jobs: %v", err)
	}

	got, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}

	wantIDs := []string{"failed", "processing", "pending", "done"}
	if len(got) != len(wantIDs) {
		t.Fatalf("expected %d jobs, got %d", len(wantIDs), len(got))
	}
	for i, wantID := range wantIDs {
		if got[i].ID != wantID {
			t.Fatalf("expected job %d to be %q, got %q", i, wantID, got[i].ID)
		}
	}
}

func TestJobServiceRetryFailedResetsExistingJob(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
	failedAt := now.Add(-time.Hour)
	message := "boom"
	svc.now = func() time.Time { return now }

	job := models.Job{
		ID:          "failed-job",
		Type:        "test.retry",
		Payload:     []byte("{}"),
		Status:      models.JobStatusFailed,
		Attempts:    3,
		MaxRetry:    3,
		RunAt:       failedAt,
		LastError:   &message,
		LastErrorAt: &failedAt,
		FinishedAt:  &failedAt,
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("seed job: %v", err)
	}

	if err := svc.RetryFailed(ctx, job.ID); err != nil {
		t.Fatalf("retry failed job: %v", err)
	}

	got, err := svc.GetByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if got.Status != models.JobStatusPending {
		t.Fatalf("expected pending status, got %q", got.Status)
	}
	if got.Attempts != 0 {
		t.Fatalf("expected attempts to reset, got %d", got.Attempts)
	}
	if !got.RunAt.Equal(now) {
		t.Fatalf("expected run_at %s, got %s", now, got.RunAt)
	}
	if got.LastError != nil {
		t.Fatalf("expected last_error to be cleared, got %v", got.LastError)
	}
	if got.LastErrorAt != nil {
		t.Fatalf("expected last_error_at to be cleared, got %v", got.LastErrorAt)
	}
	if got.FinishedAt != nil {
		t.Fatalf("expected finished_at to be cleared, got %v", got.FinishedAt)
	}
}

func TestJobServiceRetryFailedRejectsNonFailedJob(t *testing.T) {
	db := newJobTestDB(t)
	svc := newJobTestService(db)
	ctx := context.Background()
	now := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)

	job := models.Job{
		ID:       "pending-job",
		Type:     "test.retry",
		Payload:  []byte("{}"),
		Status:   models.JobStatusPending,
		MaxRetry: 3,
		RunAt:    now,
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("seed job: %v", err)
	}

	err := svc.RetryFailed(ctx, job.ID)
	if !errors.Is(err, ErrJobRetryNotAllowed) {
		t.Fatalf("expected retry not allowed, got %v", err)
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
