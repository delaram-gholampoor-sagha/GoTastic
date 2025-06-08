package repository

import (
	"context"
	"database/sql"

	"github.com/delaram/GoTastic/internal/domain"
	"go.uber.org/zap"
)

// MySQLTodoRepository implements TodoRepository using MySQL
type MySQLTodoRepository struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewMySQLTodoRepository creates a new MySQL todo repository
func NewMySQLTodoRepository(logger *zap.Logger, db *sql.DB) *MySQLTodoRepository {
	return &MySQLTodoRepository{
		logger: logger,
		db:     db,
	}
}

// Create creates a new todo item
func (r *MySQLTodoRepository) Create(ctx context.Context, todo *domain.TodoItem) error {
	query := `
		INSERT INTO todo_items (id, description, due_date, file_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		todo.ID,
		todo.Description,
		todo.DueDate,
		todo.FileID,
		todo.CreatedAt,
		todo.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create todo", zap.Error(err))
		return err
	}

	return nil
}

// Get retrieves a todo item by ID
func (r *MySQLTodoRepository) Get(ctx context.Context, id string) (*domain.TodoItem, error) {
	query := `
		SELECT id, description, due_date, file_id, created_at, updated_at
		FROM todo_items
		WHERE id = ?
	`

	var todo domain.TodoItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.Description,
		&todo.DueDate,
		&todo.FileID,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get todo", zap.Error(err))
		return nil, err
	}

	return &todo, nil
}

// List retrieves all todo items
func (r *MySQLTodoRepository) List(ctx context.Context) ([]*domain.TodoItem, error) {
	query := `
		SELECT id, description, due_date, file_id, created_at, updated_at
		FROM todo_items
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to list todos", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var todos []*domain.TodoItem
	for rows.Next() {
		var todo domain.TodoItem
		err := rows.Scan(
			&todo.ID,
			&todo.Description,
			&todo.DueDate,
			&todo.FileID,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan todo", zap.Error(err))
			return nil, err
		}
		todos = append(todos, &todo)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Failed to iterate todos", zap.Error(err))
		return nil, err
	}

	return todos, nil
}

// Update updates a todo item
func (r *MySQLTodoRepository) Update(ctx context.Context, todo *domain.TodoItem) error {
	query := `
		UPDATE todo_items
		SET description = ?, due_date = ?, file_id = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Description,
		todo.DueDate,
		todo.FileID,
		todo.UpdatedAt,
		todo.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update todo", zap.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a todo item
func (r *MySQLTodoRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM todo_items
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete todo", zap.Error(err))
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
