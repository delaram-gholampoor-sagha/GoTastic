package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/latolukasz/beeorm"
)

type TodoItem struct {
	beeorm.ORM `beeorm:"table=todo_items"`

	ID          uint64    `beeorm:"id" json:"-"`
	PublicID    string    `beeorm:"unique" json:"id"`
	Description string    `beeorm:"required" json:"description"`
	DueDate     time.Time `beeorm:"required" json:"due_date"`
	FileID      string    `beeorm:"null" json:"file_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTodoItem(description string, dueDate time.Time, fileID string) *TodoItem {
	now := time.Now()
	return &TodoItem{
		PublicID:    uuid.NewString(),
		Description: description,
		DueDate:     dueDate,
		FileID:      fileID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *TodoItem) Validate() error {
	if t.Description == "" {
		return NewError("description is required")
	}
	if t.DueDate.IsZero() {
		return NewError("due date is required")
	}
	return nil
}

type Error struct {
	message string
}

func NewError(message string) error {
	return &Error{message: message}
}

func (e *Error) Error() string {
	return e.message
}
