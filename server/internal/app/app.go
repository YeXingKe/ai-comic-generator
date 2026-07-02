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
	db, err := initDB(cfg) // 连接 MySQL，配置连接池
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	redisClient, err := initRedis(cfg) // 连接 Redis，Session 依赖它
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
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

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN() // 从配置拼 DSN：user:pass@tcp(host:port)/dbname?...

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印 SQL 日志，便于开发调试
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB() // 获取底层 *sql.DB，用于设置连接池
	if err != nil {
		return nil, fmt.Errorf("get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns) // 空闲连接数上限
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns) // 最大打开连接数

	log.Println("database connected")
	return db, nil
}



func initRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(), // host:port
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB, // 默认 0，Session 常用独立 DB 号
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil { // 启动时探测，避免运行中才发现连不上
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Println("redis connected")
	return client, nil
}

func (a *App) Close() {
	sqlDB, err := a.DB.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
