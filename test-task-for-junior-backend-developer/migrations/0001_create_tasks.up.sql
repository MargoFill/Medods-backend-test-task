CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	frequency TEXT NOT NULL DEFAULT '',
    interval INT NOT NULL DEFAULT 0,
	next_due_date DATE,
	day_of_month INT NOT NULL DEFAULT 0,
	odd_even_type TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE TABLE IF NOT EXISTS task_dates (
	id BIGSERIAL PRIMARY KEY,
	task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	date DATE NOT NULL
);