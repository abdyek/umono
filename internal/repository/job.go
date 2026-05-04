package repository

import (
	"context"
	"errors"
	"time"

	"github.com/umono-cms/umono/internal/models"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) GetByID(ctx context.Context, id string) (models.Job, error) {
	var job models.Job
	err := r.db.WithContext(ctx).Model(&models.Job{}).Where("id = ?", id).First(&job).Error
	return job, err
}

func (r *JobRepository) GetByUniqueKey(ctx context.Context, uniqueKey string) (models.Job, error) {
	var job models.Job
	result := r.db.WithContext(ctx).Model(&models.Job{}).Where("unique_key = ?", uniqueKey).Find(&job)
	if result.Error != nil {
		return models.Job{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.Job{}, gorm.ErrRecordNotFound
	}
	return job, nil
}

func (r *JobRepository) ListAll(ctx context.Context) ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.WithContext(ctx).Model(&models.Job{}).
		Order("CASE status WHEN 'failed' THEN 0 WHEN 'processing' THEN 1 WHEN 'pending' THEN 2 WHEN 'done' THEN 3 ELSE 4 END, updated_at DESC, created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) Create(ctx context.Context, job models.Job) (models.Job, error) {
	err := r.db.WithContext(ctx).Create(&job).Error
	return job, err
}

func (r *JobRepository) AcquireDue(ctx context.Context, now time.Time) (models.Job, bool, error) {
	var job models.Job

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&models.Job{}).
			Where("status = ? AND run_at <= ?", models.JobStatusPending, now).
			Order("run_at ASC, created_at ASC").
			First(&job).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		job.Status = models.JobStatusProcessing
		job.Attempts++

		return tx.Model(&job).Select("*").Updates(job).Error
	})
	if err != nil {
		return models.Job{}, false, err
	}
	if job.ID == "" {
		return models.Job{}, false, nil
	}

	return job, true, nil
}

func (r *JobRepository) NextRunAt(ctx context.Context) (*time.Time, error) {
	var job models.Job
	err := r.db.WithContext(ctx).Model(&models.Job{}).
		Where("status = ?", models.JobStatusPending).
		Order("run_at ASC, created_at ASC").
		First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &job.RunAt, nil
}

func (r *JobRepository) MarkDone(ctx context.Context, id string, finishedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        models.JobStatusDone,
			"finished_at":   finishedAt,
			"last_error":    nil,
			"last_error_at": nil,
		}).Error
}

func (r *JobRepository) MarkRetry(ctx context.Context, id, message string, runAt, failedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        models.JobStatusPending,
			"run_at":        runAt,
			"last_error":    message,
			"last_error_at": failedAt,
		}).Error
}

func (r *JobRepository) MarkFailed(ctx context.Context, id, message string, failedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        models.JobStatusFailed,
			"last_error":    message,
			"last_error_at": failedAt,
			"finished_at":   failedAt,
		}).Error
}

func (r *JobRepository) RetryFailed(ctx context.Context, id string, runAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ? AND status = ?", id, models.JobStatusFailed).
		Updates(map[string]any{
			"status":        models.JobStatusPending,
			"attempts":      0,
			"run_at":        runAt,
			"last_error":    nil,
			"last_error_at": nil,
			"finished_at":   nil,
		})
	return result.RowsAffected > 0, result.Error
}

func (r *JobRepository) ResetProcessing(ctx context.Context, runAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where("status = ?", models.JobStatusProcessing).
		Updates(map[string]any{
			"status": models.JobStatusPending,
			"run_at": runAt,
		}).Error
}

func (r *JobRepository) DeleteDoneBefore(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("status = ? AND finished_at IS NOT NULL AND finished_at < ?", models.JobStatusDone, before).
		Delete(&models.Job{})
	return result.RowsAffected, result.Error
}
