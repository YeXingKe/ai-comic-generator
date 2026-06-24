package database

import (
	"fmt"

	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if cfg.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	return db, nil
}
