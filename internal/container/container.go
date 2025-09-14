package container

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/delaram/GoTastic/internal/boot"
	"github.com/delaram/GoTastic/internal/delivery/http"
	mysqlinfra "github.com/delaram/GoTastic/internal/infrastructure/mysql"
	redisinfra "github.com/delaram/GoTastic/internal/infrastructure/redis"
	s3infra "github.com/delaram/GoTastic/internal/infrastructure/s3"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/delaram/GoTastic/pkg/config"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/delaram/GoTastic/pkg/middleware"
	"github.com/delaram/GoTastic/pkg/response"
	"github.com/delaram/GoTastic/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	graphqlDelivery "github.com/delaram/GoTastic/internal/delivery/graphql"
)

type Container struct {
	Config          *config.Config
	Logger          logger.Logger
	DB              *sql.DB
	Redis           *redis.Client
	S3Client        *s3.Client
	TodoRepo        repository.TodoRepository
	FileRepo        repository.FileRepository
	CacheRepo       repository.CacheRepository
	TodoUseCase     *usecase.TodoUseCase
	OutboxRepo      repository.OutboxRepository
	StreamPublisher repository.StreamPublisher
	Router          *gin.Engine
	Response        *response.Response
}

func NewContainer(cfg *config.Config) (*Container, error) {
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})

	validator.Init()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&multiStatements=true&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := boot.RunMigrations(db, cfg.Database.Name, log); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	s3Client, err := s3infra.NewS3Client(
		cfg.S3.Endpoint,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	todoRepo := mysqlinfra.NewTodoRepository(db, log)
	outboxRepo := mysqlinfra.NewOutboxRepo(db)
	fileRepo := s3infra.NewFileRepository(s3Client, cfg.S3.Bucket)
	cacheRepo := repository.NewRedisCacheRepository(log, redisClient)
	streamPublisher := redisinfra.NewStreamPublisher(redisClient, log)

	todoUseCase := usecase.NewTodoUseCase(log, todoRepo, fileRepo, cacheRepo, streamPublisher, outboxRepo)

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(),
	)

	fileUseCase := usecase.NewFileUseCase(log, fileRepo)
	handler := http.NewHandler(log, todoUseCase, fileUseCase)
	handler.RegisterRoutes(router)

	graphqlDelivery.RegisterGinGraphQL(router, todoUseCase, fileUseCase)

	resp := response.New(true, nil, "", nil)

	return &Container{
		Config:          cfg,
		Logger:          log,
		DB:              db,
		Redis:           redisClient,
		S3Client:        s3Client,
		TodoRepo:        todoRepo,
		FileRepo:        fileRepo,
		CacheRepo:       cacheRepo,
		TodoUseCase:     todoUseCase,
		OutboxRepo:      outboxRepo,
		StreamPublisher: streamPublisher,
		Router:          router,
		Response:        resp,
	}, nil
}

func (c *Container) Close() error {
	if err := c.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	if err := c.Redis.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	return nil
}
