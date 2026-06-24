package app

import (
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/database"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"gorm.io/gorm"
)

type App struct {
	DB            *gorm.DB
	UserService   *service.UserService
	HealthHandler *handler.HealthHandler
	UserHandler   *handler.UserHandler
}

func New(cfg *config.Config) (*App, error) {
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		return nil, err
	}

	userService := service.NewUserService(db)
	if err := userService.EnsureAdmin(); err != nil {
		return nil, err
	}

	return &App{
		DB:            db,
		UserService:   userService,
		HealthHandler: handler.NewHealthHandler(),
		UserHandler:   handler.NewUserHandler(userService),
	}, nil
}

func (a *App) Close() {
	sqlDB, err := a.DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
