package store

import (
	"fmt"
	"time"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PayStore struct{ db *gorm.DB }

func NewPayStore(db *gorm.DB) *PayStore { return &PayStore{db: db} }

func (s *PayStore) Create(o *model.PayOrder) error { return s.db.Create(o).Error }

func (s *PayStore) GetByOrderNo(orderNo string) (*model.PayOrder, error) {
	var o model.PayOrder
	err := s.db.Where("orderNo = ?", orderNo).First(&o).Error
	return &o, err
}

func (s *PayStore) ListByUser(userID, page, size int64) ([]model.PayOrder, int64, error) {
	q := s.db.Model(&model.PayOrder{}).Where("userId = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PayOrder
	err := q.Order("id DESC").Offset(int((page-1)*size)).Limit(int(size)).Find(&list).Error
	return list, total, err
}

func (s *PayStore) CreditIfPending(orderNo, tradeNo string, expectFen int, paidYuan string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var order model.PayOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("orderNo = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == model.PayPaid {
			return nil // 支付宝会重试，直接当成功
		}
		if order.Status != model.PayPending {
			return common.ErrOperation.WithMessage("订单不可支付")
		}
		// 金额校验：通知里的 total_amount 必须等于下单金额
		want := fenToYuan(order.AmountFen)
		if expectFen > 0 && order.AmountFen != expectFen {
			return common.ErrOperation.WithMessage("金额不一致")
		}
		if paidYuan != "" && paidYuan != want {
			return common.ErrOperation.WithMessage("金额不一致")
		}

		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]any{
			"status":       model.PayPaid,
			"channelTxnId": tradeNo,
			"paidAt":       now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			"UPDATE user SET points = points + ? WHERE id = ?",
			order.Points, order.UserID,
		).Error; err != nil {
			return err
		}
		var points int
		if err := tx.Model(&model.User{}).Select("points").
			Where("id = ?", order.UserID).Scan(&points).Error; err != nil {
			return err
		}
		return tx.Create(&model.PointLog{
			UserID:       order.UserID,
			Delta:        order.Points,
			BalanceAfter: points,
			BizType:      model.BizRecharge,
			BizID:        order.OrderNo,
			Remark:       "支付宝充值",
		}).Error
	})
}

func fenToYuan(fen int) string {
	return fmt.Sprintf("%d.%02d", fen/100, fen%100)
}