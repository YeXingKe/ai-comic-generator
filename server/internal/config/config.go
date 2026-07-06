package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置（Viper mapstructure 映射 yaml 蛇形字段）
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Session SessionConfig `mapstructure:"session"`
	Log     LogConfig     `mapstructure:"log"`
	AI      AIConfig      `mapstructure:"ai"`
	Storage StorageConfig `mapstructure:"storage"`
	WeChat  WeChatConfig  `mapstructure:"wechat"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	ContextPath string `mapstructure:"context_path"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type SessionConfig struct {
	Secret string `mapstructure:"secret"`
	MaxAge int    `mapstructure:"max_age"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
}

// AIConfig 大模型与生图配置
type AIConfig struct {
	DashScope DashScopeConfig `mapstructure:"dashscope"`
	Hunyuan   HunyuanConfig   `mapstructure:"hunyuan"`
}

// DashScopeConfig 通义千问（qwen-plus，OpenAI 兼容接口）
type DashScopeConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// HunyuanConfig 腾讯混元生图
type HunyuanConfig struct {
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Model     string `mapstructure:"model"`
	Enabled   bool   `mapstructure:"enabled"`
}

// StorageConfig 本地漫画资源存储
type StorageConfig struct {
	BasePath   string `mapstructure:"base_path"`
	PublicURL  string `mapstructure:"public_url"`
}

// WeChatConfig 微信公众号发布
type WeChatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	Enabled   bool   `mapstructure:"enabled"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.AI.DashScope.Model == "" {
		cfg.AI.DashScope.Model = "qwen-plus"
	}
	if cfg.AI.DashScope.BaseURL == "" {
		cfg.AI.DashScope.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if cfg.AI.Hunyuan.Region == "" {
		cfg.AI.Hunyuan.Region = "ap-guangzhou"
	}
	if cfg.AI.Hunyuan.Model == "" {
		cfg.AI.Hunyuan.Model = "hunyuan-image"
	}
	if cfg.Storage.BasePath == "" {
		cfg.Storage.BasePath = "./data/comics"
	}
	if cfg.Storage.PublicURL == "" {
		cfg.Storage.PublicURL = "/static/comics"
	}
}

func applyEnvOverrides(cfg *Config) {
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
	if val := getEnv("REDIS_HOST", ""); val != "" {
		cfg.Redis.Host = val
	}
	if val := getEnv("REDIS_PORT", ""); val != "" {
		fmt.Sscanf(val, "%d", &cfg.Redis.Port)
	}
	if val := getEnv("REDIS_PASSWORD", ""); val != "" {
		cfg.Redis.Password = val
	}
	if val := getEnv("DASHSCOPE_API_KEY", ""); val != "" {
		cfg.AI.DashScope.APIKey = val
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
