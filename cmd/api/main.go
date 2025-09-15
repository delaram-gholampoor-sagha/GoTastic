package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"git.ice.global/packages/beeorm/v4"
	"git.ice.global/packages/hitrix"
	"git.ice.global/packages/hitrix/service"
	"git.ice.global/packages/hitrix/service/registry"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	graphqlDelivery "github.com/delaram/GoTastic/internal/delivery/graphql"
	httpDelivery "github.com/delaram/GoTastic/internal/delivery/http"
	"github.com/delaram/GoTastic/internal/domain"
	beeinfra "github.com/delaram/GoTastic/internal/infrastructure/beeorm"
	mysqlinfra "github.com/delaram/GoTastic/internal/infrastructure/mysql"
	s3infra "github.com/delaram/GoTastic/internal/infrastructure/s3"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/delaram/GoTastic/internal/worker"
	"github.com/delaram/GoTastic/pkg/config"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/delaram/GoTastic/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sarulabs/di"
)

func main() {
	h, deferFunc := hitrix.New("GoTastic", "your-app-secret").
		RegisterDIGlobalService(
			registry.ServiceProviderConfigDirectory("config"),
			&service.DefinitionGlobal{
				Name: "app_config",
				Build: func(ctn di.Container) (interface{}, error) {
					v, err := ctn.SafeGet("config_directory")
					if err != nil {
						return nil, fmt.Errorf("get config_directory: %w", err)
					}
					dir := v.(string)
					file := filepath.Join(dir, "GoTastic", "config.yaml")
					cfg, err := config.LoadFromPath(file)
					if err != nil {
						return nil, fmt.Errorf("load config from %s: %w", file, err)
					}
					log.Printf("Loaded config: %+v", cfg)
					return cfg, nil
				},
			},
			&service.DefinitionGlobal{
				Name: "logger",
				Build: func(ctn di.Container) (interface{}, error) {
					return logger.New(logger.Config{
						Level:      "info",
						TimeFormat: time.RFC3339,
						Pretty:     true,
					}), nil
				},
			},
			registry.ServiceProviderErrorLogger(),
			registry.ServiceProviderOrmRegistry(domain.Init),
			registry.ServiceProviderOrmEngine(),
			&service.DefinitionGlobal{
				Name: "redis_persistent",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("app_config")
					if err != nil {
						return nil, fmt.Errorf("failed to get config for redis_persistent: %w", err)
					}
					cfg, ok := val.(*config.Config)
					if !ok {
						return nil, fmt.Errorf("failed to cast config to *config.Config for redis_persistent")
					}
					rdb := redis.NewClient(&redis.Options{
						Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
					})
					if err := rdb.Ping(context.Background()).Err(); err != nil {
						return nil, fmt.Errorf("failed to ping redis for redis_persistent: %w", err)
					}
					return rdb, nil
				},
			},
			&service.DefinitionGlobal{
				Name: "s3_client",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("app_config")
					if err != nil {
						return nil, fmt.Errorf("failed to get config for s3_client: %w", err)
					}
					cfg, ok := val.(*config.Config)
					if !ok {
						return nil, fmt.Errorf("failed to cast config to *config.Config for s3_client")
					}
					s3Client, err := s3infra.NewS3Client(cfg.S3.Endpoint, cfg.S3.Region, cfg.S3.AccessKey, cfg.S3.SecretKey)
					if err != nil {
						return nil, fmt.Errorf("failed to create S3 client for s3_client: %w", err)
					}
					return s3Client, nil
				},
			},
			&service.DefinitionGlobal{
				Name: "outbox_repo",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet(service.ORMEngineGlobalService)
					if err != nil {
						return nil, fmt.Errorf("failed to get orm_engine for outbox_repo: %w", err)
					}
					log.Printf("orm_engine retrieved: %T", val)
					engine, ok := val.(*beeorm.Engine)
					if !ok {
						return nil, fmt.Errorf("failed to cast orm_engine to beeorm.Engine for outbox_repo")
					}
					return mysqlinfra.NewOutboxRepo(engine), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "stream_publisher",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet(service.ORMEngineGlobalService)
					if err != nil {
						return nil, fmt.Errorf("failed to get orm_engine for stream_publisher: %w", err)
					}
					engine, ok := val.(*beeorm.Engine)
					if !ok {
						return nil, fmt.Errorf("failed to cast orm_engine to beeorm.Engine for stream_publisher")
					}
					return beeinfra.NewStreamPublisher(engine, "todo-stream"), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "todo_repo",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet(service.ORMEngineGlobalService)
					if err != nil {
						return nil, fmt.Errorf("failed to get orm_engine_global for todo_repo: %w", err)
					}
					engine, ok := val.(*beeorm.Engine)
					if !ok {
						return nil, fmt.Errorf("failed to cast orm_engine_global to beeorm.Engine for todo_repo")
					}
					val, err = ctn.SafeGet("logger")
					if err != nil {
						return nil, fmt.Errorf("failed to get logger for todo_repo: %w", err)
					}
					logger, ok := val.(logger.Logger)
					if !ok {
						return nil, fmt.Errorf("failed to cast logger to logger.Logger for todo_repo")
					}
					return mysqlinfra.NewTodoRepository(engine, logger), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "file_repo",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("app_config")
					if err != nil {
						return nil, fmt.Errorf("failed to get config for file_repo: %w", err)
					}
					cfg, ok := val.(*config.Config)
					if !ok {
						return nil, fmt.Errorf("failed to cast config to *config.Config for file_repo")
					}
					val, err = ctn.SafeGet("s3_client")
					if err != nil {
						return nil, fmt.Errorf("failed to get s3_client for file_repo: %w", err)
					}
					s3Client, ok := val.(*s3.Client)
					if !ok {
						return nil, fmt.Errorf("failed to cast s3_client to *s3.Client for file_repo")
					}
					return s3infra.NewFileRepository(s3Client, cfg.S3.Bucket), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "cache_repo",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("logger")
					if err != nil {
						return nil, fmt.Errorf("failed to get logger for cache_repo: %w", err)
					}
					logger, ok := val.(logger.Logger)
					if !ok {
						return nil, fmt.Errorf("failed to cast logger to logger.Logger for cache_repo")
					}
					val, err = ctn.SafeGet("redis_persistent")
					if err != nil {
						return nil, fmt.Errorf("failed to get redis_persistent for cache_repo: %w", err)
					}
					rdb, ok := val.(*redis.Client)
					if !ok {
						return nil, fmt.Errorf("failed to cast redis_persistent to *redis.Client for cache_repo")
					}
					return repository.NewRedisCacheRepository(logger, rdb), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "todo_usecase",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("logger")
					if err != nil {
						return nil, fmt.Errorf("failed to get logger for todo_usecase: %w", err)
					}
					logger, ok := val.(logger.Logger)
					if !ok {
						return nil, fmt.Errorf("failed to cast logger to logger.Logger for todo_usecase")
					}
					val, err = ctn.SafeGet("todo_repo")
					if err != nil {
						return nil, fmt.Errorf("failed to get todo_repo for todo_usecase: %w", err)
					}
					todoRepo, ok := val.(repository.TodoRepository)
					if !ok {
						return nil, fmt.Errorf("failed to cast todo_repo to repository.TodoRepository for todo_usecase")
					}
					val, err = ctn.SafeGet("file_repo")
					if err != nil {
						return nil, fmt.Errorf("failed to get file_repo for todo_usecase: %w", err)
					}
					fileRepo, ok := val.(repository.FileRepository)
					if !ok {
						return nil, fmt.Errorf("failed to cast file_repo to repository.FileRepository for todo_usecase")
					}
					val, err = ctn.SafeGet("cache_repo")
					if err != nil {
						return nil, fmt.Errorf("failed to get cache_repo for todo_usecase: %w", err)
					}
					cacheRepo, ok := val.(repository.CacheRepository)
					if !ok {
						return nil, fmt.Errorf("failed to cast cache_repo to repository.CacheRepository for todo_usecase")
					}
					val, err = ctn.SafeGet("stream_publisher")
					if err != nil {
						return nil, fmt.Errorf("failed to get stream_publisher for todo_usecase: %w", err)
					}
					streamPublisher, ok := val.(repository.StreamPublisher)
					if !ok {
						return nil, fmt.Errorf("failed to cast stream_publisher to repository.StreamPublisher for todo_usecase")
					}
					val, err = ctn.SafeGet("outbox_repo")
					if err != nil {
						return nil, fmt.Errorf("failed to get outbox_repo for todo_usecase: %w", err)
					}
					outboxRepo, ok := val.(repository.OutboxRepository)
					if !ok {
						return nil, fmt.Errorf("failed to cast outbox_repo to repository.OutboxRepository for todo_usecase")
					}
					return usecase.NewTodoUseCase(logger, todoRepo, fileRepo, cacheRepo, streamPublisher, outboxRepo), nil
				},
			},
			&service.DefinitionGlobal{
				Name: "file_usecase",
				Build: func(ctn di.Container) (interface{}, error) {
					val, err := ctn.SafeGet("logger")
					if err != nil {
						return nil, fmt.Errorf("failed to get logger for file_usecase: %w", err)
					}
					logger, ok := val.(logger.Logger)
					if !ok {
						return nil, fmt.Errorf("failed to cast logger to logger.Logger for file_usecase")
					}
					val, err = ctn.SafeGet("file_repo")
					if err != nil {
						return nil, fmt.Errorf("failed to get file_repo for file_usecase: %w", err)
					}
					fileRepo, ok := val.(repository.FileRepository)
					if !ok {
						return nil, fmt.Errorf("failed to cast file_repo to repository.FileRepository for file_usecase")
					}
					return usecase.NewFileUseCase(logger, fileRepo), nil
				},
			},
		).
		Build()
	defer deferFunc()

	appCfg, ok := service.GetServiceRequired("app_config").(*config.Config)
	if !ok {
		log.Fatalf("Failed to cast config to *config.Config")
	}
	appLogger := service.GetServiceRequired("logger").(logger.Logger)

	// Start outbox consumer
	dispatcherCtx, stopDispatcher := context.WithCancel(context.Background())
	defer stopDispatcher()

	outboxRepo := service.GetServiceRequired("outbox_repo").(repository.OutboxRepository)
	streamPublisher := service.GetServiceRequired("stream_publisher").(repository.StreamPublisher)
	dispatcher := worker.NewOutboxDispatcher(outboxRepo, streamPublisher)
	go dispatcher.Run(dispatcherCtx)

	// Build deps from DI
	todoUseCase := service.GetServiceRequired("todo_usecase").(*usecase.TodoUseCase)
	fileUseCase := service.GetServiceRequired("file_usecase").(*usecase.FileUseCase)

	gqlSchema := graphqlDelivery.NewExecutableSchema(
		graphqlDelivery.Config{
			Resolvers: &graphqlDelivery.Resolver{
				TodoUC: todoUseCase,
				FileUC: fileUseCase,
			},
		},
	)

	p, err := strconv.Atoi(appCfg.Server.Port)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	h.RunServer(
		uint(p),
		gqlSchema,
		func(g *gin.Engine) {
			g.Use(
				middleware.RequestID(),
				middleware.Logger(),
				middleware.Recovery(),
				middleware.CORS(),
			)

			handler := httpDelivery.NewHandler(appLogger, todoUseCase, fileUseCase)
			handler.RegisterRoutes(g)
			graphqlDelivery.RegisterGinGraphQL(g, todoUseCase, fileUseCase)

		},
		func(srv *handler.Server) {
			// Optional: Add transports/middlewares if needed
		},
	)

	log.Println("Server exiting")
}
