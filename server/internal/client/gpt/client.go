package gpt // OpenAI 图片生成客户端包（支持中转站 API）

import (
    "bytes"          // 用于将 JSON 字节转换为 io.Reader
    "context"        // 控制请求超时与取消
    "encoding/json"  // JSON 序列化与反序列化
    "fmt"            // 格式化错误信息
    "io"             // 读取 HTTP 响应体与文件拷贝
    "net/http"       // HTTP 客户端请求
    "os"             // 文件创建与目录操作
    "path/filepath"  // 路径解析与目录提取
    "time"           // 设置 HTTP 超时时间

    "github.com/ai-comic-generator/server/internal/config" // 读取配置（API Key、BaseURL 等）
)

// Client OpenAI 图片生成客户端（支持中转站）
type Client struct {
    apiKey  string       // OpenAI API 密钥（或中转站密钥）
    baseURL string       // API 基础 URL（如 https://api.wszhu.top/v1）
    model   string       // 模型名称（gpt-image-2 或 gpt-image-2-4k）
    client  *http.Client // HTTP 客户端实例（配置超时时间）
    enabled bool         // 是否已启用（密钥未配置时为 false）
}

// imageRequest OpenAI 图片生成请求体
type imageRequest struct {
    Model          string `json:"model"`           // 模型名称（gpt-image-2）
    Prompt         string `json:"prompt"`          // 生图提示词
    N              int    `json:"n"`               // 生成图片数量（固定为 1）
    Size           string `json:"size"`            // 图片尺寸（1024x1024 等）
    Quality        string `json:"quality,omitempty"` // 图片质量（standard 或 hd，可选）
    ResponseFormat string `json:"response_format"` // 响应格式（url 或 b64_json）
}

// imageResponse OpenAI 图片生成响应
type imageResponse struct {
    Created int64 `json:"created"` // 生成时间戳（Unix 时间）
    Data    []struct {            // 生成的图片列表（通常只有一张）
        URL           string `json:"url,omitempty"`            // 图片下载 URL（临时链接）
        B64JSON       string `json:"b64_json,omitempty"`       // Base64 编码的图片数据（可选）
        RevisedPrompt string `json:"revised_prompt,omitempty"` // AI 修订后的提示词（可选）
    } `json:"data"`
}

// errorResponse OpenAI 错误响应
type errorResponse struct {
    Error struct {                 // 错误详情对象
        Message string `json:"message"` // 错误描述信息
        Type    string `json:"type"`    // 错误类型（如 invalid_request_error）
        Code    string `json:"code"`    // 错误代码（如 invalid_api_key）
    } `json:"error"`
}

// NewClient 创建 OpenAI 图片客户端（支持中转站）
func NewClient(cfg *config.DalleConfig) (*Client, error) {
    if !cfg.Enabled || cfg.APIKey == "" { // 检查是否启用且 API Key 已配置
        return &Client{enabled: false}, nil // 返回禁用状态的客户端（不报错）
    }

    baseURL := cfg.BaseURL // 读取配置中的 API 基础 URL
    if baseURL == "" {     // 如果未配置，使用 OpenAI 官方 API
        baseURL = "https://api.openai.com/v1" // 官方 API 地址（国内可能需要代理）
    }

    model := cfg.Model // 读取配置中的模型名称
    if model == "" {   // 如果未配置，使用默认模型
        model = "dall-e-3" // 默认兼容模型（也可用 gpt-image-2）
    }

    timeout := time.Duration(cfg.Timeout) * time.Second // 将配置的秒数转换为 Duration
    if timeout == 0 {                                   // 如果未配置超时时间
        timeout = 120 * time.Second // 默认 120 秒（图片生成通常需要较长时间）
    }

    return &Client{ // 返回已配置的客户端实例
        apiKey:  cfg.APIKey,  // 设置 API 密钥
        baseURL: baseURL,     // 设置 API 基础 URL
        model:   model,       // 设置模型名称
        client: &http.Client{ // 创建 HTTP 客户端
            Timeout: timeout, // 配置请求超时时间
        },
        enabled: true, // 标记为已启用
    }, nil
}

// Enabled 返回是否已启用
func (c *Client) Enabled() bool {
    return c != nil && c.enabled // 检查客户端实例非空且已启用标志为 true
}

// Name 返回生成器名称
func (c *Client) Name() string {
    return "openai" // 返回固定名称用于日志标识和监控
}

