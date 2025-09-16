package mysql

import (
	"context"
	"strings"
	"time"

	"git.ice.global/packages/beeorm/v4"
	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/pkg/logger"
)

type TodoRepository struct {
	engine *beeorm.Engine
	logger logger.Logger
}

func NewTodoRepository(engine *beeorm.Engine, logger logger.Logger) repository.TodoRepository {
	return &TodoRepository{engine: engine, logger: logger}
}

func (r *TodoRepository) BeginTx(ctx context.Context) (repository.Tx, error) {
	// Reuses beeTx from the same package (defined in outbox repo file)
	db := r.engine.GetMysql()
	db.Begin()
	return &beeTx{db: db}, nil
}

// ---- CreateTx (kept simple; uses BeeORM flusher like your pattern) ----
func (r *TodoRepository) CreateTx(ctx context.Context, _ repository.Tx, t *domain.TodoItem) error {
	// If you truly need domain+outbox atomic, pass a shared Flusher/UoW instead.
	fl := r.engine.NewFlusher()
	fl.Track(t)
	return fl.FlushWithCheck()
}

func (r *TodoRepository) Create(ctx context.Context, todo *domain.TodoItem) error {
	fl := r.engine.NewFlusher()
	fl.Track(todo)
	return fl.FlushWithCheck()
}

func (r *TodoRepository) GetByID(ctx context.Context, id string) (*domain.TodoItem, error) {
	var todo domain.TodoItem
	// use real column name
	if ok := r.engine.SearchOne(beeorm.NewWhere("UUID = ?", id), &todo); !ok {
		return nil, repository.ErrNotFound
	}
	return &todo, nil
}

func (r *TodoRepository) List(ctx context.Context) ([]*domain.TodoItem, error) {
	var todos []*domain.TodoItem
	// If your BeeORM build doesn’t allow ORDER BY in Where, remove it or switch to DB.Query.
	where := beeorm.NewWhere("1=1 ORDER BY DueDate ASC")
	pager := beeorm.NewPager(1, 1000) // cap; adjust as needed
	r.engine.Search(where, pager, &todos)
	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, todo *domain.TodoItem) error {
	var existing domain.TodoItem
	if ok := r.engine.SearchOne(beeorm.NewWhere("UUID = ?", todo.UUID), &existing); !ok {
		return repository.ErrNotFound
	}

	existing.Description = todo.Description
	existing.DueDate = todo.DueDate
	existing.FileID = todo.FileID
	existing.UpdatedAt = time.Now().UTC()

	fl := r.engine.NewFlusher()
	fl.Track(&existing)
	return fl.FlushWithCheck()
}

func (r *TodoRepository) ListPaged(
	ctx context.Context,
	f domain.TodoFilter,
	s domain.TodoSort,
	limit, offset int,
) ([]*domain.TodoItem, int64, error) {
	// WHERE
	conds := []string{"1=1"}
	args := []any{}

	if f.Q != nil && *f.Q != "" {
		conds = append(conds, "Description LIKE ?")
		args = append(args, "%"+*f.Q+"%")
	}
	if f.DueFrom != nil {
		conds = append(conds, "DueDate >= ?")
		args = append(args, *f.DueFrom)
	}
	if f.DueTo != nil {
		conds = append(conds, "DueDate <= ?")
		args = append(args, *f.DueTo)
	}
	if f.HasFile != nil {
		if *f.HasFile {
			conds = append(conds, "FileID IS NOT NULL AND FileID <> ''")
		} else {
			conds = append(conds, "(FileID IS NULL OR FileID = '')")
		}
	}

	// SORT
	field := "UpdatedAt"
	switch s.Field {
	case domain.SortCreatedAt:
		field = "CreatedAt"
	case domain.SortDueDate:
		field = "DueDate"
	case domain.SortUpdatedAt:
		field = "UpdatedAt"
	case domain.SortDescription:
		field = "Description"
	}
	dir := "DESC"
	if s.Direction == domain.SortAsc {
		dir = "ASC"
	}
	orderBy := field + " " + dir

	// PAGE calc
	if limit <= 0 {
		limit = 50
	}
	page := offset/limit + 1

	// COUNT
	whereSQL := strings.Join(conds, " AND ")
	db := r.engine.GetMysql()

	var total int64
	{
		rows, close := db.Query("SELECT COUNT(*) FROM TodoItem WHERE "+whereSQL, args...)
		defer close()
		if rows.Next() {
			rows.Scan(&total)
		}
	}

	// PAGE rows via BeeORM (params passed at construction; no Bind)
	var todos []*domain.TodoItem
	where := beeorm.NewWhere(whereSQL+" ORDER BY "+orderBy, args...)
	pager := beeorm.NewPager(page, limit)
	r.engine.Search(where, pager, &todos)

	return todos, total, nil
}

func (r *TodoRepository) Delete(ctx context.Context, uuid string) error {
	var todo domain.TodoItem
	if ok := r.engine.SearchOne(beeorm.NewWhere("UUID = ?", uuid), &todo); !ok {
		return repository.ErrNotFound
	}
	fl := r.engine.NewFlusher()
	fl.Delete(&todo)
	return fl.FlushWithCheck()
}
