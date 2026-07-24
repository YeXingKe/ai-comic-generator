package store

import (
	"time"

	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/gorm"
)

// StatStore 统计数据访问层，只读，所有数据从 comic / user 表聚合计算
type StatStore struct {
	db *gorm.DB
}

// NewStatStore 创建统计 Store 实例
func NewStatStore(db *gorm.DB) *StatStore {
	return &StatStore{db: db}
}

// trendRow 趋势查询内部扫描结构（AvgDurationMin 可能为 NULL）
type trendRow struct {
	Date           string
	Count          int
	Completed      int
	AvgDurationMin *float64 // 用指针接收 NULL，无完成记录时 AVG 返回 NULL
}

// bucketRow 分组聚合内部扫描结构，用于饼图 / 柱状图 / 漏斗图
type bucketRow struct {
	Key   string // 维度键，如状态码、风格、角色
	Value int    // 该维度的数量
}

// CountComics 统计指定时间之后的总创作数
func (s *StatStore) CountComics(startTime time.Time) (int64, error) {
	var count int64
	err := s.db.Model(&model.Comic{}).  // 指定查询 comic 表
		Scopes(NotDeleted).             // 过滤软删除（isDelete = 0）
		Where("createTime >= ?", startTime). // 限定时间窗口
		Count(&count).Error             // 执行 COUNT(*)，结果写入 count
	return count, err
}

// CountCompletedComics 统计指定时间之后完成的创作数
func (s *StatStore) CountCompletedComics(startTime time.Time) (int64, error) {
	var count int64
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ? AND status = ?", startTime, model.ComicStatusCompleted). // 限定时间 + 状态为 COMPLETED
		Count(&count).Error
	return count, err
}

// AvgDurationMin 统计指定时间范围内完成任务的平均耗时（分钟），无数据返回 0
func (s *StatStore) AvgDurationMin(startTime time.Time) (float64, error) {
	var avg *float64 // 指针类型：AVG 在无数据时返回 NULL，*float64 能区分 0 和 NULL
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("status = ? AND createTime >= ?", model.ComicStatusCompleted, startTime). // 只计算已完成任务
		Select("AVG(TIMESTAMPDIFF(MINUTE, createTime, completedTime))").                // 用 TIMESTAMPDIFF 算分钟差后取平均
		Scan(&avg).Error                                                                // Scan 扫描单值，不用 Find
	if err != nil {
		return 0, err
	}
	if avg == nil { // NULL 说明范围内没有已完成任务
		return 0, nil
	}
	return *avg, nil
}

// CountUsers 统计全量用户数（不受 range 影响）
func (s *StatStore) CountUsers() (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}). // 查 user 表，全量统计
		Scopes(NotDeleted).           // 过滤已注销账号
		Count(&count).Error
	return count, err
}

// CountRecentComics 统计指定时间窗口内的新增创作数（weeklyNewComics 由 service 传入 now-7天）
func (s *StatStore) CountRecentComics(startTime time.Time) (int64, error) {
	var count int64
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ?", startTime). // startTime 由 service 控制，store 不感知"7天"含义
		Count(&count).Error
	return count, err
}

// GetTrend 按天聚合创作趋势，只返回有数据的天，补零逻辑由 service 层处理
func (s *StatStore) GetTrend(startTime time.Time) ([]model.StatTrendPoint, error) {
	var rows []trendRow
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ?", startTime).
		Select(`DATE(createTime) as date,
			COUNT(*) as count,
			COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed,
			AVG(CASE WHEN status = 'COMPLETED'
				THEN TIMESTAMPDIFF(MINUTE, createTime, completedTime) END) as avg_duration_min`). // 条件聚合：只对已完成记录算耗时
		Group("DATE(createTime)"). // 按天分组，DATE() 截断时分秒
		Order("date ASC").         // 升序排列，前端折线图从左到右
		Scan(&rows).Error          // Scan 接收多行自定义结构
	if err != nil {
		return nil, err
	}
	// 将内部扫描结构转换为对外 DTO，同时处理 NULL 耗时
	result := make([]model.StatTrendPoint, 0, len(rows))
	for _, r := range rows {
		avg := 0.0
		if r.AvgDurationMin != nil { // NULL 说明当天无完成记录，耗时置 0
			avg = *r.AvgDurationMin
		}
		result = append(result, model.StatTrendPoint{
			Date:           r.Date,
			Count:          r.Count,
			Completed:      r.Completed,
			AvgDurationMin: avg,
		})
	}
	return result, nil
}

// GetStatusDistribution 按任务状态分组统计数量，key 为 ComicStatus 常量
func (s *StatStore) GetStatusDistribution(startTime time.Time) ([]bucketRow, error) {
	var rows []bucketRow
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ?", startTime).
		Select("status as `key`, COUNT(*) as value"). // key 用反引号避免与 MySQL 保留字冲突
		Group("status").                              // 按 status 字段分组
		Scan(&rows).Error
	return rows, err
}

// GetPhaseFunnel 按流水线阶段分组统计，用于漏斗图，排序由 service 层按阶段顺序控制
func (s *StatStore) GetPhaseFunnel(startTime time.Time) ([]bucketRow, error) {
	var rows []bucketRow
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ?", startTime).
		Select("phase as `key`, COUNT(*) as value"). // 取 phase 字段作为维度键
		Group("phase").
		Scan(&rows).Error
	return rows, err
}

// GetStyleDistribution 按漫画风格分组统计
func (s *StatStore) GetStyleDistribution(startTime time.Time) ([]bucketRow, error) {
	var rows []bucketRow
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("createTime >= ?", startTime).
		Select("style as `key`, COUNT(*) as value"). // 取 style 字段（cartoon/realistic/chibi）
		Group("style").
		Scan(&rows).Error
	return rows, err
}

// GetRoleDistribution 按用户角色分组统计（全量，不受 range 影响）
func (s *StatStore) GetRoleDistribution() ([]bucketRow, error) {
	var rows []bucketRow
	err := s.db.Model(&model.User). // 查 user 表，不传 startTime
		Scopes(NotDeleted).
		Select("userRole as `key`, COUNT(*) as value"). // 取 userRole 字段（user/admin/vip）
		Group("userRole").
		Scan(&rows).Error
	return rows, err
}

// GetPublishDistribution 按发布状态分组统计，从 publishResult JSON 列提取 status 字段
func (s *StatStore) GetPublishDistribution(startTime time.Time) ([]bucketRow, error) {
	var rows []bucketRow
	err := s.db.Model(&model.Comic{}).
		Scopes(NotDeleted).
		Where("publishResult IS NOT NULL AND createTime >= ?", startTime).                          // 排除未发布（publishResult 为 NULL）的记录
		Select("JSON_UNQUOTE(JSON_EXTRACT(publishResult, '$.status')) as `key`, COUNT(*) as value"). // JSON_EXTRACT 提取嵌套字段，JSON_UNQUOTE 去掉引号
		Group("JSON_EXTRACT(publishResult, '$.status')").                                           // GROUP BY 用原始 JSON 表达式，避免别名在部分 MySQL 版本不可用
		Having("`key` IS NOT NULL").                                                                // 过滤 JSON 里 status 字段为 null 的异常记录
		Scan(&rows).Error
	return rows, err
}

// GetDB 暴露 db 给 service 层做并发查询
func (s *StatStore) GetDB() *gorm.DB {
	return s.db
}
