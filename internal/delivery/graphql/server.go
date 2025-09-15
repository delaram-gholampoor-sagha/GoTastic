package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/99designs/gqlgen generate --config internal/delivery/graphql/gqlgen.yml --verbose

func NewHandlers(todoUC *usecase.TodoUseCase, fileUC *usecase.FileUseCase) (playgroundH http.Handler, gqlH http.Handler) {
	es := NewExecutableSchema(Config{Resolvers: &Resolver{TodoUC: todoUC, FileUC: fileUC}})
	return playground.Handler("GraphQL Playground", "/graphql/query"), handler.NewDefaultServer(es)
}

func RegisterGinGraphQL(r *gin.Engine, todoUC *usecase.TodoUseCase, fileUC *usecase.FileUseCase) {
	pg, gql := NewHandlers(todoUC, fileUC)
	g := r.Group("/graphql")
	{
		// Playground
		g.GET("", gin.WrapH(pg))
		g.GET("/", gin.WrapH(pg))

		// GraphQL POST
		g.POST("", gin.WrapH(gql))
		g.POST("/", gin.WrapH(gql))
		g.POST("/query", gin.WrapH(gql))
	}
}
