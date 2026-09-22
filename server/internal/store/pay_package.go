package store

import (
	"github.com/ai-comic-generator/server/internal/model"
)

// 本文件：充值套餐（PayPackagePlan）及管理端订单联表查询，方法挂载在 PayStore 上。

// ListEnabledPackages 查询已启用的充值套餐。
//
// 参数：无。
// 返回：
//   - []model.PayPackagePlan：enabled=1 的套餐，按 sortOrder、id 升序；
//   - error：数据库查询失败时非 nil。
func (s *PayStore) ListEnabledPackages() ([]model.PayPackagePlan, error) {
	var list []model.PayPackagePlan
	err := s.db.Where("enabled = ?", 1).Order("sortOrder ASC, id ASC").Find(&list).Error
	return list, err
}

// ListAllPackages 管理端分页列出全部套餐（含未启用）。
//
// 参数：
//   - page：页码，从 1 开始；
//   - size：每页条数。
// 返回：
//   - []model.PayPackagePlan：当前页记录；
//   - int64：总条数；
//   - error：统计或查询失败时非 nil。
func (s *PayStore) ListAllPackages(page, size int64) ([]model.PayPackagePlan, int64, error) {
	q := s.db.Model(&model.PayPackagePlan{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PayPackagePlan
	err := q.Order("sortOrder ASC, id ASC").
		Offset(int((page - 1) * size)).Limit(int(size)).Find(&list).Error
	return list, total, err
}

// GetPackageByCode 按业务编码查询套餐。
//
// 参数：
//   - code：套餐唯一编码（表字段 code）。
// 返回：
//   - *model.PayPackagePlan：命中记录；
//   - error：未找到时为 gorm.ErrRecordNotFound，其它为数据库错误。
func (s *PayStore) GetPackageByCode(code string) (*model.PayPackagePlan, error) {
	var p model.PayPackagePlan
	err := s.db.Where("code = ?", code).First(&p).Error
	return &p, err
}

// GetPackageByID 按主键查询套餐。
//
// 参数：
//   - id：套餐主键。
// 返回：
//   - *model.PayPackagePlan：命中记录；
//   - error：未找到或数据库错误。
func (s *PayStore) GetPackageByID(id int64) (*model.PayPackagePlan, error) {
	var p model.PayPackagePlan
	err := s.db.Where("id = ?", id).First(&p).Error
	return &p, err
}

// CreatePackage 新增充值套餐。
//
// 参数：
//   - p：待插入的套餐实体（含 code、name、amountFen、points 等）。
// 返回：
//   - error：插入失败时非 nil；成功时 p.ID 由数据库回填。
func (s *PayStore) CreatePackage(p *model.PayPackagePlan) error {
	return s.db.Create(p).Error
}

// UpdatePackage 按主键更新套餐可编辑字段。
//
// 参数：
//   - p：须带有效 ID；更新 name、amountFen、points、sortOrder、enabled。
// 返回：
//   - error：更新失败时非 nil。
func (s *PayStore) UpdatePackage(p *model.PayPackagePlan) error {
	return s.db.Model(&model.PayPackagePlan{}).Where("id = ?", p.ID).Updates(map[string]any{
		"name":      p.Name,
		"amountFen": p.AmountFen,
		"points":    p.Points,
		"sortOrder": p.SortOrder,
		"enabled":   p.Enabled,
	}).Error
}

// DeletePackage 按主键物理删除套餐。
//
// 参数：
//   - id：套餐主键。
// 返回：
//   - error：删除失败时非 nil。
func (s *PayStore) DeletePackage(id int64) error {
	return s.db.Delete(&model.PayPackagePlan{}, id).Error
}

// CountPackageByCode 统计指定 code 的套餐数量（用于唯一性校验）。
//
// 参数：
//   - code：待检查的编码；
//   - excludeID：大于 0 时排除该 id（更新场景避免与自身冲突）。
// 返回：
//   - int64：匹配条数；
//   - error：统计失败时非 nil。
func (s *PayStore) CountPackageByCode(code string, excludeID int64) (int64, error) {
	q := s.db.Model(&model.PayPackagePlan{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

// payOrderWithUser 管理端订单联表扫描结构（订单字段 + 用户账号）。
type payOrderWithUser struct {
	model.PayOrder
	UserAccount string `gorm:"column:userAccount"`
}

// ListOrdersAdmin 管理端分页查询支付订单（联表 user）。
//
// 参数：
//   - req：分页与筛选；可选 Status、UserID、OrderNo（模糊）、UserAccount（模糊）。
// 返回：
//   - []payOrderWithUser：当前页订单及 userAccount；
//   - int64：符合条件的总条数；
//   - error：统计或 Scan 失败时非 nil。
func (s *PayStore) ListOrdersAdmin(req *model.AdminPayOrderPageRequest) ([]payOrderWithUser, int64, error) {
	q := s.db.Table("pay_order").
		Select("pay_order.*, user.userAccount").
		Joins("LEFT JOIN user ON user.id = pay_order.userId AND user.isDelete = 0")

	if req.Status != nil && *req.Status != "" {
		q = q.Where("pay_order.status = ?", *req.Status)
	}
	if req.UserID != nil && *req.UserID > 0 {
		q = q.Where("pay_order.userId = ?", *req.UserID)
	}
	if req.OrderNo != nil && *req.OrderNo != "" {
		q = q.Where("pay_order.orderNo LIKE ?", "%"+*req.OrderNo+"%")
	}
	if req.UserAccount != nil && *req.UserAccount != "" {
		q = q.Where("user.userAccount LIKE ?", "%"+*req.UserAccount+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []payOrderWithUser
	err := q.Order("pay_order.id DESC").
		Offset(int((req.PageNum - 1) * req.PageSize)).
		Limit(int(req.PageSize)).
		Scan(&list).Error
	return list, total, err
}
