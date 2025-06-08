package mysql

import (
	"context"
	"database/sql"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
)

// TodoRepository implements repository.TodoRepository for MySQL
type TodoRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewTodoRepository creates a new MySQL todo repository
func NewTodoRepository(db *sql.DB, logger logger.Logger) repository.TodoRepository {
	return &TodoRepository{db: db, logger: logger}
}

// Create stores a new todo item
func (r *TodoRepository) Create(ctx context.Context, todo *domain.TodoItem) error {
	query := `
		INSERT INTO todo_items (id, description, due_date, file_id)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		todo.ID,
		todo.Description,
		todo.DueDate,
		todo.FileID,
	)

	if err != nil {
		r.logger.Error("Failed to create todo", err)
	}
	return err
}

// GetByID retrieves a todo item by ID
func (r *TodoRepository) GetByID(ctx context.Context, id string) (*domain.TodoItem, error) {
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
			return nil, repository.ErrNotFound
		}
		r.logger.Error("Failed to get todo", err)
		return nil, err
	}

	return &todo, nil
}

// List retrieves all todo items
func (r *TodoRepository) List(ctx context.Context) ([]*domain.TodoItem, error) {
	query := `
		SELECT id, description, due_date, file_id, created_at, updated_at
		FROM todo_items
		ORDER BY due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to list todos", err)
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
			r.logger.Error("Failed to scan todo", err)
			return nil, err
		}

		todos = append(todos, &todo)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Failed to iterate todos", err)
		return nil, err
	}

	return todos, nil
}

// Update modifies an existing todo item
func (r *TodoRepository) Update(ctx context.Context, todo *domain.TodoItem) error {
	query := `
		UPDATE todo_items
		SET description = ?, due_date = ?, file_id = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		todo.Description,
		todo.DueDate,
		todo.FileID,
		todo.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update todo", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete removes a todo item
func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM todo_items
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete todo", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", err)
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}
