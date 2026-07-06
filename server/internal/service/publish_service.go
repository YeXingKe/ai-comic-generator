package service // 业务逻辑层：流水线第 6 步公众号发布

import (
	"context" // 控制微信 API 请求超时
	"fmt"     // 包装上传错误
	"time"    // 记录发布时间

	"github.com/ai-comic-generator/server/internal/client/wechat" // 微信公众号 API 客户端
	"github.com/ai-comic-generator/server/internal/model"         // 流水线状态与发布结果
	"github.com/ai-comic-generator/server/internal/storage"       // 本地合成图路径
)

// PublishService 步骤 6：微信公众号素材上传
type PublishService struct {
	wechat *wechat.MPClient // 微信公众号客户端
	store  *storage.Local   // 本地文件存储（读取合成图路径）
}

// NewPublishService 创建发布服务
func NewPublishService(store *storage.Local, wc *wechat.MPClient) *PublishService {
	return &PublishService{store: store, wechat: wc} // 注入存储与微信客户端
}

// Publish 上传合成图到微信素材库，或标记为草稿
func (s *PublishService) Publish(ctx context.Context, state *model.ComicState) error {
	composed := s.store.ComposedPath(state.TaskID) // 获取合成图本地路径
	title := state.Topic                         // 默认标题用用户输入的主题
	synopsis := ""                               // 故事梗概，暂未用于发布
	if state.StoryIdeation != nil {              // 故事构思步骤已完成
		if state.StoryIdeation.Title != "" { // AI 生成了更好标题
			title = state.StoryIdeation.Title // 优先使用 AI 标题
		}
		synopsis = state.StoryIdeation.Synopsis // 保存梗概供后续扩展图文消息
	}

	result := &model.PublishResult{ // 初始化发布结果
		Platform: "WECHAT_MP", // 发布平台：微信公众号
		Title:    title,       // 发布标题
		Status:   "DRAFT",      // 默认草稿状态
	}

	if s.wechat.Enabled() { // 微信 API 已配置并启用
		mediaID, err := s.wechat.UploadImage(ctx, composed) // 上传合成图为永久素材
		if err != nil { // 上传失败
			result.Status = "FAILED"                      // 标记发布失败
			return fmt.Errorf("wechat upload: %w", err)   // 终止本步骤
		}
		result.MediaID = mediaID     // 记录微信返回的 media_id
		result.Status = "PUBLISHED"  // 标记已发布到素材库
		now := time.Now()            // 取当前时间
		result.PublishedAt = &now    // 记录发布时间
	} else { // 微信未启用
		result.ArticleURL = "" // 文章链接留空
		_ = synopsis           // 梗概暂未使用，避免未使用变量编译警告
	}

	state.PublishResult = result                    // 写入流水线内存态
	state.Phase = model.ComicPhaseWechatPublish       // 更新当前阶段为公众号发布
	return nil                                        // 本步骤完成（草稿也算成功）
}
