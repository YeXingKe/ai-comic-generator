// Package router 注册 HTTP 路由（Gin）。
package router

import (
	"github.com/ai-comic-generator/server/internal/app"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Register 挂载全部 API 与静态资源（前缀见 cfg.Server.ContextPath，一般为 /api）。
func Register(r *gin.Engine, cfg *config.Config, application *app.App) {
	// Swagger UI：http://localhost:<port>/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group(cfg.Server.ContextPath)

	api.GET("/health", application.HealthHandler.Check) // 健康检查，探活与部署校验

	comicAuth := middleware.AuthCheck(application.UserService, "")
	if application.ComicHandler != nil {
		comic := api.Group("/comic")
		comic.POST("/create", comicAuth, application.ComicHandler.Create)                         // 创建漫画任务，异步生成标题候选
		comic.POST("/confirm-title", comicAuth, application.ComicHandler.ConfirmTitle)           // 用户选定标题，进入待开始状态
		comic.POST("/start", comicAuth, application.ComicHandler.Start)                          // 确认标题后启动六步生成流水线
		comic.POST("/confirm-storyboard", comicAuth, application.ComicHandler.ConfirmStoryboard) // 确认分镜后继续后续生图与排版
		comic.POST("/retry", comicAuth, application.ComicHandler.Retry)                          // 失败任务从当前阶段重试
		comic.POST("/regenerate-panel", comicAuth, application.ComicHandler.RegeneratePanel)     // 单格分镜重新生图
		comic.POST("/publish", comicAuth, application.ComicHandler.Publish)                      // 发布到公众号（或草稿降级）
		comic.GET("/get", comicAuth, application.ComicHandler.Get)                               // 按 taskId 查询漫画详情与进度
		comic.POST("/page", comicAuth, application.ComicHandler.ListPage)                        // 当前用户漫画任务分页列表
	}

	if application.CustomComicHandler != nil {
		custom := api.Group("/comic/custom")
		custom.POST("/create", comicAuth, application.CustomComicHandler.Create)                    // 创建自定义分镜漫画任务
		custom.GET("/get", comicAuth, application.CustomComicHandler.Get)                           // 查询自定义任务详情
		custom.POST("/page", comicAuth, application.CustomComicHandler.ListPage)                    // 自定义任务分页列表
		custom.GET("/download", comicAuth, application.CustomComicHandler.DownloadZip)              // 打包下载生成图片 ZIP
		custom.POST("/regenerate-panel", comicAuth, application.CustomComicHandler.RegeneratePanel) // 自定义任务单格重绘
	}

	r.Static(application.Config.Storage.PublicURL, application.Config.Storage.BasePath) // 漫画本地静态资源访问

	user := api.Group("/user")
	user.POST("/register", application.UserHandler.Register) // 注册新用户
	user.POST("/login", application.UserHandler.Login)       // 登录并写入 Session Cookie
	user.GET("/info", application.UserHandler.GetLoginUser)  // 获取当前登录用户信息（含积分）
	user.POST("/logout", application.UserHandler.Logout)     // 退出登录，清除 Session

	userAuth := middleware.AuthCheck(application.UserService, "")
	user.POST("/profile/update", userAuth, application.UserHandler.UpdateProfile)   // 更新昵称、头像、简介
	user.POST("/password/update", userAuth, application.UserHandler.UpdatePassword) // 修改登录密码

	adminAuth := middleware.AuthCheck(application.UserService, common.AdminRole)
	user.POST("/page/vo", adminAuth, application.UserHandler.ListPageVO) // 管理员分页查询用户
	user.POST("/add", adminAuth, application.UserHandler.Add)          // 管理员创建用户
	user.POST("/update", adminAuth, application.UserHandler.Update)    // 管理员更新用户（角色、积分、状态等）
	user.POST("/delete", adminAuth, application.UserHandler.Delete)    // 管理员逻辑删除用户

	statAdminAuth := middleware.AuthCheck(application.UserService, common.AdminRole)
	stat := api.Group("/stat")
	stat.GET("/dashboard", statAdminAuth, application.StatHandler.Dashboard) // 管理端数据统计看板

	payAuth := middleware.AuthCheck(application.UserService, "")
	pay := api.Group("/pay")
	pay.GET("/packages", payAuth, application.PayHandler.ListPackages)  // 充值套餐与支付能力（支付宝/模拟）
	pay.POST("/order", payAuth, application.PayHandler.Create)          // 创建充值订单，返回扫码 URL 等
	pay.GET("/order", payAuth, application.PayHandler.Get)              // 查询单笔订单状态（可触发查单补入账）
	pay.POST("/order/page", payAuth, application.PayHandler.ListPage)   // 当前用户充值订单分页
	pay.POST("/mock-pay", payAuth, application.PayHandler.MockPay)      // 开发环境模拟支付成功
	api.POST("/pay/notify/alipay", application.PayHandler.AlipayNotify) // 支付宝异步通知（无 Session，验签后入账）

	payAdmin := middleware.AuthCheck(application.UserService, common.AdminRole)
	payAdminGroup := api.Group("/pay/admin")
	payAdminGroup.POST("/package/page", payAdmin, application.PayHandler.AdminListPackagePlans)    // 充值方案分页
	payAdminGroup.POST("/package/add", payAdmin, application.PayHandler.AdminAddPackagePlan)       // 新增充值方案
	payAdminGroup.POST("/package/update", payAdmin, application.PayHandler.AdminUpdatePackagePlan) // 更新充值方案
	payAdminGroup.POST("/package/delete", payAdmin, application.PayHandler.AdminDeletePackagePlan)   // 删除充值方案
	payAdminGroup.POST("/order/page", payAdmin, application.PayHandler.AdminListOrders)              // 全站支付记录分页
	payAdminGroup.POST("/receipt/page", payAdmin, application.PayHandler.AdminListReceipts)          // 收款记录（仅已支付订单）
}
