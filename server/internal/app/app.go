// Package app 负责应用级「组装」：连接基础设施、创建各层依赖并完成注入。
package app

import (
	"context"
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/client/hunyuan"
	"github.com/ai-comic-generator/server/internal/client/wechat"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/ai-comic-generator/server/internal/storage"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// App 应用程序容器
type App struct {
	Config        *config.Config
	DB            *gorm.DB
	RedisClient   *redis.Client
	HealthHandler *handler.HealthHandler
	UserHandler   *handler.UserHandler
	ComicHandler  *handler.ComicHandler
	UserService   *service.UserService
}

// New 创建并组装整个应用
func New(cfg *config.Config) (*App, error) {
	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	redisClient, err := initRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	userStore := store.NewUserStore(db)
	comicStore := store.NewComicStore(db)

	userService := service.NewUserService(userStore)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler()

	localStore, err := storage.NewLocal(&cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	hyClient, err := hunyuan.NewClient(&cfg.AI.Hunyuan)
	if err != nil {
		return nil, fmt.Errorf("init hunyuan: %w", err)
	}
	wechatClient := wechat.NewMPClient(&cfg.WeChat)

	llm, llmErr := service.NewLLM(cfg)
	if llmErr != nil {
		log.Printf("warn: dashscope llm not ready (%v), comic create disabled", llmErr)
	}

	imageSvc := service.NewImageService(cfg, localStore, hyClient, llm)
	composeSvc := service.NewComposeService(localStore)
	publishSvc := service.NewPublishService(localStore, wechatClient)

	var comicHandler *handler.ComicHandler
	if llmErr == nil {
		orchestrator := service.NewComicOrchestrator(
			llm, comicStore, imageSvc, composeSvc, publishSvc,
		)
		comicService := service.NewComicService(comicStore, orchestrator)
		comicHandler = handler.NewComicHandler(comicService)
	}

	return &App{
		Config:        cfg,
		DB:            db,
		RedisClient:   redisClient,
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		ComicHandler:  comicHandler,
		UserService:   userService,
	}, nil
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	log.Println("database connected")
	return db, nil
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	log.Println("redis connected")
	return client, nil
}

func (a *App) Close() error {
	sqlDB, err := a.DB.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}
	if a.RedisClient != nil {
		if err := a.RedisClient.Close(); err != nil {
			log.Printf("close redis: %v", err)
		}
	}
	return nil
}
