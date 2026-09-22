package main

import (
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/app"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/middleware"
	"github.com/ai-comic-generator/server/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	defer application.Close()

	r := gin.Default()
	r.Use(middleware.CORS(cfg))
	if err := middleware.SetupSession(r, cfg); err != nil {
		log.Fatalf("setup session: %v", err)
	}

	router.Register(r, cfg, application)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting at http://localhost%s%s", addr, cfg.Server.ContextPath)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
