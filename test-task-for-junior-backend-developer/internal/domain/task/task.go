package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type FrequencyType string

const (
	FrequencyDaily 		FrequencyType = "daily"
	FrequencyMonthly 	FrequencyType = "monthly" //once per month
	FrequencyCertainDate FrequencyType = "certain_date" //choosing certain date
	FrequencyOddEven     FrequencyType = "odd_even"
)

type OddEvenType string

const(
	OddDays OddEvenType = "odd"
	EvenDays OddEvenType = "even"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Frequency 	FrequencyType `json:"frequency"`
	Interval  	int           `json:"interval"`
	NextDueDate *time.Time    `json:"next_due_date,omitempty"`
	DayOfMonth  int           `json:"day_of_month"` //for monthly freqtype (~interval)
	OddEvenType     OddEvenType	  `json:"odd_even_type,omitempty"`		
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
