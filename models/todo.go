package models

import (
	"time"

	"gorm.io/gorm"
)

// Status is the lifecycle state of a todo.
type Status string

// Priority is the importance level of a todo.
type Priority string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"

	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// ValidStatuses / ValidPriorities are the allowed enum values, used by
// request validation.
var (
	ValidStatuses   = []string{string(StatusPending), string(StatusInProgress), string(StatusCompleted)}
	ValidPriorities = []string{string(PriorityLow), string(PriorityMedium), string(PriorityHigh)}
)

// IsValid reports whether s is an allowed status value.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	}
	return false
}

// IsValid reports whether p is an allowed priority value.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}

// Todo belongs to a single User. Deletes are soft: gorm.DeletedAt records the
// deletion time and GORM automatically excludes soft-deleted rows from queries.
type Todo struct {
	ID          uint           `json:"id"          gorm:"primaryKey"`
	UserID      uint           `json:"user_id"     gorm:"not null;index"`
	Title       string         `json:"title"       gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      Status         `json:"status"      gorm:"type:enum('pending','in_progress','completed');not null;default:pending"`
	Priority    Priority       `json:"priority"    gorm:"type:enum('low','medium','high');not null;default:medium"`
	DueDate     *Date          `json:"due_date"    gorm:"type:date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"           gorm:"index"`

	Categories []Category `json:"categories,omitempty" gorm:"many2many:todo_categories;constraint:OnDelete:CASCADE"`
}

// TableName pins the table name.
func (Todo) TableName() string { return "todos" }
