package middleware

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func AuthCheck(userService *service.UserService, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := handler.GetLoginUserID(c)
		if !ok {
			c.JSON(http.StatusOK, common.Error(40100, "未登录"))
			c.Abort()
			return
		}

		vo, err := userService.GetByID(userID)
		if err != nil {
			c.JSON(http.StatusOK, common.Error(40100, "未登录"))
			c.Abort()
			return
		}

		if requiredRole == common.AdminRole && vo.UserRole != common.AdminRole {
			c.JSON(http.StatusOK, common.Error(40101, "无权限"))
			c.Abort()
			return
		}

		c.Set("loginUser", vo)
		c.Next()
	}
}
