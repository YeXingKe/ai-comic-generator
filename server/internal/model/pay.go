package model



import "time"



const (
	PayPending    = "PENDING"  // 待支付
	PayPaid       = "PAID"     // 已支付
	PayClosed     = "CLOSED"   // 已关闭（超时或取消）
	ChannelAlipay = "alipay"   // 支付宝
	ChannelMock   = "mock"     // 本地模拟支付（仅开发）
	BizRecharge   = "recharge" // 积分流水业务类型：充值入账

	// 支付宝支付形态（由 config pay.alipay.mode 决定，前端不传）
	AlipayModeQRCode = "qrcode" // 当面付扫码
	AlipayModePage   = "page"   // 电脑网站支付
)

// PayOrder 积分充值订单（表 pay_order）

type PayOrder struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`                       // 主键 ID
	OrderNo      string     `gorm:"column:orderNo;uniqueIndex" json:"orderNo"`                // 商户订单号（out_trade_no）
	UserID       int64      `gorm:"column:userId" json:"userId"`                              // 下单用户 ID
	Channel      string     `gorm:"column:channel" json:"channel"`                            // 支付渠道：alipay / mock
	PackageCode  string     `gorm:"column:packageCode" json:"packageCode"`                    // 充值方案编码
	AmountFen    int        `gorm:"column:amountFen" json:"amountFen"`                        // 应付金额（分）
	Points       int        `gorm:"column:points" json:"points"`                              // 到账积分
	Status       string     `gorm:"column:status" json:"status"`                              // 订单状态：PENDING / PAID / CLOSED
	ChannelTxnID *string    `gorm:"column:channelTxnId" json:"channelTxnId"`                  // 渠道交易号（如支付宝 trade_no）
	CodeURL      *string    `gorm:"column:codeUrl" json:"codeUrl"`                            // 扫码支付链接（支付宝 qr_code）
	ExpireAt     time.Time  `gorm:"column:expireAt" json:"expireAt"`                          // 订单过期时间
	PaidAt       *time.Time `gorm:"column:paidAt" json:"paidAt"`                              // 支付成功时间
	CreateTime   time.Time  `gorm:"column:createTime;autoCreateTime" json:"createTime"`       // 创建时间
	UpdateTime   time.Time  `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`       // 更新时间
}



func (PayOrder) TableName() string { return "pay_order" }



// PointLog 积分变动流水（表 point_log）

type PointLog struct {
	ID           int64     `gorm:"primaryKey" json:"id"`                                 // 主键 ID
	UserID       int64     `gorm:"column:userId" json:"userId"`                          // 用户 ID
	Delta        int       `gorm:"column:delta" json:"delta"`                            // 积分变动量（正增负减）
	BalanceAfter int       `gorm:"column:balanceAfter" json:"balanceAfter"`                // 变动后余额
	BizType      string    `gorm:"column:bizType" json:"bizType"`                          // 业务类型（如 recharge）
	BizID        string    `gorm:"column:bizId" json:"bizId"`                              // 业务关联 ID（如订单号）
	Remark       string    `gorm:"column:remark" json:"remark"`                            // 备注
	CreateTime   time.Time `gorm:"column:createTime;autoCreateTime" json:"createTime"`    // 创建时间
}



func (PointLog) TableName() string { return "point_log" }



// PayPackagePlan 充值方案（库表 pay_package）

type PayPackagePlan struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`                   // 主键 ID
	Code       string    `gorm:"column:code;uniqueIndex" json:"code"`                  // 套餐编码（唯一）
	Name       string    `gorm:"column:name" json:"name"`                              // 展示名称
	AmountFen  int       `gorm:"column:amountFen" json:"amountFen"`                    // 售价（分）
	Points     int       `gorm:"column:points" json:"points"`                          // 到账积分
	SortOrder  int       `gorm:"column:sortOrder;default:0" json:"sortOrder"`          // 排序，越小越靠前
	Enabled    int       `gorm:"column:enabled;default:1" json:"enabled"`                // 上架状态：1 上架，0 下架
	CreateTime time.Time `gorm:"column:createTime;autoCreateTime" json:"createTime"`   // 创建时间
	UpdateTime time.Time `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`   // 更新时间
}



func (PayPackagePlan) TableName() string { return "pay_package" }



// PayPackage 用户端展示的充值套餐项

type PayPackage struct {
	Code      string `json:"code"`      // 套餐编码
	Name      string `json:"name"`      // 展示名称
	AmountFen int    `json:"amountFen"` // 售价（分）
	Points    int    `json:"points"`    // 到账积分
}



func (p *PayPackagePlan) ToPayPackage() PayPackage {
	if p == nil {
		return PayPackage{}

	}
	return PayPackage{Code: p.Code, Name: p.Name, AmountFen: p.AmountFen, Points: p.Points}
}



// CreatePayOrderRequest 创建充值订单

type CreatePayOrderRequest struct {
	PackageCode string `json:"packageCode" binding:"required"` // 充值方案编码
	Channel     string `json:"channel" binding:"required"`     // 支付渠道：alipay | mock
}



// CreatePayOrderVO 创建订单响应

