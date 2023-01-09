package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/red-rocket-software/reminder-go/internal/app/model"
	"github.com/red-rocket-software/reminder-go/pkg/logging"
)

// StorageTodo handles database communication with PostgreSQL.
type StorageTodo struct {
	// Postgres database.PGX
	Postgres *pgxpool.Pool
	// Logrus logger
	logger *logging.Logger
}

// NewStorageTodo  return new SorageTodo with Postgres pool and logger
func NewStorageTodo(postgres *pgxpool.Pool, logger *logging.Logger) *StorageTodo {
	return &StorageTodo{Postgres: postgres, logger: logger}
}

// GetAllReminds return all todos in DB PostgreSQL
func (s *StorageTodo) GetAllReminds(ctx context.Context, fetchParams FetchParams) ([]model.Todo, int, error) {
	var reminds []model.Todo

	const sql = `SELECT "Id", "Description", "CreatedAt", "DeadlineAt", "FinishedAt", "Completed" FROM todo WHERE Id > $1  ORDER BY "CreatedAt" DESC LIMIT $2`

	rows, err := s.Postgres.Query(ctx, sql, fetchParams.Cursor, fetchParams.Limit)

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
func (s *StorageTodo) CreateRemind(ctx context.Context, todo model.Todo) error {
	var id int
	const sql = `INSERT INTO todo ("Description", "CreatedAt", "DeadlineAt") 
				 VALUES ($1, $2, $3) returning "Id"`
	row := s.Postgres.QueryRow(ctx, sql, todo.Description, todo.CreatedAt, todo.DeadlineAt)
	err := row.Scan(&id)
	if err != nil {
		s.logger.Errorf("Error create remind: %v", err)
		return err
	}
	return nil
}
func (s *StorageTodo) UpdateRemind(ctx context.Context, id string) (model.Todo, error) {
	return model.Todo{}, nil
	//TODO implement me
}
func (s *StorageTodo) DeleteRemind(ctx context.Context, id string) error {
	return nil
	//TODO implement me
}
func (s *StorageTodo) GetRemindByID(ctx context.Context, id string) (model.Todo, error) {
	return model.Todo{}, nil
	//TODO implement me
}
func (s *StorageTodo) GetComplitedReminds(ctx context.Context) ([]model.Todo, error) {
	return nil, nil
	//TODO implement me
}
func (s *StorageTodo) GetNewReminds(ctx context.Context) ([]model.Todo, error) {
	return nil, nil
	//TODO implement me
}
