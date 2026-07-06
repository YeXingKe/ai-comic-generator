package hunyuan // 腾讯混元生图 API 客户端包

import (
	"context"         // 控制生图请求超时与取消
	"encoding/base64" // 解码 API 返回的 base64 图片数据
	"fmt"             // 格式化错误信息
	"os"              // 写入图片文件到本地磁盘
	"path/filepath"   // 解析目标路径的目录部分

	"github.com/ai-comic-generator/server/internal/config"                              // 读取混元相关配置（密钥、区域等）
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"                     // 腾讯云 SDK 公共类型（凭证、字符串指针）
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"             // 客户端连接配置（Endpoint 等）
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901" // 混元生图 API SDK（v20230901 版本）
)

// Client 腾讯混元生图客户端
type Client struct {
	api     *hunyuan.Client // 腾讯云混元 SDK 底层客户端实例
	enabled bool            // 是否已启用（配置了有效密钥时为 true）
}

// NewClient 根据配置创建混元客户端；未启用时返回 enabled=false 的空客户端
func NewClient(cfg *config.HunyuanConfig) (*Client, error) {
	if !cfg.Enabled || cfg.SecretID == "" || cfg.SecretKey == "" { // 未开启或密钥未填写
		return &Client{enabled: false}, nil // 返回禁用状态的客户端，不报错
	}
	credential := common.NewCredential(cfg.SecretID, cfg.SecretKey) // 构造腾讯云 API 鉴权凭证
	cpf := profile.NewClientProfile()                                // 创建客户端配置对象
	cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"         // 指定混元 API 域名
	region := cfg.Region                                             // 读取配置中的地域
	if region == "" {                                                // 未配置地域时使用默认值
		region = "ap-guangzhou" // 默认广州地域
	}
	client, err := hunyuan.NewClient(credential, region, cpf) // 初始化混元 SDK 客户端
	if err != nil {                                           // SDK 初始化失败
		return nil, fmt.Errorf("hunyuan client: %w", err) // 包装错误并返回
	}
	return &Client{api: client, enabled: true}, nil // 返回已启用的客户端实例
}

// Enabled 返回混元生图是否可用
func (c *Client) Enabled() bool {
	return c != nil && c.enabled // 客户端非空且已配置有效密钥
}

// Generate 根据 Prompt 生图并保存到 destPath
func (c *Client) Generate(ctx context.Context, prompt, destPath string) error {
	if !c.Enabled() { // 混元未启用则拒绝调用
		return fmt.Errorf("hunyuan disabled") // 返回禁用错误
	}
	req := hunyuan.NewTextToImageLiteRequest()       // 创建文生图精简版请求对象
	req.Prompt = common.StringPtr(prompt)            // 设置生图 Prompt（转为 SDK 字符串指针）
	req.Resolution = common.StringPtr("768:1024")    // 设置输出分辨率（竖版漫画比例）
	req.RspImgType = common.StringPtr("base64")      // 要求 API 以 base64 编码返回图片

	resp, err := c.api.TextToImageLiteWithContext(ctx, req) // 带上下文调用混元文生图 API
	if err != nil {                                         // 网络或 API 层错误
		return fmt.Errorf("hunyuan text to image: %w", err) // 包装 API 错误
	}
	if resp.Response == nil || resp.Response.ResultImage == nil { // 响应体或图片字段为空
		return fmt.Errorf("hunyuan empty image response") // 视为无效响应
	}

	data, err := base64.StdEncoding.DecodeString(*resp.Response.ResultImage) // 将 base64 图片解码为字节数组
	if err != nil { // base64 解码失败
		return fmt.Errorf("decode image: %w", err) // 包装解码错误
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil { // 确保目标目录存在
		return err // 目录创建失败则直接返回
	}
	return os.WriteFile(destPath, data, 0o644) // 将图片字节写入本地文件
}
