// internal/infrastructure/mysql/todo_repository.go
package mysqlinfra

import (
	"context"
	"fmt"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/latolukasz/beeorm"
)

type todoRepository struct{ engine beeorm.Engine }

func NewTodoRepository(engine beeorm.Engine) repository.TodoRepository {
	return &todoRepository{engine: engine}
}

func (r *todoRepository) Create(ctx context.Context, todo *domain.TodoItem) error {
	flusher := r.engine.NewFlusher()
	flusher.Track(todo)
	return flusher.FlushWithCheck()
}

func (r *todoRepository) GetByID(ctx context.Context, publicID string) (*domain.TodoItem, error) {
	var todo domain.TodoItem
	where := beeorm.NewWhere("PublicID = ?", publicID)
	if ok := r.engine.SearchOne(where, &todo); !ok {
		return nil, repository.ErrNotFound
	}
	return &todo, nil
}

func (r *todoRepository) List(ctx context.Context) ([]*domain.TodoItem, error) {
	var todos []*domain.TodoItem
	where := beeorm.NewWhere("1=1")
	pager := beeorm.NewPager(1, 1000)
	r.engine.Search(where, pager, &todos)
	return todos, nil
}

func (r *todoRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.TodoItem, error) {
	var todos []*domain.TodoItem
	where := beeorm.NewWhere("1=1")
	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	pager := beeorm.NewPager(page, limit)
	r.engine.Search(where, pager, &todos)

	out := make([]domain.TodoItem, len(todos))
	for i, t := range todos {
		out[i] = *t
	}
	return out, nil
}

func (r *todoRepository) FindByID(ctx context.Context, id uint64) (domain.TodoItem, error) {
	var todo domain.TodoItem
	if !r.engine.LoadByID(id, &todo) {
		return todo, fmt.Errorf("todo with id %d not found", id)
	}
	return todo, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *domain.TodoItem) error {
	var existing domain.TodoItem
	if ok := r.engine.SearchOne(beeorm.NewWhere("PublicID = ?", todo.PublicID), &existing); !ok {
		return repository.ErrNotFound
	}

	existing.Description = todo.Description
	existing.DueDate = todo.DueDate
	existing.FileID = todo.FileID
	existing.UpdatedAt = todo.UpdatedAt

	flusher := r.engine.NewFlusher()
	flusher.Track(&existing)
	return flusher.FlushWithCheck()
}

func (r *todoRepository) Delete(ctx context.Context, publicID string) error {
	var todo domain.TodoItem
	if ok := r.engine.SearchOne(beeorm.NewWhere("PublicID = ?", publicID), &todo); !ok {
		return repository.ErrNotFound
	}
	flusher := r.engine.NewFlusher()
	flusher.Delete(&todo)
	return flusher.FlushWithCheck()
}
