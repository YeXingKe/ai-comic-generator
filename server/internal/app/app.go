package app

import (
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/database"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/ai-comic-generator/server/internal/store"
	"gorm.io/gorm"
)

// App 应用程序容器，持有基础设施与各 Handler，供 main 注册路由时使用。
type App struct {
	Config        *config.Config
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

	userStore := store.NewUserStore(db)
	userService := service.NewUserService(userStore)
	if err := userService.EnsureAdmin(); err != nil {
		return nil, err
	}

	return &App{
		Config:        cfg,
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
