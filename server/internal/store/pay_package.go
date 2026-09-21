package store

import (
	"github.com/ai-comic-generator/server/internal/model"
)

func (s *PayStore) ListEnabledPackages() ([]model.PayPackagePlan, error) {
	var list []model.PayPackagePlan
	err := s.db.Where("enabled = ?", 1).Order("sortOrder ASC, id ASC").Find(&list).Error
	return list, err
}

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

func (s *PayStore) GetPackageByCode(code string) (*model.PayPackagePlan, error) {
	var p model.PayPackagePlan
	err := s.db.Where("code = ?", code).First(&p).Error
	return &p, err
}

func (s *PayStore) GetPackageByID(id int64) (*model.PayPackagePlan, error) {
	var p model.PayPackagePlan
	err := s.db.Where("id = ?", id).First(&p).Error
	return &p, err
}

func (s *PayStore) CreatePackage(p *model.PayPackagePlan) error {
	return s.db.Create(p).Error
}

func (s *PayStore) UpdatePackage(p *model.PayPackagePlan) error {
	return s.db.Model(&model.PayPackagePlan{}).Where("id = ?", p.ID).Updates(map[string]any{
		"name":      p.Name,
		"amountFen": p.AmountFen,
		"points":    p.Points,
		"sortOrder": p.SortOrder,
		"enabled":   p.Enabled,
	}).Error
}

func (s *PayStore) DeletePackage(id int64) error {
	return s.db.Delete(&model.PayPackagePlan{}, id).Error
}

func (s *PayStore) CountPackageByCode(code string, excludeID int64) (int64, error) {
	q := s.db.Model(&model.PayPackagePlan{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

type payOrderWithUser struct {
	model.PayOrder
	UserAccount string `gorm:"column:userAccount"`
}

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
