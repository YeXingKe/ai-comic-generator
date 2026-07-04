package main

import (
	"fmt"
	"log"

	"github.com/ai-comic-generator/server/internal/app"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/middleware"
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
	r.Use(middleware.CORS())
	if err := middleware.SetupSession(r, cfg); err != nil {
		log.Fatalf("setup session: %v", err)
	}

	api := r.Group(cfg.Server.ContextPath)
	{
		api.GET("/health", application.HealthHandler.Check)

		user := api.Group("/user")
		{
			user.POST("/register", application.UserHandler.Register)
			user.POST("/login", application.UserHandler.Login)
			user.GET("/getLoginUser", application.UserHandler.GetLoginUser)
			user.POST("/logout", application.UserHandler.Logout)

			adminAuth := middleware.AuthCheck(application.UserService, common.AdminRole)
			user.POST("/list/page/vo", adminAuth, application.UserHandler.ListPageVO)
			user.POST("/delete", adminAuth, application.UserHandler.Delete)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting at http://localhost%s%s", addr, cfg.Server.ContextPath)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
