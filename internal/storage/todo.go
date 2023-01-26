package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/red-rocket-software/reminder-go/internal/app/model"
	"github.com/red-rocket-software/reminder-go/pkg/logging"
	"github.com/red-rocket-software/reminder-go/pkg/pagination"
)

// TodoStorage handles database communication with PostgreSQL.
type TodoStorage struct {
	// Postgres database.PGX
	Postgres *pgxpool.Pool
	// Logrus logger
	logger *logging.Logger
}

// NewStorageTodo  return new SorageTodo with Postgres pool and logger
func NewStorageTodo(postgres *pgxpool.Pool, logger *logging.Logger) ReminderRepo {
	return &TodoStorage{Postgres: postgres, logger: logger}
}

type FetchParam struct {
	Limit    int
	CursorID int
}

// GetAllReminds return all todos in DB PostgreSQL
func (s *TodoStorage) GetAllReminds(ctx context.Context, fetchParams FetchParam) ([]model.Todo, int, error) {
	reminds := []model.Todo{}

	const sql = `SELECT "Id", "Description", "CreatedAt", "DeadlineAt", "FinishedAt", "Completed" FROM todo WHERE "Id" > $1  ORDER BY "CreatedAt" DESC LIMIT $2`

	rows, err := s.Postgres.Query(ctx, sql, fetchParams.CursorID, fetchParams.Limit)

	if err != nil {
		s.logger.Errorf("error get all reminds from db: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var remind model.Todo

		if err := rows.Scan(
			&remind.ID,
			&remind.Description,
			&remind.CreatedAt,
			&remind.DeadlineAt,
			&remind.FinishedAt,
			&remind.Completed,
		); err != nil {
			s.logger.Errorf("remind doesnt exist: %v", err)
			return nil, 0, err
		}
		reminds = append(reminds, remind)
	}

	var nextCursor int

	if len(reminds) > 0 {
		nextCursor = reminds[len(reminds)-1].ID
	}

	return reminds, nextCursor, nil
}

// CreateRemind  store new remind entity to DB PostgreSQL
func (s *TodoStorage) CreateRemind(ctx context.Context, todo model.Todo) (int, error) {
	var id int

	const sql = `INSERT INTO todo ("Description", "CreatedAt", "DeadlineAt") 
				 VALUES ($1, $2, $3) returning "Id"`
	row := s.Postgres.QueryRow(ctx, sql, todo.Description, todo.CreatedAt, todo.DeadlineAt)
	err := row.Scan(&id)
	if err != nil {
		s.logger.Errorf("Error create remind: %v", err)
		return 0, err
	}
	return id, nil
}

// UpdateRemind update remind, can change Description, Completed and FihishedAt if Completed = true
func (s *TodoStorage) UpdateRemind(ctx context.Context, id int, input model.TodoUpdate) error {
	const sql = `UPDATE todo SET "Description" = $1, "FinishedAt" = $2, "Completed" = $3 WHERE "Id" = $4`

	ct, err := s.Postgres.Exec(ctx, sql, input.Description, input.FinishedAt, input.Completed, id)
	if err != nil {
		s.logger.Printf("unable to update remind %v", err)
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("remind not found")
	}

	return nil
}

// UpdateStatus update Completed field
func (s *TodoStorage) UpdateStatus(ctx context.Context, id int, updateInput model.TodoUpdateStatusInput) error {
	const sql = `UPDATE todo SET "FinishedAt" = $1, "Completed" = $2 WHERE "Id" = $3`

	ct, err := s.Postgres.Exec(ctx, sql, updateInput.FinishedAt, updateInput.Completed, id)
	if err != nil {
		s.logger.Printf("unable to update status %v", err)
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("remind not found")
	}

	return nil
}

// DeleteRemind deletes remind from DB
func (s *TodoStorage) DeleteRemind(ctx context.Context, id int) error {
	const sql = `DELETE FROM todo WHERE "Id" = $1`
	res, err := s.Postgres.Exec(ctx, sql, id)

	if err != nil {
		s.logger.Errorf("error don't found remind: %v", err)
		return ErrCantFindRemindWithID
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Errorf("Error delete remind: %v", err)
		return ErrDeleteFailed
	}

	return nil
}

// GetRemindByID takes out one remind from PostgreSQL by id
func (s *TodoStorage) GetRemindByID(ctx context.Context, id int) (model.Todo, error) {
	var todo model.Todo

	const sql = `SELECT "Id", "Description", "CreatedAt", "DeadlineAt", "Completed", "FinishedAt" FROM todo
    WHERE "Id" = $1 LIMIT 1`

	row := s.Postgres.QueryRow(ctx, sql, id)

	err := row.Scan(
		&todo.ID,
		&todo.Description,
		&todo.CreatedAt,
		&todo.DeadlineAt,
		&todo.Completed,
		&todo.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Todo{}, nil
	}
	if err != nil {
		s.logger.Printf("cannot get product from database: %v\n", err)
		return model.Todo{}, errors.New("cannot get product from database")
	}

	return todo, nil
}

// GetComplitedReminds returns list of completed reminds and error
func (s *TodoStorage) GetComplitedReminds(ctx context.Context, params FetchParam) ([]model.Todo, int, error) {

	var sql = `SELECT * FROM todo WHERE "Completed" = true`

	if params.CursorID > 0 {
		sql += fmt.Sprintf(` AND "Id" < %d`, params.CursorID)
	}

	sql += fmt.Sprintf(` ORDER BY "CreatedAt" DESC LIMIT %d`, params.Limit)

	rows, err := s.Postgres.Query(ctx, sql)
	if err != nil {
		s.logger.Errorf("error to select completed reminds: %v", err)
		return nil, 0, err
	}

	reminds := []model.Todo{}
	for rows.Next() {
		var remind model.Todo

		if err := rows.Scan(
			&remind.ID,
			&remind.Description,
			&remind.CreatedAt,
			&remind.DeadlineAt,
			&remind.FinishedAt,
			&remind.Completed,
		); err != nil {
			s.logger.Errorf("remind doesn't exist: %v", err)
			return nil, 0, err
		}
		reminds = append(reminds, remind)
	}

	var nextCursor int
	if len(reminds) > 0 {
		nextCursor = reminds[len(reminds)-1].ID
	}

	return reminds, nextCursor, nil
}

// GetNewReminds get all no completed reminds from DB with pagination.
func (s *TodoStorage) GetNewReminds(ctx context.Context, params pagination.Page) ([]model.Todo, int, error) {
	sql := `SELECT * FROM todo WHERE "Completed" = false`

	//if passed cursorID we add condition to query
	if params.Cursor > 0 {
		sql += fmt.Sprintf(` AND "Id" < %d`, params.Cursor)
	}

	//always add sort and LIMIT to query
	sql += fmt.Sprintf(` ORDER BY "CreatedAt" DESC LIMIT %d`, params.Limit)

	rows, err := s.Postgres.Query(ctx, sql)
	if err != nil {
		s.logger.Errorf("error to select completed reminds: %v", err)
		return nil, 0, err
	}

	reminds := []model.Todo{}

	for rows.Next() {
		var remind model.Todo

		if err := rows.Scan(&remind.ID,
			&remind.Description,
			&remind.CreatedAt,
			&remind.DeadlineAt,
			&remind.FinishedAt,
			&remind.Completed,
		); err != nil {
			s.logger.Errorf("remind doesn't exist %v", err)
			return nil, 0, err
		}

		reminds = append(reminds, remind)
	}

	var nextCursor int
	if len(reminds) > 0 {
		nextCursor = reminds[len(reminds)-1].ID
	}

	return reminds, nextCursor, nil
}

// Truncate removes all seed data from the test database.
func (s *TodoStorage) Truncate() error {
	stmt := "TRUNCATE TABLE todo;"

	if _, err := s.Postgres.Exec(context.Background(), stmt); err != nil {
		return fmt.Errorf("truncate test database tables %v", err)
	}

	return nil
}

// SeedTodos seed todos for tests
func (s *TodoStorage) SeedTodos() ([]model.Todo, error) {
	date := time.Date(2023, time.April, 1, 1, 0, 0, 0, time.UTC)
	now := time.Now().Truncate(1 * time.Millisecond).UTC()

	todos := []model.Todo{
		{
			Description: "tes1",
			CreatedAt:   now,
			DeadlineAt:  date,
		},
		{
			Description: "tes2",
			CreatedAt:   now,
			DeadlineAt:  date,
			Completed:   true,
		},
		{
			Description: "tes3",
			CreatedAt:   now,
			DeadlineAt:  date,
		},
		{
			Description: "tes4",
			CreatedAt:   now,
			DeadlineAt:  date,
		},
		{
			Description: "tes5",
			CreatedAt:   now,
			DeadlineAt:  date,
		},
	}

	for i := range todos {
		const sql = `INSERT INTO todo ("Description", "CreatedAt", "DeadlineAt", "Completed") 
				 VALUES ($1, $2, $3, $4) returning "Id"`

		row := s.Postgres.QueryRow(context.Background(), sql, todos[i].Description, todos[i].CreatedAt, todos[i].DeadlineAt, todos[i].Completed)

		err := row.Scan(&todos[i].ID)
		if err != nil {
			s.logger.Errorf("Error create remind: %v", err)
		}

	}

	return todos, nil
}
