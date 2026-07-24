package model

// StatRange 统计时间范围
type StatRange string

const (
	StatRange7d  StatRange = "7d"  // 最近 7 天
	StatRange30d StatRange = "30d" // 最近 30 天
	StatRange90d StatRange = "90d" // 最近 90 天
)

// StatQueryRequest 统计查询请求
type StatQueryRequest struct {
	Range StatRange `form:"range" binding:"required"` // 时间范围：7d / 30d / 90d
}

// StatBucket 通用聚合桶，用于饼图 / 柱状图 / 漏斗图
type StatBucket struct {
	Key   string `json:"key"`   // 维度键，如状态码、风格；前端用于映射颜色和逻辑
	Label string `json:"label"` // 展示标签，如"已完成"、"卡通风格"
	Value int    `json:"value"` // 数量
}

// StatTrendPoint 按天的时间序列点，用于折线图
type StatTrendPoint struct {
	Date           string  `json:"date"`           // 日期，格式 YYYY-MM-DD
	Count          int     `json:"count"`          // 当日创作数
	Completed      int     `json:"completed"`      // 当日完成数
	AvgDurationMin float64 `json:"avgDurationMin"` // 当日平均耗时（分钟），无完成则为 0
}

// StatOverview KPI 概览，对应页面顶部的指标卡片
type StatOverview struct {
	TotalComics     int        `json:"totalComics"`     // 总创作数
	CompletedComics int        `json:"completedComics"` // 完成数
	CompletionRate  float64    `json:"completionRate"`  // 完成率 0-1，前端显示为百分比
	TotalUsers      int        `json:"totalUsers"`      // 总用户数（全量，不受 range 影响）
	WeeklyNewComics int        `json:"weeklyNewComics"` // 本周新增创作（固定取最近 7 天）
	AvgDurationMin  float64    `json:"avgDurationMin"`  // 平均耗时（分钟）
	Deltas          StatDeltas `json:"deltas"`          // 各 KPI 环比变化
}

// StatDeltas KPI 环比变化（%），正负均可
type StatDeltas struct {
	TotalComics     float64 `json:"totalComics"`     // 总创作数环比
	CompletionRate  float64 `json:"completionRate"`  // 完成率环比
	TotalUsers      float64 `json:"totalUsers"`      // 总用户数环比
	WeeklyNewComics float64 `json:"weeklyNewComics"` // 本周新增环比
}

// StatDashboard 统计页聚合响应，一次返回所有图表数据
type StatDashboard struct {
	Overview            StatOverview     `json:"overview"`            // KPI 概览卡片
	Trend               []StatTrendPoint `json:"trend"`               // 创作趋势（按天，范围内每天都有，无数据补零）
	StatusDistribution  []StatBucket     `json:"statusDistribution"`  // 任务状态分布，key 为 ComicStatus 常量
	PhaseFunnel         []StatBucket     `json:"phaseFunnel"`         // 流水线阶段漏斗，按阶段顺序排列
	StyleDistribution   []StatBucket     `json:"styleDistribution"`   // 漫画风格占比，key 为 style 字段值
	RoleDistribution    []StatBucket     `json:"roleDistribution"`    // 用户角色分布，key 为 UserRole 常量
	PublishDistribution []StatBucket     `json:"publishDistribution"` // 发布状态分布，key 为 publishResult.status
}

// IsValid 判断时间范围是否合法
func (r StatRange) IsValid() bool {
	return r == StatRange7d || r == StatRange30d || r == StatRange90d
}

// ToDays 返回对应天数
func (r StatRange) ToDays() int {
	switch r {
	case StatRange7d:
		return 7
	case StatRange30d:
		return 30
	case StatRange90d:
		return 90
	default:
		return 30
	}
}
