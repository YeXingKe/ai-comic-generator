package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-comic-generator/server/internal/config"
)

// Local 本地漫画资源存储
type Local struct {
	basePath  string
	publicURL string
}

func NewLocal(cfg *config.StorageConfig) (*Local, error) {
	if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
		return nil, err
	}
	return &Local{basePath: cfg.BasePath, publicURL: cfg.PublicURL}, nil
}

func (l *Local) TaskDir(taskID string) string {
	return filepath.Join(l.basePath, taskID)
}

func (l *Local) PanelPath(taskID string, panelNo int) string {
	return filepath.Join(l.TaskDir(taskID), fmt.Sprintf("panel_%d.png", panelNo))
}

func (l *Local) ComposedPath(taskID string) string {
	return filepath.Join(l.TaskDir(taskID), "composed.png")
}

func (l *Local) PublicURL(taskID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", trimSlash(l.publicURL), taskID, filename)
}

func trimSlash(s string) string {
	if len(s) > 1 && s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}

func (l *Local) EnsureTaskDir(taskID string) error {
	return os.MkdirAll(l.TaskDir(taskID), 0o755)
}
