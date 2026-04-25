package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at, frequency, interval, next_due_date, day_of_month, odd_even_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, title, description, status, created_at, updated_at, frequency, interval, next_due_date, day_of_month, odd_even_type
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt, task.Frequency, task.Interval, task.NextDueDate, task.DayOfMonth, task.OddEvenType)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}


	//save dates if exist
	if len(task.SpecificDates) > 0 {
        if err := r.saveSpecificDates(ctx, created.ID, task.SpecificDates); err != nil {
            return nil, err
        }
        created.SpecificDates = task.SpecificDates
    }

	return created, nil
}

func (r *Repository) saveSpecificDates(ctx context.Context, taskID int64, dates []time.Time) error {
    for _, date := range dates {
        _, err := r.pool.Exec(ctx, `
            INSERT INTO task_dates (task_id, date) VALUES ($1, $2)
        `, taskID, date)
        if err != nil {
            return err
        }
    }
    return nil
}

func (r *Repository) getSpecificDates(ctx context.Context, taskID int64) ([]time.Time, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT date FROM task_dates WHERE task_id = $1 ORDER BY date
    `, taskID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var dates []time.Time
    for rows.Next() {
        var date time.Time
        if err := rows.Scan(&date); err != nil {
            return nil, err
        }
        dates = append(dates, date)
    }
    return dates, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, frequency, interval, next_due_date, day_of_month, odd_even_type
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}
	found.SpecificDates, err = r.getSpecificDates(ctx, found.ID)
	if err != nil {
    	return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4,
			frequency = $5,
			interval = $6,
			next_due_date = $7,
			day_of_month = $8,
			odd_even_type = $9
		WHERE id = $10
		RETURNING id, title, description, status, created_at, updated_at, frequency, interval, next_due_date, day_of_month, odd_even_type
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.Frequency, task.Interval, task.NextDueDate, task.DayOfMonth, task.OddEvenType, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}
	// после получения updated
	_, err = r.pool.Exec(ctx, `DELETE FROM task_dates WHERE task_id = $1`, task.ID)
	if err != nil {
   		return nil, err
	}
	if len(task.SpecificDates) > 0 {
    	if err := r.saveSpecificDates(ctx, task.ID, task.SpecificDates); err != nil {
        	return nil, err
    	}
	}
	updated.SpecificDates = task.SpecificDates

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, frequency, interval, next_due_date, day_of_month, odd_even_type
		FROM tasks
		WHERE next_due_date is NULL OR next_due_date<=CURRENT_DATE
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		task.SpecificDates, err = r.getSpecificDates(ctx, task.ID)	
		if err != nil {
        	return nil, err
    	}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}


	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.Frequency,
		&task.Interval,
		&task.NextDueDate,
		&task.DayOfMonth,
		&task.OddEvenType,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
