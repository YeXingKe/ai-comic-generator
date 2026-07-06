package common

// Session 相关常量
const (
	UserLoginState      = "userLoginState" // Session 中存储用户 ID 的键
	LoginUserContextKey = "loginUser"      // AuthCheck 中间件写入 gin.Context 的键
	AdminRole           = "admin"
	UserRole       = "user"
	VIPRole        = "vip"
)

// 密码相关常量
const (
	PasswordSalt      = "mason" // 须与 sql/create_table.sql 初始化数据中的 MD5 盐值一致
	DefaultPassword   = "12345678"
	MinAccountLength  = 4
	MinPasswordLength = 8
)

// 分页相关常量
const (
	DefaultPageNum  = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)
