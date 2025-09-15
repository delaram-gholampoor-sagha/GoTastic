package domain

import (
	"time"

	"git.ice.global/packages/beeorm/v4"
)

func Init(registry *beeorm.Registry) {
	registry.RegisterEntity(&TodoItem{})
	registry.RegisterEntity(&Outbox{})
}

type Outbox struct {
	beeorm.ORM    `beeorm:"table:outbox"`
	ID            uint64    `beeorm:"column:id;pk;auto_increment"`
	AggregateType string    `beeorm:"column:aggregate_type;size:64;notnull;index"`
	AggregateID   string    `beeorm:"column:aggregate_id;size:64;notnull;index"`
	EventType     string    `beeorm:"column:event_type;size:128;notnull;index"`
	Payload       []byte    `beeorm:"column:payload;type:json;notnull"`
	Headers       *string   `beeorm:"column:headers;type:json"`
	Status        string    `beeorm:"column:status;size:50;notnull;default:'pending';index"`
	Attempts      int       `beeorm:"column:attempts;default:0"`
	AvailableAt   time.Time `beeorm:"column:available_at;notnull;default:now();index"`
}

type TodoItem struct {
	beeorm.ORM  `beeorm:"table:todos"`
	ID          uint64     `beeorm:"column:id;pk;auto_increment"`
	UUID        string     `beeorm:"column:uuid;size:36;notnull;unique;index"`
	Description string     `beeorm:"column:description;size:255;notnull"`
	DueDate     *time.Time `beeorm:"column:due_date;type:datetime;index"`
	FileID      *string    `beeorm:"column:file_id;size:255;index"`
	CreatedAt   time.Time  `beeorm:"column:created_at;type:datetime;default:now()"`
	UpdatedAt   time.Time  `beeorm:"column:updated_at;type:datetime;default:now();on_update:now()"`
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

type Error struct {
	message string
}

func NewError(message string) error {
	return &Error{message: message}
}

func (e *Error) Error() string {
	return e.message
}
