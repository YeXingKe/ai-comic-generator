package service

import (
	"strings"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
)

// reloadSalePackages 从 DB 加载上架套餐；无数据时回退 config.yaml
func (s *PayService) reloadSalePackages() {
	list, err := s.store.ListEnabledPackages()
	if err == nil && len(list) > 0 {
		m := make(map[string]model.PayPackage, len(list))
		for i := range list {
			p := list[i].ToPayPackage()
			m[p.Code] = p
		}
		s.packages = m
		return
	}
	m := make(map[string]model.PayPackage)
	for _, p := range s.cfg.Pay.Packages {
		m[p.Code] = model.PayPackage{
			Code: p.Code, Name: p.Name, AmountFen: p.AmountFen, Points: p.Points,
		}
	}
	s.packages = m
}

func (s *PayService) ListPackagePlans(pageNum, pageSize int64) (*model.PageResult, error) {
	if pageNum <= 0 {
		pageNum = common.DefaultPageNum
	}
	if pageSize <= 0 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
	}
	list, total, err := s.store.ListAllPackages(pageNum, pageSize)
	if err != nil {
		return nil, common.ErrSystem
	}
	return &model.PageResult{
		Total: total, Records: list, PageNum: pageNum, PageSize: pageSize,
	}, nil
}

func (s *PayService) AddPackagePlan(req *model.AddPayPackagePlanRequest) (int64, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return 0, common.ErrParams.WithMessage("套餐编码不能为空")
	}
	n, err := s.store.CountPackageByCode(code, 0)
	if err != nil {
		return 0, common.ErrSystem
	}
	if n > 0 {
		return 0, common.ErrOperation.WithMessage("套餐编码已存在")
	}
	plan := &model.PayPackagePlan{
		Code:      code,
		Name:      strings.TrimSpace(req.Name),
		AmountFen: req.AmountFen,
		Points:    req.Points,
		SortOrder: 0,
		Enabled:   1,
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.Enabled != nil {
		plan.Enabled = *req.Enabled
	}
	if err := s.store.CreatePackage(plan); err != nil {
		return 0, common.ErrSystem
	}
	s.reloadSalePackages()
	return plan.ID, nil
}

func (s *PayService) UpdatePackagePlan(req *model.UpdatePayPackagePlanRequest) error {
	plan, err := s.store.GetPackageByID(req.ID)
	if err != nil {
		return common.ErrNotFound
	}
	plan.Name = strings.TrimSpace(req.Name)
	plan.AmountFen = req.AmountFen
	plan.Points = req.Points
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.Enabled != nil {
		plan.Enabled = *req.Enabled
	}
	if err := s.store.UpdatePackage(plan); err != nil {
		return common.ErrSystem
	}
	s.reloadSalePackages()
	return nil
}

func (s *PayService) DeletePackagePlan(id int64) error {
	if err := s.store.DeletePackage(id); err != nil {
		return common.ErrSystem
	}
	s.reloadSalePackages()
	return nil
}

func (s *PayService) ListAdminOrders(req *model.AdminPayOrderPageRequest) (*model.PageResult, error) {
	if req.PageNum <= 0 {
		req.PageNum = common.DefaultPageNum
	}
	if req.PageSize <= 0 || req.PageSize > common.MaxPageSize {
		req.PageSize = common.DefaultPageSize
	}
	list, total, err := s.store.ListOrdersAdmin(req)
	if err != nil {
		return nil, common.ErrSystem
	}
	vos := make([]model.PayOrderAdminVO, 0, len(list))
	for i := range list {
		row := list[i]
		vo := model.PayOrderAdminVO{
			PayOrderVO:  row.PayOrder.ToVO(),
			UserID:      row.UserID,
			UserAccount: row.UserAccount,
			PackageCode: row.PackageCode,
			ChannelTxnID: row.ChannelTxnID,
		}
		vos = append(vos, vo)
	}
	return &model.PageResult{
		Total: total, Records: vos, PageNum: req.PageNum, PageSize: req.PageSize,
	}, nil
}

// ListAdminReceipts 收款记录：仅已支付订单
func (s *PayService) ListAdminReceipts(req *model.AdminPayOrderPageRequest) (*model.PageResult, error) {
	paid := model.PayPaid
	req.Status = &paid
	return s.ListAdminOrders(req)
}
