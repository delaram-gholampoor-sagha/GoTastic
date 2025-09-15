package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	S3       S3Config
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "todo"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:4566"),
			Bucket:    getEnv("S3_BUCKET", "todo-files"),
			AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
			Region:    getEnv("S3_REGION", "us-east-1"),
		},
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "120s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.name", "todo")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("s3.endpoint", "http://localhost:4566")
	viper.SetDefault("s3.bucket", "todo-files")
	viper.SetDefault("s3.region", "us-east-1")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func LoadFromPath(fileOrDir string) (*Config, error) {
	v := viper.New()
	commonSetup(v)

	stat, err := os.Stat(fileOrDir)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", fileOrDir, err)
	}

	var file string
	if stat.IsDir() {
		// Prefer nested app subdir if present, else fallback to config.yaml in the dir
		candidate := filepath.Join(fileOrDir, "GoTastic", "config.yaml")
		if _, err := os.Stat(candidate); err == nil {
			file = candidate
		} else {
			file = filepath.Join(fileOrDir, "config.yaml")
		}
	} else {
		file = fileOrDir
	}

	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config from %s: %w", file, err)
	}

	return buildFromViper(v), nil
}

func commonSetup(v *viper.Viper) {
	// Map env like "SERVER_PORT" -> "server.port"
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaultsOn(v)
}

func setDefaultsOn(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "120s")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "3306")
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "todo")

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("s3.endpoint", "http://localhost:4566")
	v.SetDefault("s3.bucket", "todo-files")
	v.SetDefault("s3.region", "us-east-1")
	v.SetDefault("s3.access_key", "minioadmin")
	v.SetDefault("s3.secret_key", "minioadmin")
}

// buildFromViper creates the final Config, supporting either:
// - server.read_timeout / write_timeout / idle_timeout
// - OR a single server.timeout used for all three (fallback)
func buildFromViper(v *viper.Viper) *Config {
	// support optional "server.timeout"
	globalTimeout := v.GetDuration("server.timeout")

	readTO := v.GetDuration("server.read_timeout")
	if readTO == 0 && globalTimeout > 0 {
		readTO = globalTimeout
	}
	if readTO == 0 {
		readTO = 5 * time.Second
	}

	writeTO := v.GetDuration("server.write_timeout")
	if writeTO == 0 && globalTimeout > 0 {
		writeTO = globalTimeout
	}
	if writeTO == 0 {
		writeTO = 10 * time.Second
	}

	idleTO := v.GetDuration("server.idle_timeout")
	if idleTO == 0 && globalTimeout > 0 {
		idleTO = globalTimeout
	}
	if idleTO == 0 {
		idleTO = 120 * time.Second
	}

	return &Config{
		Server: ServerConfig{
			Port:         v.GetString("server.port"),
			ReadTimeout:  readTO,
			WriteTimeout: writeTO,
			IdleTimeout:  idleTO,
		},
		Database: DatabaseConfig{
			Host:     v.GetString("database.host"),
			Port:     v.GetString("database.port"),
			User:     v.GetString("database.user"),
			Password: v.GetString("database.password"),
			Name:     v.GetString("database.name"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("redis.host"),
			Port:     v.GetString("redis.port"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		S3: S3Config{
			Endpoint:  v.GetString("s3.endpoint"),
			Bucket:    v.GetString("s3.bucket"),
			AccessKey: v.GetString("s3.access_key"),
			SecretKey: v.GetString("s3.secret_key"),
			Region:    v.GetString("s3.region"),
		},
	}
}
