package alipayx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ai-comic-generator/server/internal/common"
	"github.com/smartwalle/alipay/v3"
)

type Config struct {
	Enabled         bool   `mapstructure:"enabled"`
	AppID           string `mapstructure:"app_id"`
	PrivateKey      string `mapstructure:"private_key"`
	AlipayPublicKey string `mapstructure:"alipay_public_key"`
	Sandbox         bool   `mapstructure:"sandbox"`
}

type Client struct {
	api       *alipay.Client
	notifyURL string
	returnURL string // 新增：电脑网站支付默认回跳（可再拼 orderNo）
}

// Enabled 是否已初始化可用客户端
func (c *Client) Enabled() bool {
	return c != nil && c.api != nil
}

func New(cfg *Config, notifyURL, returnURL string) (*Client, error) {
	if cfg == nil || !cfg.Enabled {
		return &Client{}, nil
	}
	isProd := !cfg.Sandbox
	api, err := alipay.New(cfg.AppID, cfg.PrivateKey, isProd)
	if err != nil {
		return nil, fmt.Errorf("alipay new: %w", err)
	}
	if err := api.LoadAliPayPublicKey(cfg.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("alipay public key: %w", err)
	}
	return &Client{api: api, notifyURL: notifyURL, returnURL: returnURL}, nil
}

func fenToYuan(fen int) string {
	return fmt.Sprintf("%d.%02d", fen/100, fen%100)
}

// Precreate 当面付扫码，返回 qr_code
func (c *Client) Precreate(outTradeNo, subject string, amountFen int) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("alipay 未启用")
	}
	p := alipay.TradePreCreate{
		Trade: alipay.Trade{
			NotifyURL:      c.notifyURL,
			Subject:        subject,
			OutTradeNo:     outTradeNo,
			TotalAmount:    fenToYuan(amountFen),
			ProductCode:    "FACE_TO_FACE_PAYMENT",
			TimeoutExpress: "5m",
		},
	}
	rsp, err := c.api.TradePreCreate(context.Background(), p)
	if err != nil {
		return "", err
	}
	if rsp.Code != alipay.CodeSuccess {
		return "", fmt.Errorf("alipay precreate %s %s", rsp.Code, rsp.Msg)
	}
	return rsp.QRCode, nil
}

// PagePay 电脑网站支付：返回跳转支付宝收银台的 URL
func (c *Client) PagePay(outTradeNo, subject string, amountFen int, returnURL string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("alipay 未启用")
	}
	returnURL = common.FirstNonBlank(returnURL, c.returnURL)
	if common.IsBlank(returnURL) {
		return "", fmt.Errorf("缺少 return_url")
	}
	p := alipay.TradePagePay{}
	p.NotifyURL = c.notifyURL
	p.ReturnURL = returnURL
	p.Subject = subject
	p.OutTradeNo = outTradeNo
	p.TotalAmount = fenToYuan(amountFen)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.TimeoutExpress = "5m"
	u, err := c.api.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

type Notify struct {
	OutTradeNo  string
	TradeNo     string
	TradeStatus string
	TotalAmount string
}

func (c *Client) ParseNotify(r *http.Request) (*Notify, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("alipay 未启用")
	}
	n, err := c.api.GetTradeNotification(r)
	if err != nil {
		return nil, err
	}
	return &Notify{
		OutTradeNo:  n.OutTradeNo,
		TradeNo:     n.TradeNo,
		TradeStatus: string(n.TradeStatus),
		TotalAmount: n.TotalAmount,
	}, nil
}

// Query 通知延迟时，前端轮询可走服务端查单补入账
func (c *Client) Query(outTradeNo string) (tradeNo, status string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("alipay 未启用")
	}
	rsp, err := c.api.TradeQuery(context.Background(), alipay.TradeQuery{OutTradeNo: outTradeNo})
	if err != nil {
		return "", "", err
	}
	if rsp.Code != alipay.CodeSuccess {
		return "", "", fmt.Errorf("alipay query %s %s", rsp.Code, rsp.Msg)
	}
	return rsp.TradeNo, string(rsp.TradeStatus), nil
}