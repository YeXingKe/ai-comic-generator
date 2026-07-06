package wechat // 微信公众号 API 客户端包

import (
	"bytes"           // 构建 multipart 请求体缓冲区
	"context"         // 控制 HTTP 请求超时与取消
	"encoding/json"   // 解析微信 API JSON 响应
	"fmt"             // 格式化 URL 与错误信息
	"io"              // 读取 HTTP 响应体、复制文件流
	"mime/multipart"  // 构造上传图片的 multipart/form-data 请求
	"net/http"        // 发起微信官方 HTTP 请求
	"os"              // 打开本地图片文件
	"sync"            // 保护 access_token 并发读写
	"time"            // 记录 token 过期时间、设置 HTTP 超时

	"github.com/ai-comic-generator/server/internal/config" // 读取微信公众号 AppID/AppSecret 配置
)

// MPClient 微信公众号基础 API 客户端
type MPClient struct {
	appID     string       // 公众号 AppID
	appSecret string       // 公众号 AppSecret
	enabled   bool         // 是否已启用（配置完整且 enabled=true）
	http      *http.Client // 复用的 HTTP 客户端（带超时）

	mu          sync.Mutex // 互斥锁，保护 token 缓存的并发安全
	accessToken string     // 缓存的 access_token
	tokenExpire time.Time  // token 过期时间（提前 120 秒刷新）
}

// NewMPClient 根据配置创建微信公众号客户端
func NewMPClient(cfg *config.WeChatConfig) *MPClient {
	return &MPClient{ // 组装客户端实例
		appID:     cfg.AppID,                                              // 写入 AppID
		appSecret: cfg.AppSecret,                                          // 写入 AppSecret
		enabled:   cfg.Enabled && cfg.AppID != "" && cfg.AppSecret != "", // 三项齐全才视为启用
		http:      &http.Client{Timeout: 30 * time.Second},              // 单次请求最长等待 30 秒
	}
}

// Enabled 返回微信公众号 API 是否可用
func (c *MPClient) Enabled() bool {
	return c.enabled // 配置完整且已开启
}

// tokenResp 微信获取 access_token 接口的响应结构
type tokenResp struct {
	AccessToken string `json:"access_token"` // 接口调用凭证
	ExpiresIn   int    `json:"expires_in"`   // 凭证有效时间（秒）
	ErrCode     int    `json:"errcode"`      // 错误码，0 表示成功
	ErrMsg      string `json:"errmsg"`       // 错误描述
}

// uploadResp 微信上传永久素材接口的响应结构
type uploadResp struct {
	MediaID string `json:"media_id"` // 素材库中的 media_id，用于图文消息引用
	ErrCode int    `json:"errcode"`  // 错误码，0 表示成功
	ErrMsg  string `json:"errmsg"`   // 错误描述
}

// getAccessToken 获取（或复用缓存的）access_token
func (c *MPClient) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()         // 加锁，防止并发重复刷新 token
	defer c.mu.Unlock() // 函数退出时释放锁
	if c.accessToken != "" && time.Now().Before(c.tokenExpire) { // 缓存 token 仍有效
		return c.accessToken, nil // 直接返回缓存，避免多余请求
	}
	url := fmt.Sprintf( // 拼接微信获取 token 的 GET 地址
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.appID, c.appSecret, // 填入 AppID 与 AppSecret
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // 创建带上下文的 GET 请求
	if err != nil { // 请求对象创建失败
		return "", err // 返回错误
	}
	resp, err := c.http.Do(req) // 发送 HTTP 请求
	if err != nil { // 网络层错误
		return "", err // 返回错误
	}
	defer resp.Body.Close()              // 确保响应体被关闭
	body, _ := io.ReadAll(resp.Body)     // 读取响应 JSON 正文
	var tr tokenResp                     // 声明 token 响应解析变量
	if err := json.Unmarshal(body, &tr); err != nil { // 解析 JSON
		return "", err // JSON 格式异常
	}
	if tr.ErrCode != 0 { // 微信返回业务错误
		return "", fmt.Errorf("wechat token: %s", tr.ErrMsg) // 包装微信错误信息
	}
	c.accessToken = tr.AccessToken                                              // 缓存新 token
	c.tokenExpire = time.Now().Add(time.Duration(tr.ExpiresIn-120) * time.Second) // 提前 120 秒过期，留出刷新余量
	return c.accessToken, nil                                                   // 返回可用 token
}

// UploadImage 上传永久素材图片，返回 media_id
func (c *MPClient) UploadImage(ctx context.Context, imagePath string) (string, error) {
	if !c.Enabled() { // 微信未启用则拒绝调用
		return "", fmt.Errorf("wechat disabled") // 返回禁用错误
	}
	token, err := c.getAccessToken(ctx) // 先获取 access_token
	if err != nil { // token 获取失败
		return "", err // 无法继续上传
	}

	file, err := os.Open(imagePath) // 打开本地合成图文件
	if err != nil { // 文件不存在或无权读取
		return "", err // 返回文件错误
	}
	defer file.Close() // 确保文件句柄关闭

	var buf bytes.Buffer           // 用于存放 multipart 请求体
	w := multipart.NewWriter(&buf) // 创建 multipart 写入器
	part, err := w.CreateFormFile("media", filepathBase(imagePath)) // 创建名为 media 的文件字段
	if err != nil { // 表单字段创建失败
		return "", err // 返回错误
	}
	if _, err := io.Copy(part, file); err != nil { // 将图片内容复制到表单字段
		return "", err // 复制失败
	}
	_ = w.Close() // 关闭 multipart 写入器，写入结尾 boundary

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=%s&type=image", token) // 拼接上传永久素材 URL
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf) // 创建 POST 请求，body 为 multipart 数据
	if err != nil { // 请求对象创建失败
		return "", err // 返回错误
	}
	req.Header.Set("Content-Type", w.FormDataContentType()) // 设置 Content-Type 含 boundary

	resp, err := c.http.Do(req) // 发送上传请求
	if err != nil { // 网络层错误
		return "", err // 返回错误
	}
	defer resp.Body.Close()          // 关闭响应体
	body, _ := io.ReadAll(resp.Body) // 读取响应 JSON
	var ur uploadResp                // 声明上传响应解析变量
	if err := json.Unmarshal(body, &ur); err != nil { // 解析 JSON
		return "", err // JSON 格式异常
	}
	if ur.ErrCode != 0 { // 微信返回业务错误
		return "", fmt.Errorf("wechat upload: %s", ur.ErrMsg) // 包装微信错误信息
	}
	return ur.MediaID, nil // 返回素材 media_id 供后续图文发布使用
}

// filepathBase 从路径中提取文件名（兼容 Windows 与 Unix 分隔符）
func filepathBase(path string) string {
	for i := len(path) - 1; i >= 0; i-- { // 从路径末尾向前扫描
		if path[i] == '/' || path[i] == '\\' { // 遇到路径分隔符
			return path[i+1:] // 返回分隔符后的文件名部分
		}
	}
	return path // 无分隔符则整个字符串即为文件名
}
