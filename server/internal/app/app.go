// Package app 负责应用级「组装」：连接基础设施、创建各层依赖并完成注入。
package app

import (
	"context"
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/client/cos"
	"github.com/ai-comic-generator/server/internal/client/gpt"
	"github.com/ai-comic-generator/server/internal/client/hunyuan"
	"github.com/ai-comic-generator/server/internal/client/wechat"
	"github.com/ai-comic-generator/server/internal/common"
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
	Config             *config.Config
	DB                 *gorm.DB
	RedisClient        *redis.Client
	HealthHandler      *handler.HealthHandler
	UserHandler        *handler.UserHandler
	ComicHandler       *handler.ComicHandler
	CustomComicHandler *handler.CustomComicHandler
	StatHandler        *handler.StatHandler
	UserService        *service.UserService
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
	customComicStore := store.NewCustomComicStore(db)
	statStore := store.NewStatStore(db)

	userService := service.NewUserService(userStore)
	statService := service.NewStatService(statStore)
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler()
	statHandler := handler.NewStatHandler(statService)

	localStore, err := storage.NewLocal(&cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}

	hyClient, err := hunyuan.NewClient(&cfg.AI.Hunyuan)
	if err != nil {
		return nil, fmt.Errorf("init hunyuan: %w", err)
	}
	cosClient, err := cos.NewClient(&cfg.COS)
	if err != nil {
		return nil, fmt.Errorf("init cos: %w", err)
	}
	wechatClient := wechat.NewMPClient(&cfg.WeChat)

	llm, llmErr := service.NewLLM(cfg)
	if llmErr != nil {
		log.Printf("warn: dashscope llm not ready (%v), comic create disabled", llmErr)
	}

	gptClient1K, err := gpt.NewClient(&cfg.AI.OpenAIImage1K)
	if err != nil {
		return nil, fmt.Errorf("init gpt image 1k: %w", err)
	}
	gptClient4K, err := gpt.NewClient(&cfg.AI.OpenAIImage4K)
	if err != nil {
		return nil, fmt.Errorf("init gpt image 4k: %w", err)
	}
	// 生图后端注册表：用户在创建漫画时按需选择，未选择时回退到 hunyuan
	generators := map[string]service.ImageGenerator{
		common.ImageBackendHunyuan:       hyClient,
		common.ImageBackendOpenAIImage1K: gptClient1K,
		common.ImageBackendOpenAIImage4K: gptClient4K,
	}
	promptBuilder := common.NewPromptBuilder(cfg.AI.PromptLang)
	imageSvc := service.NewImageService(cfg, localStore, generators, cosClient, llm, promptBuilder)
	composeSvc := service.NewComposeService(localStore, cosClient)
	publishSvc := service.NewPublishService(localStore, wechatClient)

	var comicHandler *handler.ComicHandler
	if llmErr == nil {
		orchestrator := service.NewComicOrchestrator(
			llm, promptBuilder, comicStore, imageSvc, composeSvc,
		)
		comicService := service.NewComicService(comicStore, userStore, orchestrator, publishSvc)
		comicHandler = handler.NewComicHandler(comicService)
	}

	// 自定义创作不依赖 LLM 必配：无 LLM 时用提示词拆分兜底 + 占位/生图
	customComicService := service.NewCustomComicService(customComicStore, userStore, localStore, generators, cosClient, llm)
	customComicHandler := handler.NewCustomComicHandler(customComicService)

	return &App{
		Config:             cfg,
		DB:                 db,
		RedisClient:        redisClient,
		HealthHandler:      healthHandler,
		UserHandler:        userHandler,
		ComicHandler:       comicHandler,
		CustomComicHandler: customComicHandler,
		StatHandler:        statHandler,
		UserService:        userService,
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
