package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/usecase/task"
)

type recurrenceDTO struct {
	Type       task.RecurrenceType `json:"type"`
	EveryNDays int                 `json:"every_n_days,omitempty"`
	DayOfMonth int                 `json:"day_of_month,omitempty"`
	StartDate  string              `json:"start_date,omitempty"`
	EndDate    string              `json:"end_date,omitempty"`
	Dates      []string            `json:"dates,omitempty"`
}

type taskCreateDTO struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	ScheduledFor string            `json:"scheduled_for,omitempty"`
	Recurrence   *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	ScheduledFor string            `json:"scheduled_for"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type taskUpdateDTO struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	ScheduledFor string            `json:"scheduled_for,omitempty"`
}

type createTasksResponse struct {
	Tasks []taskDTO `json:"tasks"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		ScheduledFor: task.ScheduleFor.Format(time.DateOnly),
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}
