package models

import "time"

const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusDone       = "done"
	JobStatusFailed     = "failed"
)

type Job struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id" db:"id"`
	Type        string     `gorm:"not null;index" json:"type" db:"type"`
	Payload     []byte     `gorm:"not null" json:"payload" db:"payload"`
	Status      string     `gorm:"not null;index" json:"status" db:"status"`
	Attempts    int        `gorm:"not null;default:0" json:"attempts" db:"attempts"`
	MaxRetry    int        `gorm:"not null;default:3" json:"max_retry" db:"max_retry"`
	RunAt       time.Time  `gorm:"not null;index" json:"run_at" db:"run_at"`
	LastError   *string    `gorm:"type:text" json:"last_error" db:"last_error"`
	LastErrorAt *time.Time `json:"last_error_at" db:"last_error_at"`
	FinishedAt  *time.Time `json:"finished_at" db:"finished_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
