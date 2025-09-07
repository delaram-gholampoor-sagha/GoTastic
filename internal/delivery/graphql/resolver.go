package graphql

import "github.com/delaram/GoTastic/internal/usecase"

type Resolver struct {
	TodoUC *usecase.TodoUseCase
	FileUC *usecase.FileUseCase
}
