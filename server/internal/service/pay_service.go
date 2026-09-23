package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	alipayx "github.com/ai-comic-generator/server/internal/client/alipay"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/config"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
)

type PayService struct {
	cfg      *config.Config
	store    *store.PayStore
	alipay   *alipayx.Client
	packages map[string]model.PayPackage
}

func NewPayService(cfg *config.Config, st *store.PayStore, ali *alipayx.Client) *PayService {
	s := &PayService{cfg: cfg, store: st, alipay: ali, packages: map[string]model.PayPackage{}}
	s.reloadSalePackages()
	return s
}

func (s *PayService) alipayMode() string {
	mode := strings.TrimSpace(s.cfg.Pay.Alipay.Mode)
	if mode == "" {
		return model.AlipayModeQRCode
	}
	return mode
}

func (s *PayService) salePackageList() []model.PayPackage {
	list, err := s.store.ListEnabledPackages()
	if err == nil && len(list) > 0 {
		out := make([]model.PayPackage, 0, len(list))
		for i := range list {
			out = append(out, list[i].ToPayPackage())
		}
		return out
	}
	out := make([]model.PayPackage, 0, len(s.cfg.Pay.Packages))
	for _, p := range s.cfg.Pay.Packages {
		out = append(out, model.PayPackage{
			Code: p.Code, Name: p.Name, AmountFen: p.AmountFen, Points: p.Points,
		})
	}
	return out
}

func (s *PayService) GetCatalog() model.PayCatalogVO {
	s.reloadSalePackages()
	alipayOn := s.cfg.Pay.Alipay.Enabled && s.alipay.Enabled()
	return model.PayCatalogVO{
		MockEnabled:   s.cfg.Pay.MockEnabled,
		AlipayEnabled: alipayOn,
		AlipaySandbox: alipayOn && s.cfg.Pay.Alipay.Sandbox,
		Packages:      s.salePackageList(),
	}
}

func (s *PayService) CreateOrder(userID int64, req *model.CreatePayOrderRequest) (*model.CreatePayOrderVO, error) {
	pkg, ok := s.packages[req.PackageCode]
	if !ok {
		return nil, common.ErrParams.WithMessage("套餐不存在")
	}
	if req.Channel != model.ChannelAlipay && req.Channel != model.ChannelMock {
		return nil, common.ErrParams.WithMessage("不支持的渠道")
	}
	if req.Channel == model.ChannelMock && !s.cfg.Pay.MockEnabled {
		return nil, common.ErrForbidden.WithMessage("未开启模拟支付")
	}
	if req.Channel == model.ChannelAlipay && !s.cfg.Pay.Alipay.Enabled {
		return nil, common.ErrOperation.WithMessage("支付宝未配置")
	}

	payMode := s.alipayMode()
	if req.Channel == model.ChannelMock {
		payMode = model.AlipayModeQRCode
	}
	if req.Channel == model.ChannelAlipay &&
		payMode != model.AlipayModeQRCode && payMode != model.AlipayModePage {
		return nil, common.ErrParams.WithMessage("pay.alipay.mode 仅支持 qrcode 或 page")
	}

	orderNo := fmt.Sprintf("A%s%06d", time.Now().Format("20060102150405"), rand.Intn(1_000_000))
	order := &model.PayOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		Channel:     req.Channel,
		PackageCode: pkg.Code,
		AmountFen:   pkg.AmountFen,
		Points:      pkg.Points,
		Status:      model.PayPending,
		ExpireAt:    time.Now().Add(5 * time.Minute),
	}

	var payURL string
	if req.Channel == model.ChannelAlipay {
		if !s.alipay.Enabled() {
			return nil, common.ErrOperation.WithMessage("支付宝未就绪")
		}
		subject := "积分充值-" + pkg.Name
		switch payMode {
		case model.AlipayModeQRCode:
			qr, err := s.alipay.Precreate(orderNo, subject, pkg.AmountFen)
			if err != nil {
				return nil, common.ErrOperation.WithMessage("支付宝下单失败")
			}
			order.CodeURL = &qr
		case model.AlipayModePage:
			if strings.TrimSpace(s.cfg.Pay.ReturnBaseURL) == "" {
				return nil, common.ErrOperation.WithMessage("未配置 return_base_url")
			}
			ret := strings.TrimRight(s.cfg.Pay.ReturnBaseURL, "/") + "/user/recharge?orderNo=" + orderNo
			u, err := s.alipay.PagePay(orderNo, subject, pkg.AmountFen, ret)
			if err != nil {
				return nil, common.ErrOperation.WithMessage("支付宝电脑支付下单失败")
			}
			payURL = u
			order.CodeURL = &u
		}
	}

	if err := s.store.Create(order); err != nil {
		return nil, common.ErrSystem
	}
	vo := &model.CreatePayOrderVO{
		OrderNo:   order.OrderNo,
		AmountFen: order.AmountFen,
		Points:    order.Points,
		Channel:   order.Channel,
		PayMode:   payMode,
		ExpireAt:  order.ExpireAt.Format(time.RFC3339),
		Status:    order.Status,
		PayURL:    payURL,
	}
	if payMode == model.AlipayModeQRCode && order.CodeURL != nil {
		vo.CodeURL = *order.CodeURL
	}
	return vo, nil
}

