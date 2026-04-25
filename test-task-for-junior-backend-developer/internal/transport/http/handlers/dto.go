package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`

	Frequency taskdomain.FrequencyType `json:"frequency"`
	Interval  int                      `json:"interval,omitempty"`
	DayOfMonth int					   `json:"day_of_month,omitempty"`
	OddEvenType taskdomain.OddEvenType `json:"odd_even_type,omitempty"`
	SpecificDates []time.Time 		   `json:"specific_dates,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	Frequency taskdomain.FrequencyType `json:"frequency"`
	Interval  int                      `json:"interval,omitempty"`
	NextDueDate *time.Time			   `json:"next_due_date,omitempty"`
	DayOfMonth int					   `json:"day_of_month,omitempty"`
	OddEvenType taskdomain.OddEvenType `json:"odd_even_type,omitempty"`
	SpecificDates []time.Time 		   `json:"specific_dates,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Frequency:   task.Frequency,
		Interval:    task.Interval,
		NextDueDate: task.NextDueDate,
		DayOfMonth:  task.DayOfMonth,
		OddEvenType: task.OddEvenType,
		SpecificDates: task.SpecificDates,
	}
}