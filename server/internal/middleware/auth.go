package middleware

import (
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/handler"
	"github.com/ai-comic-generator/server/internal/service"
	"github.com/gin-gonic/gin"
)

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
