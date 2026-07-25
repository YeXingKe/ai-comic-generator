package cos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ai-comic-generator/server/internal/config"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
)

// Client 腾讯云 COS 客户端
type Client struct {
	c         *cossdk.Client
	bucketURL string
	enabled   bool
}

// NewClient 根据配置创建 COS 客户端；未启用时返回 enabled=false 的空客户端
func NewClient(cfg *config.COSConfig) (*Client, error) {
	if !cfg.Enabled || cfg.BucketURL == "" || cfg.SecretID == "" || cfg.SecretKey == "" {
		return &Client{enabled: false}, nil
	}
	u, err := url.Parse(cfg.BucketURL)
	if err != nil {
		return nil, fmt.Errorf("cos bucket_url invalid: %w", err)
	}
	b := &cossdk.BaseURL{BucketURL: u}
	c := cossdk.NewClient(b, &http.Client{
		Transport: &cossdk.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})
	bucketURL := strings.TrimRight(cfg.BucketURL, "/")
	return &Client{c: c, bucketURL: bucketURL, enabled: true}, nil
}

// Enabled 返回 COS 是否可用
func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

// UploadFile 上传本地文件到 COS，返回公网访问 URL
func (c *Client) UploadFile(ctx context.Context, key, localPath string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("cos open file: %w", err)
	}
	defer f.Close()

	key = strings.TrimLeft(key, "/")
	_, err = c.c.Object.Put(ctx, key, f, nil)
	if err != nil {
		return "", fmt.Errorf("cos upload: %w", err)
	}
	return c.bucketURL + "/" + key, nil
}
