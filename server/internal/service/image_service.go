package service // 业务逻辑层：流水线第 4 步画面生成

import (
	"context"     // 控制混元 API 请求超时
	"fmt"         // 格式化分镜错误信息
	"image/color" // 占位图背景与边框颜色
	"log"         // 记录混元未启用时的降级日志
	"strings"     // 拼接角色外貌参考文本

	"github.com/ai-comic-generator/server/internal/client/hunyuan" // 腾讯混元生图客户端
	"github.com/ai-comic-generator/server/internal/config"         // 全局配置
	"github.com/ai-comic-generator/server/internal/model"          // 流水线状态与分镜图片结果
	"github.com/ai-comic-generator/server/internal/storage"        // 本地分镜图路径与 URL
	"github.com/fogleman/gg"                                       // 绘制占位分镜图
)

// ImageService 步骤 4：混元生图（未启用时生成占位图）
type ImageService struct {
	hunyuan *hunyuan.Client // 混元生图客户端
	store   *storage.Local  // 本地文件存储
	cfg     *config.Config  // 全局配置（预留扩展）
}

// NewImageService 创建画面生成服务
func NewImageService(cfg *config.Config, store *storage.Local, hy *hunyuan.Client) *ImageService {
	return &ImageService{cfg: cfg, store: store, hunyuan: hy} // 注入配置、存储、混元客户端
}

// GeneratePanels 为每个分镜格生成图片并写入 state.PanelImages
func (s *ImageService) GeneratePanels(ctx context.Context, state *model.ComicState) error {
	if state.Storyboard == nil || len(state.Storyboard.Panels) == 0 { // 分镜脚本必须存在
		return fmt.Errorf("storyboard empty") // 无分镜则无法生图
	}
	if err := s.store.EnsureTaskDir(state.TaskID); err != nil { // 确保任务目录存在
		return err // 目录创建失败
	}

	charRef, _ := buildCharacterRef(state.Characters) // 构建角色外貌参考文本（预留一致性生图）
	_ = charRef                                        // 暂未拼入 Prompt，避免未使用变量警告
	results := make([]model.PanelImageResult, 0, len(state.Storyboard.Panels)) // 预分配结果切片

	for _, panel := range state.Storyboard.Panels { // 逐格生成
		dest := s.store.PanelPath(state.TaskID, panel.PanelNo) // 本格图片本地保存路径
		prompt := panel.ImagePrompt                            // 默认使用分镜中的生图 Prompt

		var genErr error // 本格生成错误
		if s.hunyuan.Enabled() { // 混元已启用
			hyPrompt := panel.ImagePrompt // 优先用 imagePrompt
			if hyPrompt == "" {           // imagePrompt 为空
				hyPrompt = panel.Scene // 退而用场景描述
			}
			if state.Style != "" { // 有漫画风格
				hyPrompt = state.Style + " comic style, " + hyPrompt // 在 Prompt 前附加风格
			}
			genErr = s.hunyuan.Generate(ctx, hyPrompt, dest) // 调用混元 API 生图并保存
			prompt = hyPrompt                                // 记录实际使用的 Prompt
		} else { // 混元未启用，降级为占位图
			log.Printf("hunyuan disabled, use placeholder panel: taskId=%s panel=%d", state.TaskID, panel.PanelNo) // 记日志
			genErr = renderPlaceholderPanel(dest, panel.PanelNo, panel.Scene) // 用 gg 绘制占位图
		}
		if genErr != nil { // 本格生成失败
			return fmt.Errorf("panel %d: %w", panel.PanelNo, genErr) // 终止整个步骤
		}

		url := s.store.PublicURL(state.TaskID, fmt.Sprintf("panel_%d.png", panel.PanelNo)) // 生成静态访问 URL
		results = append(results, model.PanelImageResult{ // 追加本格结果
			PanelNo:     panel.PanelNo,                        // 分镜格序号
			URL:         url,                                    // 图片访问地址
			Method:      panelImageMethod(s.hunyuan.Enabled()),  // 来源：AI_GENERATE 或 PLACEHOLDER
			ImagePrompt: prompt,                                 // 实际生图 Prompt
		})
	}

	state.PanelImages = results                      // 写入流水线内存态
	state.Phase = model.ComicPhaseImageGeneration    // 更新当前阶段为画面生成
	return nil                                       // 本步骤完成
}

// panelImageMethod 根据混元是否启用返回图片来源标识
func panelImageMethod(hunyuanOn bool) string {
	if hunyuanOn { // 混元已启用
		return "AI_GENERATE" // 标记为 AI 生成
	}
	return "PLACEHOLDER" // 标记为占位图
}

// buildCharacterRef 将角色列表拼接为外貌参考字符串
func buildCharacterRef(chars []model.ComicCharacter) (string, error) {
	if len(chars) == 0 { // 无角色
		return "", nil // 返回空串
	}
	parts := make([]string, 0, len(chars)) // 预分配片段切片
	for _, c := range chars { // 遍历每个角色
		parts = append(parts, fmt.Sprintf("%s: %s", c.Name, c.Appearance)) // 格式：名字: 外貌
	}
	return strings.Join(parts, "; "), nil // 用分号连接所有角色描述
}

// renderPlaceholderPanel 用 gg 绘制带格号和场景文字的占位分镜图
func renderPlaceholderPanel(path string, panelNo int, scene string) error {
	const w, h = 512, 640                                      // 占位图尺寸（与排版格子比例一致）
	dc := gg.NewContext(w, h)                                  // 创建画布
	dc.SetColor(color.RGBA{240, 240, 250, 255})                // 浅紫灰背景
	dc.Clear()                                                 // 填充背景
	dc.SetColor(color.RGBA{120, 100, 220, 255})                // 紫色边框
	dc.DrawRectangle(20, 20, float64(w-40), float64(h-40))     // 绘制内边框矩形
	dc.SetLineWidth(4)                                         // 边框线宽
	dc.Stroke()                                                // 描边
	dc.SetColor(color.Black)                                   // 文字黑色
	if err := dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 22); err != nil { // 尝试加载 Windows 微软雅黑
		_ = dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 22) // 失败则用 Linux 字体
	}
	dc.DrawStringAnchored(fmt.Sprintf("Panel %d", panelNo), float64(w/2), 60, 0.5, 0.5) // 居中绘制格号
	wrap := wordWrap(scene, 18) // 将场景描述按 18 字换行
	y := 120.0                  // 文字起始纵坐标
	for _, line := range wrap { // 逐行绘制场景描述
		dc.DrawStringAnchored(line, float64(w/2), y, 0.5, 0) // 水平居中
		y += 28 // 行间距
	}
	return dc.SavePNG(path) // 保存为 PNG 文件
}

// wordWrap 按固定字符数将文本拆成多行（支持中文 rune）
func wordWrap(text string, width int) []string {
	runes := []rune(text) // 转为 rune 切片以正确计数字符
	if len(runes) <= width { // 不超过行宽
		return []string{text} // 单行返回
	}
	var lines []string // 多行结果
	for i := 0; i < len(runes); i += width { // 按 width 步进切分
		end := i + width // 本行结束位置
		if end > len(runes) { // 末行不足 width
			end = len(runes) // 取到末尾
		}
		lines = append(lines, string(runes[i:end])) // 追加一行
	}
	return lines // 返回多行切片
}
