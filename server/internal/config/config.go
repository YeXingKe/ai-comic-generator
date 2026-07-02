package config

import (
	"fmt"
	"os"
	"github.com/spf13/viper"
)

// 作用：把 config.yaml 里的配置读进 Go 结构体，供全项目使用。

// Config 应用配置
// Viper 用 mapstructure 把 yaml 映射到 Go 结构体
// mapstructure 标签 yaml 蛇形 ↔ Go 驼峰


// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

// SessionConfig Session配置
type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
}


// Config 应用配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Session    SessionConfig    `mapstructure:"session"`
	Log        LogConfig        `mapstructure:"log"`
}

// LoadConfig 加载配置文件
// 支持通过环境变量覆盖配置（Docker 部署时使用）
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 环境变量覆盖配置（Docker 部署时使用）
	applyEnvOverrides(&config)

	return &config, nil
}

// applyEnvOverrides 使用环境变量覆盖配置
// 参考 Spring Boot 的 ${VAR:default} 机制
func applyEnvOverrides(cfg *Config) {
	// 数据库配置
	if val := getEnv("DB_HOST", ""); val != "" {
		cfg.Database.Host = val
	}
	if val := getEnv("DB_PORT", ""); val != "" {
		fmt.Sscanf(val, "%d", &cfg.Database.Port)
	}
	if val := getEnv("DB_NAME", ""); val != "" {
		cfg.Database.Name = val
	}
	if val := getEnv("DB_USER", ""); val != "" {
		cfg.Database.User = val
	}
	if val := getEnv("DB_PASSWORD", ""); val != "" {
		cfg.Database.Password = val
	}

	// Redis 配置
	if val := getEnv("REDIS_HOST", ""); val != "" {
		cfg.Redis.Host = val
	}
	if val := getEnv("REDIS_PORT", ""); val != "" {
		fmt.Sscanf(val, "%d", &cfg.Redis.Port)
	}
	if val := getEnv("REDIS_PASSWORD", ""); val != "" {
		cfg.Redis.Password = val
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
}

// GetRedisAddr 获取 Redis 地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}