type CreatePayOrderVO struct {
	OrderNo   string `json:"orderNo"`   // 商户订单号
	AmountFen int    `json:"amountFen"` // 应付金额（分）
	Points    int    `json:"points"`    // 到账积分
	Channel   string `json:"channel"`   // 支付渠道
	PayMode   string `json:"payMode"`   // 实际支付形态：qrcode | page（来自服务端配置）
	CodeURL   string `json:"codeUrl"`   // 扫码二维码内容（qrcode 模式）
	PayURL    string `json:"payUrl"`    // PC 跳转收银台（page 模式）
	ExpireAt  string `json:"expireAt"`  // 过期时间（RFC3339）
	Status    string `json:"status"`    // 订单状态
}



// PayOrderVO 订单摘要（用户端）

type PayOrderVO struct {
	OrderNo    string  `json:"orderNo"`    // 商户订单号
	AmountFen  int     `json:"amountFen"`  // 应付金额（分）
	Points     int     `json:"points"`     // 到账积分
	Channel    string  `json:"channel"`    // 支付渠道
	Status     string  `json:"status"`     // 订单状态
	PaidAt     *string `json:"paidAt"`     // 支付成功时间（RFC3339，未支付为空）
	CreateTime string  `json:"createTime"` // 创建时间（RFC3339）
}



// PayCatalogVO 充值页套餐与支付能力

type PayCatalogVO struct {
	MockEnabled   bool         `json:"mockEnabled"`   // 是否开启模拟支付
	AlipayEnabled bool         `json:"alipayEnabled"` // 支付宝是否可用
	AlipaySandbox bool         `json:"alipaySandbox"` // 是否为沙箱网关（前端联调提示）
	Packages      []PayPackage `json:"packages"`      // 上架中的充值方案列表
}



// PayOrderPageRequest 用户充值订单分页

type PayOrderPageRequest struct {
	PageNum  int64 `json:"pageNum" binding:"required,gt=0"`           // 页码
	PageSize int64 `json:"pageSize" binding:"required,gt=0,lte=100"` // 每页条数

}



// AdminPayOrderPageRequest 管理端订单/收款分页查询

type AdminPayOrderPageRequest struct {
	PageNum     int64   `json:"pageNum" binding:"required,gt=0"`           // 页码
	PageSize    int64   `json:"pageSize" binding:"required,gt=0,lte=100"` // 每页条数
	Status      *string `json:"status"`                                    // 订单状态筛选
	UserID      *int64  `json:"userId"`                                    // 用户 ID 筛选
	OrderNo     *string `json:"orderNo"`                                   // 订单号模糊筛选
	UserAccount *string `json:"userAccount"`                               // 登录账号模糊筛选
}



// PayPackagePlanPageRequest 充值方案分页

type PayPackagePlanPageRequest struct {
	PageNum  int64 `json:"pageNum" binding:"required,gt=0"`           // 页码
	PageSize int64 `json:"pageSize" binding:"required,gt=0,lte=100"` // 每页条数
}



// AddPayPackagePlanRequest 新增充值方案

type AddPayPackagePlanRequest struct {
	Code      string `json:"code" binding:"required,min=2,max=32"` // 套餐编码
	Name      string `json:"name" binding:"required,min=1,max=64"` // 展示名称
	AmountFen int    `json:"amountFen" binding:"required,gt=0"`    // 售价（分）
	Points    int    `json:"points" binding:"required,gt=0"`       // 到账积分
	SortOrder *int   `json:"sortOrder"`                            // 排序（可选）
	Enabled   *int   `json:"enabled"`                              // 上架状态（可选，默认 1）
}



// UpdatePayPackagePlanRequest 更新充值方案

type UpdatePayPackagePlanRequest struct {
	ID        int64  `json:"id" binding:"required,gt=0"`           // 方案主键 ID
	Name      string `json:"name" binding:"required,min=1,max=64"` // 展示名称
	AmountFen int    `json:"amountFen" binding:"required,gt=0"`  // 售价（分）
	Points    int    `json:"points" binding:"required,gt=0"`     // 到账积分
	SortOrder *int   `json:"sortOrder"`                          // 排序（可选）
	Enabled   *int   `json:"enabled"`                            // 上架状态（可选）
}



// DeletePayPackagePlanRequest 删除充值方案

type DeletePayPackagePlanRequest struct {
	ID int64 `json:"id" binding:"required,gt=0"` // 方案主键 ID
}



// PayOrderAdminVO 管理端订单详情（含用户信息）

type PayOrderAdminVO struct {
	PayOrderVO
	UserID       int64   `json:"userId"`       // 用户 ID
	UserAccount  string  `json:"userAccount"`  // 登录账号
	PackageCode  string  `json:"packageCode"`  // 充值方案编码
	ChannelTxnID *string `json:"channelTxnId"` // 渠道交易号
}



func (o *PayOrder) ToVO() PayOrderVO {

	vo := PayOrderVO{
		OrderNo:    o.OrderNo,
		AmountFen:  o.AmountFen,
		Points:     o.Points,
		Channel:    o.Channel,
		Status:     o.Status,
		CreateTime: o.CreateTime.Format(time.RFC3339),
	}

	if o.PaidAt != nil {
		s := o.PaidAt.Format(time.RFC3339)
		vo.PaidAt = &s
	}

	return vo

}


