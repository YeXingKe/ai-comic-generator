package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置（Viper mapstructure 映射 yaml 蛇形字段）
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Session  SessionConfig  `mapstructure:"session"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Log      LogConfig      `mapstructure:"log"`
	AI       AIConfig       `mapstructure:"ai"`
	Storage  StorageConfig  `mapstructure:"storage"`
	COS      COSConfig      `mapstructure:"cos"`
	WeChat   WeChatConfig   `mapstructure:"wechat"`
	Pay      PayConfig      `mapstructure:"pay"`
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
	Secure   bool   `mapstructure:"secure"`
    SameSite string `mapstructure:"same_site"` // "lax" | "strict" | "none"
}


type CORSConfig struct {
    AllowOrigins []string `mapstructure:"allow_origins"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"file_path"`
}

// AIConfig 大模型与生图配置
type AIConfig struct {
	DashScope     DashScopeConfig   `mapstructure:"dashscope"`
	Hunyuan       HunyuanConfig     `mapstructure:"hunyuan"`
	ImageBackend  string            `mapstructure:"image_backend"`
	PromptLang    string            `mapstructure:"prompt_lang"` // zh（默认，通义千问）或 en（GPT）
	OpenAIImage1K OpenAIImageConfig `mapstructure:"openai_image_1k"`
	OpenAIImage4K OpenAIImageConfig `mapstructure:"openai_image_4k"`
}

// OpenAIImageConfig OpenAI 兼容生图（gpt-image-1 / gpt-image-2 等）
type OpenAIImageConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
	Size    string `mapstructure:"size"`
	Quality string `mapstructure:"quality"`
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

// Config OpenAI 图片生成配置
type GPTConfig struct {
    APIKey  string `mapstructure:"api_key"`
    Model   string `mapstructure:"model"`
    BaseURL string `mapstructure:"base_url"`
    Enabled bool   `mapstructure:"enabled"`
    Timeout int    `mapstructure:"timeout"`
    Size    string `mapstructure:"size"`    // 图片尺寸，如 "1024x1024"
    Quality string `mapstructure:"quality"` // standard | hd
}

// StorageConfig 本地漫画资源存储
type StorageConfig struct {
	BasePath   string `mapstructure:"base_path"`
	PublicURL  string `mapstructure:"public_url"`
}

// COSConfig 腾讯云 COS 图片存储
type COSConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	BucketURL string `mapstructure:"bucket_url"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
}

// WeChatConfig 微信公众号发布
type WeChatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	Enabled   bool   `mapstructure:"enabled"`
}

// PayConfig 积分充值（支付宝当面付等）
type PayConfig struct {
	MockEnabled   bool              `mapstructure:"mock_enabled"`
	NotifyBaseURL string            `mapstructure:"notify_base_url"`
	ReturnBaseURL string           `mapstructure:"return_base_url"` // 新增：电脑网站支付回跳根地址
	Packages      []PayPackageItem  `mapstructure:"packages"`
	Alipay        PayAlipayConfig   `mapstructure:"alipay"`
}

// PayPackageItem 充值套餐（与 model.PayPackage 字段一致）
type PayPackageItem struct {
	Code      string `mapstructure:"code"`
	Name      string `mapstructure:"name"`
	AmountFen int    `mapstructure:"amount_fen"`
	Points    int    `mapstructure:"points"`
}

// PayAlipayConfig 支付宝开放平台
type PayAlipayConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	AppID           string `mapstructure:"app_id"`
	PrivateKey      string `mapstructure:"private_key"`
	AlipayPublicKey string `mapstructure:"alipay_public_key"`
	Sandbox         bool   `mapstructure:"sandbox"`
	Mode            string `mapstructure:"mode"` // qrcode=当面付扫码；page=电脑网站支付
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
	if len(cfg.CORS.AllowOrigins) == 0 {
		cfg.CORS.AllowOrigins = []string{"http://localhost:5173"}
	}
	if cfg.Session.SameSite == "" {
		cfg.Session.SameSite = "lax"
	}
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
	if cfg.AI.ImageBackend == "" {
		cfg.AI.ImageBackend = "hunyuan"
	}
	if cfg.AI.OpenAIImage1K.Timeout == 0 {
		cfg.AI.OpenAIImage1K.Timeout = 120
	}
	if cfg.AI.OpenAIImage1K.Size == "" {
		cfg.AI.OpenAIImage1K.Size = "1024x1024"
	}
	if cfg.AI.OpenAIImage4K.Timeout == 0 {
		cfg.AI.OpenAIImage4K.Timeout = 120
	}
	if cfg.AI.OpenAIImage4K.Size == "" {
		cfg.AI.OpenAIImage4K.Size = "1024x1024"
	}
	if cfg.AI.PromptLang == "" {
		cfg.AI.PromptLang = "zh"
	}
	if cfg.Storage.BasePath == "" {
		cfg.Storage.BasePath = "./data/comics"
	}
	if cfg.Storage.PublicURL == "" {
		cfg.Storage.PublicURL = "/static/comics"
	}
	if len(cfg.Pay.Packages) == 0 {
		cfg.Pay.Packages = []PayPackageItem{
			{Code: "p60", Name: "入门包", AmountFen: 600, Points: 60},
			{Code: "p180", Name: "常用包", AmountFen: 1800, Points: 200},
			{Code: "p680", Name: "创作包", AmountFen: 6800, Points: 800},
		}
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
