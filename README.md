# Implemented Features

Added support for task recurrence. When creating a task, you can specify the recurrence type (`frequency`) and additional parameters. The system automatically calculates `next_due_date` — the date of the next execution. The task list (`GET /tasks`) only displays tasks where `next_due_date <= today` or tasks without a recurrence schedule.

## When a task status is changed to `done`, the system automatically:

* recalculates `next_due_date` for the next period
* resets the status back to `new`

This way, the task "comes back" on its next scheduled date without the need to create it again.

# Recurrence Types

## 1. Daily (`daily`)

The task repeats every N days.

Parameter: `interval` — how often the task repeats, in days (1 or higher).

When created, `next_due_date = today`. After completion, `next_due_date = previous date + interval days`.

### JSON

```json
{
  "title": "Call patients",
  "frequency": "daily",
  "interval": 1
}
```

## 2. Monthly (`monthly`)

The task repeats on the N-th day of each month.

Parameter: `day_of_month` — day of the month, from 1 to 31.

If the selected day exceeds the number of days in a month, the last day of that month is used. For example, day 30 in February becomes February 28.

### JSON

```json
{
  "title": "Generate report",
  "frequency": "monthly",
  "day_of_month": 15
}
```

## 3. Specific Dates (`certain_date`)

The task is created only for the specified dates. The dates are stored in a separate `task_dates` table. The system automatically selects the nearest date that has not yet passed.

Parameter: `specific_dates` — an array of dates in ISO 8601 format.

### JSON

```json
{
  "title": "Inventory check",
  "frequency": "certain_date",
  "specific_dates": ["2026-05-01T00:00:00Z", "2026-09-01T00:00:00Z"]
}
```

## 4. Even/Odd Days (`odd_even`)

The task appears only on even or odd days of the month.

Parameter: `odd_even_type` — `"even"` or `"odd"`.

### JSON

```json
{
  "title": "Prepare reports",
  "frequency": "odd_even",
  "odd_even_type": "even"
}
```

# Design Decisions and Assumptions

* **`next_due_date` as the primary mechanism** — instead of generating separate records for each occurrence, a single task is stored with a `next_due_date` field. This is simpler and avoids creating unnecessary database records.
* **Filtering in LIST** — `GET /tasks` returns only currently relevant tasks (`next_due_date <= today`). Tasks without a recurrence schedule are always displayed.
* **Calculation based on the previous date** — when a task is marked as `done`, the next date is calculated from the previous `next_due_date`, rather than from the current date. This prevents the schedule from shifting if the task is completed later than scheduled.
* **The 31st day of the month** — if the 31st is selected and the current month has fewer days, the last day of the month is used. In the following month, the date returns to the 31st if the month has enough days.
* **Specific dates in a separate table** — `task_dates` is linked to `tasks` using `ON DELETE CASCADE`. When a task is deleted, all its associated dates are automatically deleted as well.
* **`done` status** — when a task is changed to `done`, its status is immediately reset to `new` and the next date is recalculated. A task never remains in the `done` state permanently — this is intentional for recurring tasks.
* **Starting from today** — when a task is created, `next_due_date` is always set to today (or to the nearest suitable day for `odd_even`). The user does not need to specify a start date, so the task immediately appears in the list and is ready to be completed.

# Database Schema

```sql
tasks (
  id, title, description, status,
  created_at, updated_at,
  frequency, interval,
  next_due_date,
  day_of_month,
  odd_even_type
)

task_dates (
  id,
  task_id → tasks(id) CASCADE,
  date
)
```





















# Что реализовано

Добавлена возможность задавать периодичность задач. При создании задачи указывается тип периодичности (frequency) и дополнительные параметры. Система автоматически вычисляет next_due_date -> дату следующего выполнения. В списке задач (GET /tasks) отображаются только те задачи, у которых next_due_date <= сегодня или периодичность не задана.

## При смене статуса задачи на done система автоматически:
- пересчитывает next_due_date на следующий период
- сбрасывает статус обратно на new

Таким образом задача "возрождается" к следующей дате без необходимости создавать её заново.

# Типы периодичности

## 1. Ежедневные (daily)

Задача повторяется каждые N дней.

Параметры: interval — раз в сколько дней (от 1 и выше).

При создании next_due_date = сегодня. После выполнения next_due_date = предыдущая дата + interval дней.

### json
```
{
  "title": "Обзвон пациентов",
  "frequency": "daily",
  "interval": 1
}
```
## 2. Ежемесячные (monthly)
Задача повторяется каждое N-е число месяца.

Параметры: day_of_month — число месяца от 1 до 31.

Если выбранное число больше количества дней в месяце — берётся последний день месяца. Например, 30 в феврале -> 28 февраля.

### json
```
{
  "title": "Формирование отчёта",
  "frequency": "monthly",
  "day_of_month": 15
}
```

## 3. Конкретные даты (certain_date)
Задача создаётся только на указанные даты. Даты хранятся в отдельной таблице task_dates. Система автоматически выбирает ближайшую дату которая ещё не прошла.

Параметры: specific_dates — массив дат в формате ISO 8601.

### json
```
{
  "title": "Инвентаризация",
  "frequency": "certain_date",
  "specific_dates": ["2026-05-01T00:00:00Z", "2026-09-01T00:00:00Z"]
}
```
## 4. Чётные/нечётные дни (odd_even)
Задача появляется только в чётные или только в нечётные дни месяца.

Параметры: odd_even_type — "even" (чётные) или "odd" (нечётные).

### json
```
{
  "title": "Формирование отчётности",
  "frequency": "odd_even",
  "odd_even_type": "even"
}
```
# Принятые решения и предположения
- next_due_date как основной механизм — вместо генерации отдельных записей на каждую дату, задача хранится одна и имеет поле next_due_date. Это проще и не создаёт лишних записей в базе.
- Фильтрация в LIST — GET /tasks возвращает только актуальные задачи (next_due_date <= сегодня). Задачи без периодичности показываются всегда.
- Пересчёт от последней даты — при done новая дата считается от предыдущего next_due_date, а не от текущего момента. Это предотвращает смещение расписания если задача выполнена не в тот день.
- 31-е число — если выбрано 31-е, а в месяце меньше дней, берётся последний день месяца. В следующем месяце дата возвращается к 31-му если дней достаточно.
- Конкретные даты в отдельной таблице — task_dates связана с tasks через ON DELETE CASCADE. При удалении задачи все её даты удаляются автоматически.
- Статус done — при смене на done статус сразу сбрасывается на new и пересчитывается дата. Задача никогда не остаётся в статусе done постоянно — это сделано намеренно для периодических задач.
- Старт с сегодняшнего дня -> при создании задачи next_due_date всегда устанавливается на сегодня (или ближайший подходящий день для odd_even). Пользователю не нужно указывать дату начала -> задача сразу появляется в списке и готова к выполнению

# Схема базы данных
sql
```
tasks (
  id, title, description, status,
  created_at, updated_at,
  frequency, interval,
  next_due_date,
  day_of_month,
  odd_even_type
)

task_dates (
  id,
  task_id → tasks(id) CASCADE,
  date
)
```
