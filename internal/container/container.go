package container

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/delaram/GoTastic/internal/domain"
	mysqlinfra "github.com/delaram/GoTastic/internal/infrastructure/mysql"
	redisinfra "github.com/delaram/GoTastic/internal/infrastructure/redis"
	s3infra "github.com/delaram/GoTastic/internal/infrastructure/s3"
	"github.com/delaram/GoTastic/internal/repository"
	"github.com/delaram/GoTastic/internal/usecase"
	"github.com/delaram/GoTastic/pkg/config"
	"github.com/delaram/GoTastic/pkg/logger"
	"github.com/delaram/GoTastic/pkg/validator"
	"github.com/latolukasz/beeorm"
)

type Container struct {
	Config      *config.Config
	Logger      logger.Logger
	ORMEngine   beeorm.Engine
	S3Client    *s3.Client
	TodoRepo    repository.TodoRepository
	FileRepo    repository.FileRepository
	CacheRepo   repository.CacheRepository
	TodoUseCase *usecase.TodoUseCase
	FileUseCase *usecase.FileUseCase
}

func NewContainer(cfg *config.Config) (*Container, error) {
	// Logger
	log := logger.New(logger.Config{
		Level:      "info",
		TimeFormat: time.RFC3339,
		Pretty:     true,
	})

	// Validator
	validator.Init()

	// BeeORM registry
	registry := beeorm.NewRegistry()

	// MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	registry.RegisterMySQLPool(dsn, "default")

	// Redis
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	registry.RegisterRedis(redisAddr, "", cfg.Redis.DB, "default")

	// Redis stream (align with publisher)
	registry.RegisterRedisStream("todo:stream", "default", []string{"async-consumer"})

	// Entities
	registry.RegisterEntity(&domain.TodoItem{})

	// Encoding / collation
	registry.SetDefaultEncoding("utf8mb4")
	registry.SetDefaultCollate("utf8mb4_0900_ai_ci")

	// Engine
	validated, err := registry.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate BeeORM registry: %w", err)
	}
	engine := validated.CreateEngine()

	// S3
	s3Client, err := s3infra.NewS3Client(
		cfg.S3.Endpoint,
		cfg.S3.Region,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	// Repos
	todoRepo := mysqlinfra.NewTodoRepository(engine)
	fileRepo := s3infra.NewFileRepository(s3Client, cfg.S3.Bucket)
	cacheRepo := repository.NewRedisCacheRepository(log, engine.GetRedis("default"))
	streamPublisher := redisinfra.NewStreamPublisher(engine.GetRedis("default"), log)

	// Use cases
	todoUseCase := usecase.NewTodoUseCase(log, todoRepo, fileRepo, cacheRepo, streamPublisher)
	fileUseCase := usecase.NewFileUseCase(log, fileRepo)

	return &Container{
		Config:      cfg,
		Logger:      log,
		ORMEngine:   engine,
		S3Client:    s3Client,
		TodoRepo:    todoRepo,
		FileRepo:    fileRepo,
		CacheRepo:   cacheRepo,
		TodoUseCase: todoUseCase,
		FileUseCase: fileUseCase,
	}, nil
}
