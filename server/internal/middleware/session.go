package middleware

import (
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func SetupSession(r *gin.Engine, cfg *config.ServerConfig) {
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: 2,
	})
	r.Use(sessions.Sessions("ai_comic_session", store))
}
