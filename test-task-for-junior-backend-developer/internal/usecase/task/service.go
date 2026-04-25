package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Frequency:   normalized.Frequency,
		Interval:    normalized.Interval,
        NextDueDate: CalculateNextDueDate(normalized.Frequency, 0, normalized.DayOfMonth, normalized.OddEvenType, normalized.SpecificDates, now),
		DayOfMonth:  normalized.DayOfMonth,
		OddEvenType: normalized.OddEvenType,
		SpecificDates: normalized.SpecificDates,
	}

	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	// получаем текущую задачу
    current, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

	now := s.now()

	status := normalized.Status

	var nextDueDate *time.Time
	//nextDueDate := CalculateNextDueDate(normalized.Frequency, normalized.Interval, normalized.DayOfMonth, normalized.OddEvenType, normalized.SpecificDates, now)
	if status == taskdomain.StatusDone && current.NextDueDate != nil {
		from := *current.NextDueDate
		if normalized.Frequency == taskdomain.FrequencyOddEven {
			from = from.AddDate(0, 0, 1)
		}
		nextDueDate = CalculateNextDueDate(normalized.Frequency, normalized.Interval, normalized.DayOfMonth, normalized.OddEvenType, normalized.SpecificDates, from)
		status = taskdomain.StatusNew
	} else{
		 nextDueDate =  current.NextDueDate
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      status,
		UpdatedAt:   now,
		Frequency:   normalized.Frequency,
		Interval:    normalized.Interval,
		NextDueDate: nextDueDate,
		DayOfMonth:  normalized.DayOfMonth,
		OddEvenType: normalized.OddEvenType,
		SpecificDates: normalized.SpecificDates,
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}



func CalculateNextDueDate(frequency taskdomain.FrequencyType, interval int, day_of_month int, odd_even_type taskdomain.OddEvenType, specificDates []time.Time, from time.Time) *time.Time{
	switch frequency{
	case taskdomain.FrequencyDaily:
		t:=from.AddDate(0,0,interval)
		return &t

	case taskdomain.FrequencyMonthly:
		t := LastDayIfExceeds(from.Year(), from.Month(), day_of_month, from.Location()) //проверяем текущий месяц 
		if 	!t.Before(from){
			t= LastDayIfExceeds(from.Year(), from.Month()+1, day_of_month, from.Location())//берем след месяц, если число уже прошло
		}
		return &t
	case taskdomain.FrequencyOddEven:
		t := from
		for{
			day := t.Day()
			if odd_even_type == taskdomain.EvenDays && day%2 == 0{
				return &t
			}
			if odd_even_type == taskdomain.OddDays && day%2 != 0{
				return &t
			}
			t=t.AddDate(0, 0, 1)
		}

	case taskdomain.FrequencyCertainDate:
		if len(specificDates) == 0 {
			return nil
		}
		// ищем ближайшую дату которая >= сегодня
		for _, date := range specificDates {
			if !date.Before(from) {
				return &date
			}
		}
		return nil // все даты прошли

	default:
		return nil
	}
}



//берем последний день месяца, если число, которое было выбрано > кол-во дней в след месяце
func LastDayIfExceeds(year int, month time.Month, day int, loc *time.Location) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
	if day > lastDay{
		day=lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Frequency == taskdomain.FrequencyMonthly && (input.DayOfMonth < 1 || input.DayOfMonth > 31) {
    	return CreateInput{}, fmt.Errorf("%w: day_of_month must be between 1 and 31", ErrInvalidInput)
	}
	if input.Frequency == taskdomain.FrequencyOddEven && input.OddEvenType != taskdomain.OddDays && input.OddEvenType != taskdomain.EvenDays {
    	return CreateInput{}, fmt.Errorf("%w: odd_even_type must be 'odd' or 'even'", ErrInvalidInput)
	}
	if input.Frequency == taskdomain.FrequencyCertainDate && len(input.SpecificDates) == 0 {
		return CreateInput{}, fmt.Errorf("%w: specific_dates is required for certain_date frequency", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Frequency == taskdomain.FrequencyMonthly && (input.DayOfMonth < 1 || input.DayOfMonth > 31) {
    	return UpdateInput{}, fmt.Errorf("%w: day_of_month must be between 1 and 31", ErrInvalidInput)
	}
	if input.Frequency == taskdomain.FrequencyOddEven && input.OddEvenType != taskdomain.OddDays && input.OddEvenType != taskdomain.EvenDays {
    	return UpdateInput{}, fmt.Errorf("%w: odd_even_type must be 'odd' or 'even'", ErrInvalidInput)
	}
	if input.Frequency == taskdomain.FrequencyCertainDate && len(input.SpecificDates) == 0 {
		return UpdateInput{}, fmt.Errorf("%w: specific_dates is required for certain_date frequency", ErrInvalidInput)
	}

	return input, nil
}
