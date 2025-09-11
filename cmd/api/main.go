package main

// ... imports ...
import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.ice.global/packages/hitrix"

	"github.com/delaram/GoTastic/internal/container"
	graphqlDelivery "github.com/delaram/GoTastic/internal/delivery/graphql"
	httpDelivery "github.com/delaram/GoTastic/internal/delivery/http"
	"github.com/delaram/GoTastic/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctr, err := container.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	appSecret := os.Getenv("APP_SECRET")
	if appSecret == "" {
		appSecret = "dev-insecure"
	}
	reg := hitrix.New("GoTastic", appSecret)
	_, cleanup := reg.Build()
	defer cleanup()

	// ------------------------------------------

	gqlSchema := graphqlDelivery.BuildExecutableSchema(ctr.TodoUseCase, ctr.FileUseCase)

	ginInit := func(e *gin.Engine) {
		e.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
		e.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })

		h := httpDelivery.NewHandler(ctr.Logger, ctr.TodoUseCase, ctr.FileUseCase)
		h.RegisterRoutes(e)

		graphqlDelivery.RegisterGinGraphQL(e, ctr.TodoUseCase, ctr.FileUseCase)
	}

	router := hitrix.InitGin(gqlSchema, ginInit, nil)

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("HTTP listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Server exiting")
}
