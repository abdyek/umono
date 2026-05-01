package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/repository"
	"gorm.io/gorm"
)

const (
	DefaultJobMaxRetry      = 3
	DefaultJobPollInterval  = time.Minute
	DefaultJobDoneRetention = 14 * 24 * time.Hour
	DefaultJobCleanupEvery  = 24 * time.Hour
)

var (
	ErrJobNotFound         = errors.New("job not found")
	ErrJobTypeRequired     = errors.New("job type required")
	ErrJobPayloadInvalid   = errors.New("job payload invalid")
	ErrJobHandlerRequired  = errors.New("job handler required")
	ErrJobHandlerNotFound  = errors.New("job handler not found")
	ErrJobAlreadyStarted   = errors.New("job service already started")
	ErrJobMaxRetryInvalid  = errors.New("job max retry invalid")
	ErrJobRetentionInvalid = errors.New("job retention invalid")
	ErrJobRetryNotAllowed  = errors.New("job retry not allowed")
)

type JobEnqueuer interface {
	Enqueue(context.Context, EnqueueJobInput) (models.Job, error)
}

type JobHandler func(context.Context, models.Job, JobEnqueuer) error

var _ JobEnqueuer = jobEnqueuer{}

type JobService struct {
	repo *repository.JobRepository

	mu       sync.RWMutex
	handlers map[string]JobHandler

	started bool
	wake    chan struct{}

	pollInterval    time.Duration
	doneRetention   time.Duration
	cleanupInterval time.Duration
	retryDelay      func(int) time.Duration
	now             func() time.Time
}

type EnqueueJobInput struct {
	Type      string
	UniqueKey string
	Payload   []byte
	RunAt     *time.Time
	MaxRetry  int
}

type jobEnqueuer struct {
	service *JobService
}

func (e jobEnqueuer) Enqueue(ctx context.Context, input EnqueueJobInput) (models.Job, error) {
	return e.service.Enqueue(ctx, input)
}

func NewJobService(repo *repository.JobRepository) *JobService {
	return &JobService{
		repo:            repo,
		handlers:        map[string]JobHandler{},
		wake:            make(chan struct{}, 1),
		pollInterval:    DefaultJobPollInterval,
		doneRetention:   DefaultJobDoneRetention,
		cleanupInterval: DefaultJobCleanupEvery,
		retryDelay:      defaultJobRetryDelay,
		now:             time.Now,
	}
}

func (s *JobService) RegisterHandler(jobType string, handler JobHandler) error {
	jobType = strings.TrimSpace(jobType)
	if jobType == "" {
		return ErrJobTypeRequired
	}
	if handler == nil {
		return ErrJobHandlerRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[jobType] = handler
	s.notify()
	return nil
}

func (s *JobService) Enqueue(ctx context.Context, input EnqueueJobInput) (models.Job, error) {
	jobType := strings.TrimSpace(input.Type)
	if jobType == "" {
		return models.Job{}, ErrJobTypeRequired
	}

	var uniqueKey *string
	trimmedUniqueKey := strings.TrimSpace(input.UniqueKey)
	if trimmedUniqueKey != "" {
		uniqueKey = &trimmedUniqueKey
	}

	payload := input.Payload
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	if !json.Valid(payload) {
		return models.Job{}, ErrJobPayloadInvalid
	}

	maxRetry := input.MaxRetry
	if maxRetry == 0 {
		maxRetry = DefaultJobMaxRetry
	}
	if maxRetry < 0 {
		return models.Job{}, ErrJobMaxRetryInvalid
	}

	runAt := s.now()
	if input.RunAt != nil {
		runAt = *input.RunAt
	}

	id, err := uuid.NewV7()
	if err != nil {
		return models.Job{}, err
	}

	job := models.Job{
		ID:        id.String(),
		Type:      jobType,
		UniqueKey: uniqueKey,
		Payload:   payload,
		Status:    models.JobStatusPending,
		MaxRetry:  maxRetry,
		RunAt:     runAt,
	}

	job, err = s.repo.Create(ctx, job)
	if err != nil {
		return models.Job{}, err
	}

	s.notify()
	return job, nil
}

func (s *JobService) GetByID(ctx context.Context, id string) (models.Job, error) {
	job, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Job{}, ErrJobNotFound
	}
	if err != nil {
		return models.Job{}, err
	}

	return job, nil
}

func (s *JobService) ListAll(ctx context.Context) ([]models.Job, error) {
	return s.repo.ListAll(ctx)
}

func (s *JobService) RetryFailed(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrJobNotFound
	}

	job, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if job.Status != models.JobStatusFailed {
		return ErrJobRetryNotAllowed
	}

	updated, err := s.repo.RetryFailed(ctx, id, s.now())
	if err != nil {
		return err
	}
	if !updated {
		return ErrJobRetryNotAllowed
	}

	s.notify()
	return nil
}

func (s *JobService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return ErrJobAlreadyStarted
	}
	s.started = true
	s.mu.Unlock()

	go s.run(ctx)
	go s.cleanupLoop(ctx)

	return nil
}

func (s *JobService) CleanupDone(ctx context.Context) (int64, error) {
	if s.doneRetention <= 0 {
		return 0, ErrJobRetentionInvalid
	}

	return s.repo.DeleteDoneBefore(ctx, s.now().Add(-s.doneRetention))
}

func (s *JobService) run(ctx context.Context) {
	_ = s.repo.ResetProcessing(ctx, s.now())

	for {
		for {
			processed, err := s.processNext(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil || !processed {
				break
			}
		}

		wait := s.nextWait(ctx)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-s.wake:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}
	}
}

func (s *JobService) cleanupLoop(ctx context.Context) {
	_, _ = s.CleanupDone(ctx)

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = s.CleanupDone(ctx)
		}
	}
}

func (s *JobService) processNext(ctx context.Context) (bool, error) {
	now := s.now()
	job, ok, err := s.repo.AcquireDue(ctx, now)
	if err != nil || !ok {
		return ok, err
	}

	handler := s.handler(job.Type)
	if handler == nil {
		err = ErrJobHandlerNotFound
	} else {
		err = handler(ctx, job, jobEnqueuer{service: s})
	}

	finishedAt := s.now()
	if err == nil {
		return true, s.repo.MarkDone(ctx, job.ID, finishedAt)
	}

	message := err.Error()
	if job.Attempts >= job.MaxRetry {
		return true, s.repo.MarkFailed(ctx, job.ID, message, finishedAt)
	}

	return true, s.repo.MarkRetry(ctx, job.ID, message, finishedAt.Add(s.retryDelay(job.Attempts)), finishedAt)
}

func (s *JobService) handler(jobType string) JobHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.handlers[jobType]
}

func (s *JobService) nextWait(ctx context.Context) time.Duration {
	next, err := s.repo.NextRunAt(ctx)
	if err != nil || next == nil {
		return s.pollInterval
	}

	wait := next.Sub(s.now())
	if wait < 0 {
		return 0
	}

	return wait
}

func (s *JobService) notify() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func defaultJobRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}

	return time.Duration(attempt*attempt) * time.Minute
}
