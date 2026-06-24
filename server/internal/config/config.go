package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	ContextPath   string `mapstructure:"context_path"`
	SessionSecret string `mapstructure:"session_secret"`
}

type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	Charset     string `mapstructure:"charset"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.context_path", "/api")
	v.SetDefault("server.session_secret", "ai-comic-generator-secret")
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.auto_migrate", true)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
