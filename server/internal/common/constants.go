package common

// Session 相关常量
const (
	UserLoginState      = "userLoginState" // Session 中存储用户 ID 的键
	LoginUserContextKey = "loginUser"      // AuthCheck 中间件写入 gin.Context 的键
	AdminRole           = "admin"
	UserRole            = "user"
	VIPRole             = "vip"
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

// 生图后端标识（与 config.yaml 的 ai.image_backend / *.mapstructure 值一致）
const (
	ImageBackendHunyuan       = "hunyuan"
	ImageBackendOpenAIImage1K = "openai_image_1k"
	ImageBackendOpenAIImage4K = "openai_image_4k"
)

// 文案展示模式（前端创建漫画时选择，贯穿整条生成流水线）
const (
	CaptionModeNone   = "none"   // 不叠加任何文案
	CaptionModeTop    = "top"    // 顶部居中字幕（默认）
	CaptionModeBubble = "bubble" // 对话气泡（底部居中圆角气泡框）
)

// 智能体日志标识
const (
	Agent1TitleAgent      = "agent1_title_agent"
	Agent2StoryAgent      = "agent2_story_agent"
	Agent3CharacterAgent  = "agent3_character_agent"
	Agent4ScriptAgent     = "agent4_script_agent"
	Agent5ImageGeneration = "agent5_image_generation"
	Agent6LayoutCompose   = "agent6_layout_compose"
	Agent7WechatPublish   = "agent7_wechat_publish"
)
