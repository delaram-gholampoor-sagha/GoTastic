package beeinfra

import (
	"fmt"

	"git.ice.global/packages/beeorm/v4"

	"github.com/delaram/GoTastic/internal/domain"
	"github.com/delaram/GoTastic/pkg/config"
)

func NewEngine(cfg *config.Config) (*beeorm.Engine, error) {
	reg := beeorm.NewRegistry()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	reg.RegisterMySQLPool(dsn, "default")

	ns := "gotastic"
	reg.RegisterRedis(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port), ns, cfg.Redis.DB, "cache")

	// Entities used anywhere in your code must be registered
	reg.RegisterEntity(&domain.TodoItem{})
	reg.RegisterEntity(&domain.Outbox{}) // <-- you load/update this via BeeORM

	reg.SetDefaultEncoding("utf8mb4")
	reg.SetDefaultCollate("utf8mb4_general_ci")

	validated, _, err := reg.Validate()
	if err != nil {
		return nil, err
	}
	eng := validated.CreateEngine() // type: *beeorm.Engine
	return eng, nil                 // return pointer, not interface
}
