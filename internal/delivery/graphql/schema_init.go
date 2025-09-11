// internal/delivery/graphql/schema_init.go
package graphql

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/delaram/GoTastic/internal/usecase"
)

// BuildExecutableSchema returns gqlgen schema using your existing Resolver.
func BuildExecutableSchema(todoUC *usecase.TodoUseCase, fileUC *usecase.FileUseCase) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: &Resolver{
			TodoUC: todoUC,
			FileUC: fileUC,
		},
	})
}
