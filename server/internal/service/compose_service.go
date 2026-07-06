package service // 业务逻辑层：流水线第 5 步排版合成

import (
	"context" // 接口签名需要，当前实现未使用
	"fmt"     // 格式化加载/合成错误
	"image"   // 图片解码接口
	"image/color" // 画布背景、气泡、文字颜色
	_ "image/png" // 注册 PNG 解码器
	"os"      // 打开本地分镜图文件

	"github.com/ai-comic-generator/server/internal/model"   // 流水线状态与排版结果
	"github.com/ai-comic-generator/server/internal/storage" // 分镜图/合成图路径与 URL
	"github.com/disintegration/imaging"                     // 缩放、裁剪分镜图
	"github.com/fogleman/gg"                                // 绘制宫格、标题、对话气泡
)

// ComposeService 步骤 5：imaging 拼接 + gg 绘制气泡与文字
type ComposeService struct {
	store *storage.Local // 本地文件存储
}

// NewComposeService 创建排版合成服务
func NewComposeService(store *storage.Local) *ComposeService {
	return &ComposeService{store: store} // 注入本地存储
}

// Compose 将各分镜图拼成宫格长图，叠加标题与气泡，写入 state.ComposedLayout
func (s *ComposeService) Compose(ctx context.Context, state *model.ComicState) error {
	_ = ctx // 当前实现不依赖 ctx，保留接口一致性
	if state.Storyboard == nil || len(state.PanelImages) == 0 { // 分镜与图片必须就绪
		return fmt.Errorf("panels not ready") // 前置步骤未完成
	}

	const cellW, cellH = 512, 640        // 单格宽高（与分镜图/占位图一致）
	cols := 2                            // 宫格列数：2 列
	rows := (len(state.Storyboard.Panels) + cols - 1) / cols // 行数：向上取整
	canvasW := cols * cellW              // 画布总宽度
	canvasH := rows*cellH + 80           // 画布总高度（顶部留 80px 给标题）

	dc := gg.NewContext(canvasW, canvasH) // 创建 gg 画布
	dc.SetColor(color.White)              // 白色背景
	dc.Clear()                            // 填充背景

	if state.StoryIdeation != nil && state.StoryIdeation.Title != "" { // 有故事标题
		_ = dc.LoadFontFace("C:/Windows/Fonts/msyhbd.ttc", 28) // 加载粗体微软雅黑
		dc.SetColor(color.Black)                               // 标题黑色
		dc.DrawStringAnchored(state.StoryIdeation.Title, float64(canvasW)/2, 36, 0.5, 0.5) // 顶部居中绘标题
	}

	for i, panel := range state.Storyboard.Panels { // 逐格拼接到画布
		imgPath := s.store.PanelPath(state.TaskID, panel.PanelNo) // 本格本地图片路径
		src, err := loadImage(imgPath) // 从磁盘加载图片
		if err != nil { // 文件不存在或解码失败
			return fmt.Errorf("load panel %d: %w", panel.PanelNo, err) // 终止排版
		}
		resized := imaging.Fit(src, cellW-16, cellH-16, imaging.Lanczos) // 等比缩放到格内（留 16px 边距）
		col := i % cols // 列索引（0 或 1）
		row := i / cols // 行索引
		x := col*cellW + 8  // 本格左上角 X（列偏移 + 左边距）
		y := row*cellH + 72 // 本格左上角 Y（行偏移 + 标题区高度）
		dc.DrawImage(resized, x, y) // 将缩放后的分镜图贴到画布
		drawBubble(dc, panel, float64(x)+20, float64(y)+float64(resized.Bounds().Dy())-80) // 在格底部绘制对话气泡
	}

	out := s.store.ComposedPath(state.TaskID) // 合成图保存路径
	if err := dc.SavePNG(out); err != nil { // 导出 PNG
		return err // 保存失败
	}

	previewURL := s.store.PublicURL(state.TaskID, "composed.png") // 合成图静态访问 URL
	state.ComposedLayout = &model.ComposedLayoutResult{ // 写入排版结果
		Format:     "grid",              // 排版格式：宫格
		PreviewURL: previewURL,          // 预览图 URL
		AssetURLs:  collectPanelURLs(state.PanelImages), // 各分镜图 URL 列表
		CoverImage: previewURL,          // 封面用合成图
	}
	state.Phase = model.ComicPhaseLayoutCompose // 更新当前阶段为排版合成
	return nil                                  // 本步骤完成
}

// loadImage 从本地路径加载并解码图片
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path) // 打开文件
	if err != nil { // 文件不存在或无权读取
		return nil, err // 返回错误
	}
	defer f.Close()           // 确保关闭文件句柄
	img, _, err := image.Decode(f) // 按格式解码（PNG 等）
	return img, err // 返回 image.Image
}

// drawBubble 在指定位置绘制圆角矩形对话气泡及文字
func drawBubble(dc *gg.Context, panel model.StoryboardPanel, x, y float64) {
	text := "" // 气泡文字
	if len(panel.Dialogue) > 0 { // 优先使用第一句台词
		text = panel.Dialogue[0] // 取第一条对话
	} else if panel.Narration != "" { // 无台词则用旁白
		text = panel.Narration // 旁白文字
	}
	if text == "" { // 无文字可显示
		return // 跳过绘制
	}
	_ = dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 18) // 加载微软雅黑 18 号
	w, h := dc.MeasureString(text) // 测量文字宽高
	pad := 12.0                    // 气泡内边距
	bw, bh := w+pad*2, h+pad*2     // 气泡总宽高
	dc.SetColor(color.RGBA{255, 255, 255, 230}) // 半透明白色填充
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)     // 绘制圆角矩形
	dc.Fill()                                   // 填充气泡
	dc.SetColor(color.Black)                    // 黑色边框
	dc.SetLineWidth(2)                          // 边框线宽
	dc.DrawRoundedRectangle(x, y, bw, bh, 8)    // 再绘一遍矩形用于描边
	dc.Stroke()                                 // 描边
	dc.DrawStringAnchored(text, x+bw/2, y+bh/2, 0.5, 0.5) // 文字居中于气泡
}

// collectPanelURLs 从分镜图片结果中提取 URL 列表
func collectPanelURLs(images []model.PanelImageResult) []string {
	urls := make([]string, 0, len(images)) // 预分配切片
	for _, img := range images { // 遍历每格
		urls = append(urls, img.URL) // 追加 URL
	}
	return urls // 返回 URL 列表
}
