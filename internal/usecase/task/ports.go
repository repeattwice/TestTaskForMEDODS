package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	CreateMany(ctx context.Context, task []taskdomain.Task) ([]taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) ([]taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduleFor *time.Time
	Recurrence  *RecurrenceInput
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduleFor *time.Time
}

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeOddDays       RecurrenceType = "odd_days"
	RecurrenceTypeEvenDays      RecurrenceType = "even_days"
)

type RecurrenceInput struct {
	Type       RecurrenceType
	EveryNDays int
	DayOfMonth int
	StartDate  *time.Time
	EndDate    *time.Time
	Dates      []time.Time
}
