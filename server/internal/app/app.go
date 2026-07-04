// Package app 负责应用级「组装」：连接基础设施、创建各层依赖并完成注入。
//
// 为什么需要 app.go？
//   - main.go 只负责读配置、注册路由、启动 HTTP 服务，不应塞满 Store/Service/Handler 的创建逻辑
//   - 依赖关系复杂（DB → Store → Service → Handler，文章模块还涉及 Agent、SSE、COS 等），集中在一处便于维护
//   - 便于测试：测试代码可调用 app.New(cfg) 获得完整应用实例，无需重复写一遍 wiring
//   - 生命周期管理：Close() 统一释放 DB、Redis 等资源
package app

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/ai-comic-generator/server/internal/store"	
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// App 应用程序容器，持有基础设施与各 Handler，供 main 注册路由时使用
type App struct {
	Config      *config.Config // 全局配置（端口、DB、Redis、AI、COS 等）
	DB          *gorm.DB       // MySQL 连接，供 Store 层使用
	RedisClient *redis.Client  // Redis 连接，Session 等中间件使用

	// Handlers — HTTP 入口，main.go 将其实例绑定到 Gin 路由
	HealthHandler *handler.HealthHandler
	UserHandler   *handler.UserHandler

	// UserService 暴露给中间件（AuthCheck 需要查当前用户角色）
	UserService *service.UserService
}

// New 创建并组装整个应用：基础设施 → Store → Service → Handler
func New(cfg *config.Config) (*App, error) {
	db, err := initDB(cfg) // 连接 MySQL，配置连接池
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	redisClient, err := initRedis(cfg) // 连接 Redis，Session 依赖它
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	// --- 数据访问层：每个 Store 封装一张表/一组表的 CRUD ---
	userStore := store.NewUserStore(db)
	

	// --- 业务服务层：组合 Store，实现业务规则 ---
	userService := service.NewUserService(userStore)

	// --- 处理器层：解析 HTTP 请求，调用 Service，返回 JSON/SSE ---
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler()

	return &App{
		Config:        cfg,
		DB:            db,
		RedisClient:   redisClient,
		HealthHandler:  healthHandler,
		UserHandler:   userHandler,
		UserService:   userService,
	}, nil
}

// initDB 初始化 MySQL 连接并设置连接池参数
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

// initRedis 初始化 Redis 客户端并 Ping 验证连通性
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

// Close 进程退出时释放 DB、Redis 连接
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