func (s *PayService) GetMine(userID int64, orderNo string) (*model.PayOrder, error) {
	if orderNo == "" {
		return nil, common.ErrParams.WithMessage("缺少 orderNo")
	}
	o, err := s.store.GetByOrderNo(orderNo)
	if err != nil || o.UserID != userID {
		return nil, common.ErrNotFound
	}
	// 通知偶发延迟：PENDING 时主动查支付宝
	if o.Status == model.PayPending && o.Channel == model.ChannelAlipay && s.alipay.Enabled() {
		tradeNo, st, err := s.alipay.Query(orderNo)
		if err == nil && (st == "TRADE_SUCCESS" || st == "TRADE_FINISHED") {
			_ = s.store.CreditIfPending(orderNo, tradeNo, o.AmountFen, "")
			o, _ = s.store.GetByOrderNo(orderNo)
		}
	}
	return o, nil
}

func (s *PayService) ListMine(userID, pageNum, pageSize int64) (*model.PageResult, error) {
	if pageNum <= 0 {
		pageNum = common.DefaultPageNum
	}
	if pageSize <= 0 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
	}
	list, total, err := s.store.ListByUser(userID, pageNum, pageSize)
	if err != nil {
		return nil, common.ErrSystem
	}
	vos := make([]model.PayOrderVO, 0, len(list))
	for i := range list {
		vos = append(vos, list[i].ToVO())
	}
	return &model.PageResult{
		Total:    total,
		Records:  vos,
		PageNum:  pageNum,
		PageSize: pageSize,
	}, nil
}

func (s *PayService) HandleAlipayNotify(r *http.Request) error {
	if !s.alipay.Enabled() {
		return common.ErrOperation.WithMessage("支付宝未启用")
	}
	n, err := s.alipay.ParseNotify(r)
	if err != nil {
		return err
	}
	if n.TradeStatus != "TRADE_SUCCESS" && n.TradeStatus != "TRADE_FINISHED" {
		return nil // WAIT_BUYER_PAY 等忽略，仍回 success 避免狂重试
	}
	return s.store.CreditIfPending(n.OutTradeNo, n.TradeNo, 0, n.TotalAmount)
}

func (s *PayService) MockPay(userID int64, orderNo string) error {
	if !s.cfg.Pay.MockEnabled {
		return common.ErrForbidden
	}
	o, err := s.store.GetByOrderNo(orderNo)
	if err != nil || o.UserID != userID {
		return common.ErrNotFound
	}
	return s.store.CreditIfPending(orderNo, "MOCK-"+orderNo, o.AmountFen, "")
}
