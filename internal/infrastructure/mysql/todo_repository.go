package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
)

type TodoRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewTodoRepository(db *sql.DB, logger logger.Logger) repository.TodoRepository {
	return &TodoRepository{db: db, logger: logger}
}

func (r *TodoRepository) BeginTx(ctx context.Context) (repository.Tx, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{}) // read-write
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx}, nil
}

// ---- CreateTx (used by the outbox pattern) ----
func (r *TodoRepository) CreateTx(ctx context.Context, tx repository.Tx, t *domain.TodoItem) error {
	sqltx := tx.(*sqlTx).Tx

	// Adjust table and column names to your schema.
	// Assuming: todos(id, description, due_date, file_id, created_at, updated_at)
	_, err := sqltx.ExecContext(ctx, `
		INSERT INTO todos (id, description, due_date, file_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		t.ID.String(),
		t.Description,
		toNullTime(t.DueDate),
		nullOrString(t.FileID),
		t.CreatedAt.UTC(),
		t.UpdatedAt.UTC(),
	)
	return err
}

func nullOrString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func toNullTime(tt time.Time) any {
	if tt.IsZero() {
		return nil
	}
	return tt.UTC()
}
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

func (r *TodoRepository) ListPaged(
	ctx context.Context,
	f domain.TodoFilter,
	s domain.TodoSort,
	limit, offset int,
) ([]*domain.TodoItem, int64, error) {
	args := []any{}
	conds := []string{"1=1"}

	if f.Q != nil && *f.Q != "" {
		conds = append(conds, "description LIKE ?")
		args = append(args, "%"+*f.Q+"%")
	}
	if f.DueFrom != nil {
		conds = append(conds, "due_date >= ?")
		args = append(args, *f.DueFrom)
	}
	if f.DueTo != nil {
		conds = append(conds, "due_date <= ?")
		args = append(args, *f.DueTo)
	}
	if f.HasFile != nil {
		if *f.HasFile {
			conds = append(conds, "file_id IS NOT NULL AND file_id <> ''")
		} else {
			conds = append(conds, "(file_id IS NULL OR file_id = '')")
		}
	}

	field := "updated_at"
	switch s.Field {
	case domain.SortCreatedAt:
		field = "created_at"
	case domain.SortDueDate:
		field = "due_date"
	case domain.SortUpdatedAt:
		field = "updated_at"
	case domain.SortDescription:
		field = "description"
	}
	dir := "DESC"
	if s.Direction == domain.SortAsc {
		dir = "ASC"
	}
	orderBy := field + " " + dir

	where := strings.Join(conds, " AND ")

	var total int64
	countSQL := "SELECT COUNT(*) FROM todo_items WHERE " + where
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		r.logger.Error("ListPaged: count failed", err)
		return nil, 0, err
	}

	pageSQL := `
		SELECT id, description, due_date, file_id, created_at, updated_at
		FROM todo_items
		WHERE ` + where + `
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?`
	pageArgs := append(append([]any{}, args...), limit, offset)

	rows, err := r.db.QueryContext(ctx, pageSQL, pageArgs...)
	if err != nil {
		r.logger.Error("ListPaged: query failed", err)
		return nil, 0, err
	}
	defer rows.Close()

	var out []*domain.TodoItem
	for rows.Next() {
		var t domain.TodoItem
		if err := rows.Scan(
			&t.ID,
			&t.Description,
			&t.DueDate,
			&t.FileID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

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
