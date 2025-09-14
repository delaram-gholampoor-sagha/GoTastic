package domain

import (
	"time"

	"github.com/google/uuid"
)

type TodoItem struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	FileID      string    `json:"file_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTodoItem(description string, dueDate time.Time, fileID string) *TodoItem {
	now := time.Now()
	return &TodoItem{
		ID:          uuid.New(),
		Description: description,
		DueDate:     dueDate,
		FileID:      fileID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

type TodoFilter struct {
	Q       *string
	DueFrom *time.Time
	DueTo   *time.Time
	HasFile *bool
}

type SortField int

const (
	SortCreatedAt SortField = iota
	SortDueDate
	SortUpdatedAt
	SortDescription
)

type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

type TodoSort struct {
	Field     SortField
	Direction SortDirection
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