// Generate 根据 prompt 生图并保存到 destPath
func (c *Client) Generate(ctx context.Context, prompt, destPath string) error {
    if !c.Enabled() { // 检查客户端是否已启用
        return fmt.Errorf("openai image generator disabled") // 返回禁用错误
    }

    // 构造请求体（适配中转站）
    reqBody := imageRequest{                 // 创建图片生成请求结构体
        Model:          c.model,             // 使用配置的模型名称（gpt-image-2 等）
        Prompt:         prompt,              // 用户提供的生图提示词
        N:              1,                   // 生成图片数量固定为 1 张
        Size:           "1024x1024",         // 图片尺寸（中转站通常支持标准尺寸）
        Quality:        "standard",          // 图片质量（standard 或 hd，部分中转站可能不支持 hd）
        ResponseFormat: "url",               // 响应格式为 URL（也可选 b64_json）
    }

    bodyJSON, err := json.Marshal(reqBody) // 将请求结构体序列化为 JSON 字节
    if err != nil {                        // 序列化失败（理论上不会发生）
        return fmt.Errorf("marshal request: %w", err) // 包装错误并返回
    }

    // 创建 HTTP 请求
    req, err := http.NewRequestWithContext(ctx, "POST", // 创建带上下文的 POST 请求
        c.baseURL+"/images/generations",                // API 端点（修复：补全 /images/generations）
        bytes.NewBuffer(bodyJSON))                      // 将 JSON 字节作为请求体
    if err != nil {                                     // 请求创建失败
        return fmt.Errorf("create request: %w", err)    // 包装错误并返回
    }

    req.Header.Set("Authorization", "Bearer "+c.apiKey) // 设置 Authorization 头（Bearer Token 认证）
    req.Header.Set("Content-Type", "application/json")  // 设置 Content-Type 为 JSON

    // 发送请求
    resp, err := c.client.Do(req) // 发送 HTTP 请求到 OpenAI API
    if err != nil {               // 网络错误或超时
        return fmt.Errorf("http request: %w", err) // 包装错误并返回
    }
    defer resp.Body.Close() // 确保响应体在函数结束时关闭

    // 读取响应
    bodyBytes, err := io.ReadAll(resp.Body) // 读取完整的响应体字节
    if err != nil {                         // 读取失败
        return fmt.Errorf("read response: %w", err) // 包装错误并返回
    }

    // 检查错误响应
    if resp.StatusCode != http.StatusOK { // 如果 HTTP 状态码不是 200
        var errResp errorResponse                     // 定义错误响应结构体
        if err := json.Unmarshal(bodyBytes, &errResp); err == nil { // 尝试解析为错误响应格式
            return fmt.Errorf("openai api error [%d]: %s (type: %s, code: %s)", // 返回格式化的 API 错误信息
                resp.StatusCode,         // HTTP 状态码
                errResp.Error.Message,   // 错误消息
                errResp.Error.Type,      // 错误类型
                errResp.Error.Code)      // 错误代码
        }
        return fmt.Errorf("http status %d: %s", resp.StatusCode, string(bodyBytes)) // 解析失败则返回原始响应
    }

    // 解析成功响应
    var imgResp imageResponse                     // 定义图片响应结构体
    if err := json.Unmarshal(bodyBytes, &imgResp); err != nil { // 反序列化 JSON 响应
        return fmt.Errorf("unmarshal response: %w", err) // 解析失败返回错误
    }

    if len(imgResp.Data) == 0 || imgResp.Data[0].URL == "" { // 检查响应中是否有图片 URL
        return fmt.Errorf("empty image response") // 没有图片数据则返回错误
    }

    // 下载图片
    imageURL := imgResp.Data[0].URL                 // 提取第一张图片的 URL
    return c.downloadImage(ctx, imageURL, destPath) // 调用下载方法保存到本地
}

// downloadImage 下载图片到本地
func (c *Client) downloadImage(ctx context.Context, url, destPath string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil) // 创建带上下文的 GET 请求下载图片
    if err != nil {                                              // 请求创建失败
        return fmt.Errorf("create download request: %w", err)    // 包装错误并返回
    }

    resp, err := c.client.Do(req) // 发送 HTTP GET 请求
    if err != nil {                // 网络错误或超时
        return fmt.Errorf("download image: %w", err) // 包装错误并返回
    }
    defer resp.Body.Close() // 确保响应体在函数结束时关闭

    if resp.StatusCode != http.StatusOK { // 检查下载是否成功（HTTP 200）
        return fmt.Errorf("download failed: status %d", resp.StatusCode) // 下载失败返回状态码
    }

    // 确保目录存在
    if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil { // 创建目标文件所在的目录（递归创建）
        return fmt.Errorf("create directory: %w", err) // 目录创建失败返回错误
    }

    // 保存文件
    file, err := os.Create(destPath) // 创建目标文件（如果已存在则覆盖）
    if err != nil {                  // 文件创建失败
        return fmt.Errorf("create file: %w", err) // 包装错误并返回
    }
    defer file.Close() // 确保文件在函数结束时关闭

    if _, err := io.Copy(file, resp.Body); err != nil { // 将响应体内容拷贝到文件（流式写入）
        return fmt.Errorf("write file: %w", err) // 写入失败返回错误
    }

    return nil // 下载成功，返回 nil
}