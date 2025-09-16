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
	beeorm.ORM    `orm:"table=outbox"`
	ID            uint64    `orm:"pk;auto_increment"`
	AggregateType string    `orm:"size(64);index"`
	AggregateID   string    `orm:"size(64);index"`
	EventType     string    `orm:"size(128);index"`
	Payload       []byte    `orm:"type(json)"`
	Headers       *string   `orm:"type(json)"`
	Status        string    `orm:"size(50);default('pending');index"`
	Attempts      int       `orm:"default(0)"`
	AvailableAt   time.Time `orm:"default(now());index"`
}

type TodoItem struct {
	beeorm.ORM  `orm:"table=TodoItem"`
	ID          uint64     `orm:"pk;auto_increment"`
	UUID        string     `orm:"size(36);unique;index"`
	Description string     `orm:"size(255)"`
	DueDate     *time.Time `orm:"type(datetime);index"`
	FileID      *string    `orm:"size(255);index"`
	CreatedAt   time.Time  `orm:"type(datetime);default(now())"`
	UpdatedAt   time.Time  `orm:"type(datetime);default(now());on_update(now())"`
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
