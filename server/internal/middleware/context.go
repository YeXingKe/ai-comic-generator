package middleware // 请求上下文工具：与 AuthCheck 配套，读取中间件注入的登录用户

import (
	"github.com/ai-comic-generator/server/internal/common" // 登录态 Context 键名
	"github.com/ai-comic-generator/server/internal/model"  // 用户实体
	"github.com/gin-gonic/gin"                              // Web 框架 Context
)

// GetLoginUserFromContext 从 AuthCheck 中间件注入的 Context 读取当前登录用户
// 调用方路由须先挂载 AuthCheck，否则返回 false
func GetLoginUserFromContext(c *gin.Context) (*model.User, bool) {
	value, exists := c.Get(common.LoginUserContextKey) // 读取 AuthCheck 写入的 loginUser
	if !exists { // 未挂 AuthCheck 或中间件未执行
		return nil, false // 无登录用户
	}
	user, ok := value.(*model.User) // 类型断言为 *model.User
	if !ok || user == nil {         // 类型不符或空指针
		return nil, false // 视为无效
	}
	return user, true // 返回完整用户对象（含 ID、角色等）
}
