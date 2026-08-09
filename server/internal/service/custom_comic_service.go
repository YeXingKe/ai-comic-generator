package service

import (
	"archive/zip"
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ai-comic-generator/server/internal/client/cos"
	"github.com/ai-comic-generator/server/internal/common"
	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/pkg/llmjson"
	"github.com/ai-comic-generator/server/internal/storage"
	"github.com/ai-comic-generator/server/internal/store"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
)

const (
	customPanelCountMin = 2
	customPanelCountMax = 8
	customPanelDefault  = 4
)

// CustomComicService 自定义创作：一次多格生图
type CustomComicService struct {
	store      *store.CustomComicStore
	userStore  *store.UserStore
	localStore *storage.Local
	generators map[string]ImageGenerator
	cos        *cos.Client
	llm        llms.Model
}

func NewCustomComicService(
	customStore *store.CustomComicStore,
	userStore *store.UserStore,
	localStore *storage.Local,
	generators map[string]ImageGenerator,
	cosClient *cos.Client,
	llm llms.Model,
) *CustomComicService {
	return &CustomComicService{
		store:      customStore,
		userStore:  userStore,
		localStore: localStore,
		generators: generators,
		cos:        cosClient,
		llm:        llm,
	}
}

// Create 创建任务并异步生成多格
func (s *CustomComicService) Create(userID int64, req *model.CreateCustomComicRequest) (string, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return "", common.ErrParams.WithMessage("提示词不能为空")
	}

	aspectRatio := req.AspectRatio
	if aspectRatio == "" {
		aspectRatio = model.AspectRatio16x9
	}
	if !isValidAspectRatio(aspectRatio) {
		return "", common.ErrParams.WithMessage("画幅仅支持 1:1 / 16:9 / 9:16 / 2:3")
	}

	imageBackend := req.ImageBackend
	if imageBackend == "" {
		imageBackend = common.ImageBackendHunyuan
	}
	if !isValidImageBackend(imageBackend) {
		return "", common.ErrParams.WithMessage("不支持的生图模型")
	}

	panelCount := req.PanelCount
	if panelCount <= 0 {
		panelCount = customPanelDefault
	}
	if panelCount < customPanelCountMin || panelCount > customPanelCountMax {
		return "", common.ErrParams.WithMessage(fmt.Sprintf("格数需在 %d–%d 之间", customPanelCountMin, customPanelCountMax))
	}

	taskID := uuid.NewString()
	comic := &model.CustomComic{
		TaskID:       taskID,
		UserID:       userID,
		Prompt:       prompt,
		AspectRatio:  aspectRatio,
		ImageBackend: imageBackend,
		PanelCount:   panelCount,
		Status:       model.CustomComicStatusPending,
	}
	if err := s.store.Create(comic); err != nil {
		return "", common.ErrOperation.WithMessage("创建自定义任务失败")
	}

	go func() {
		ctx := context.Background()
		if err := s.runGenerate(ctx, comic); err != nil {
			log.Printf("custom comic generate failed taskId=%s err=%v", taskID, err)
		}
	}()

	return taskID, nil
}

// GetForUser 查询任务详情（本人或管理员）
func (s *CustomComicService) GetForUser(taskID string, userID int64, isAdmin bool) (*model.CustomComicInfo, error) {
	comic, err := s.store.GetByTaskID(taskID)
	if err != nil {
		return nil, common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return nil, common.ErrNoAuth
	}
	info := comic.ToCustomComicInfo()
	s.attachUsers([]*model.CustomComicInfo{info})
	return info, nil
}

// WritePanelsZip 将任务全部分镜图打包为 zip 写入 w，返回下载文件名
func (s *CustomComicService) WritePanelsZip(w io.Writer, taskID string, userID int64, isAdmin bool) (string, error) {
	comic, err := s.store.GetByTaskID(taskID)
	if err != nil {
		return "", common.ErrNotFound
	}
	if !isAdmin && comic.UserID != userID {
		return "", common.ErrNoAuth
	}
	info := comic.ToCustomComicInfo()
	if len(info.PanelImages) == 0 {
		return "", common.ErrOperation.WithMessage("暂无可下载的分镜图片")
	}

	zw := zip.NewWriter(w)
	defer zw.Close()

	client := &http.Client{Timeout: 60 * time.Second}
	added := 0
	for _, panel := range info.PanelImages {
		name := fmt.Sprintf("panel_%d.png", panel.PanelNo)
		data, readErr := s.readPanelBytes(taskID, panel, client)
		if readErr != nil {
			log.Printf("custom comic zip skip panel taskId=%s panel=%d: %v", taskID, panel.PanelNo, readErr)
			continue
		}
		fw, createErr := zw.Create(name)
		if createErr != nil {
			return "", common.ErrSystem
		}
		if _, writeErr := fw.Write(data); writeErr != nil {
			return "", common.ErrSystem
		}
		added++
	}
	if added == 0 {
		return "", common.ErrOperation.WithMessage("分镜图片读取失败，无法打包")
	}

	filename := fmt.Sprintf("custom-comic-%s.zip", shortTaskID(taskID))
	return filename, nil
}

func (s *CustomComicService) readPanelBytes(taskID string, panel model.PanelImageResult, client *http.Client) ([]byte, error) {
	localPath := s.localStore.CustomPanelPath(taskID, panel.PanelNo)
	if data, err := os.ReadFile(localPath); err == nil && len(data) > 0 {
		return data, nil
	}
	url := strings.TrimSpace(panel.URL)
	if url == "" {
		return nil, fmt.Errorf("empty panel url")
	}
	// 相对路径拼本地文件已失败，尝试 HTTP（COS / 本机静态）
	if strings.HasPrefix(url, "/") {
		return nil, fmt.Errorf("local file missing: %s", localPath)
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func shortTaskID(taskID string) string {
	id := strings.ReplaceAll(taskID, "-", "")
	if len(id) > 8 {
		return id[:8]
	}
	if id == "" {
		return "panels"
	}
	return id
}

// ListByPage 分页查询自定义创作（非管理员由 Handler 强制 userId）
func (s *CustomComicService) ListByPage(req *model.QueryCustomComicRequest) (*model.CustomComicPageResult, error) {
	if req.PageNum <= 0 {
		req.PageNum = common.DefaultPageNum
	}
	if req.PageSize <= 0 {
		req.PageSize = common.DefaultPageSize
	}
	if req.PageSize > common.MaxPageSize {
		req.PageSize = common.MaxPageSize
	}
	page, err := s.store.ListByPage(req)
	if err != nil {
		return nil, common.ErrSystem
	}
	infos := make([]*model.CustomComicInfo, 0, len(page.Records))
	for i := range page.Records {
		infos = append(infos, &page.Records[i])
	}
	s.attachUsers(infos)
	return page, nil
}

func (s *CustomComicService) attachUsers(infos []*model.CustomComicInfo) {
	if len(infos) == 0 || s.userStore == nil {
		return
	}
	idSet := make(map[int64]struct{}, len(infos))
	ids := make([]int64, 0, len(infos))
	for _, info := range infos {
		if info == nil || info.UserID <= 0 {
			continue
		}
		if _, ok := idSet[info.UserID]; ok {
			continue
		}
		idSet[info.UserID] = struct{}{}
		ids = append(ids, info.UserID)
	}
	if len(ids) == 0 {
		return
	}
	users, err := s.userStore.ListByIDs(ids)
	if err != nil {
		log.Printf("custom comic attachUsers: list users failed: %v", err)
		return
	}
	byID := make(map[int64]*model.User, len(users))
	for i := range users {
		byID[users[i].ID] = &users[i]
	}
	for _, info := range infos {
		if info == nil {
			continue
		}
		u, ok := byID[info.UserID]
		if !ok {
			continue
		}
		info.UserAccount = u.UserAccount
		info.UserName = u.UserName
	}
}

func (s *CustomComicService) runGenerate(ctx context.Context, comic *model.CustomComic) error {
	taskID := comic.TaskID
	if err := s.store.UpdateStatus(taskID, model.CustomComicStatusProcessing, nil); err != nil {
		return err
	}
	if err := s.localStore.EnsureCustomTaskDir(taskID); err != nil {
		_ = s.store.MarkFailed(taskID, err.Error())
		return err
	}

	prompts, err := s.splitPanelPrompts(ctx, comic.Prompt, comic.PanelCount, comic.AspectRatio)
	if err != nil {
		log.Printf("custom comic split prompts fallback taskId=%s err=%v", taskID, err)
		prompts = fallbackPanelPrompts(comic.Prompt, comic.PanelCount, comic.AspectRatio)
	}

	generator := s.resolveGenerator(comic.ImageBackend)
	sizeHint := aspectRatioToGeneratorSize(comic.AspectRatio, comic.ImageBackend)
	results := make([]model.PanelImageResult, 0, comic.PanelCount)

	for i := 0; i < comic.PanelCount; i++ {
		panelNo := i + 1
		imagePrompt := prompts[i]
		dest := s.localStore.CustomPanelPath(taskID, panelNo)

		var genErr error
		if generator != nil {
			if sized, ok := generator.(SizedImageGenerator); ok && sizeHint != "" {
				genErr = sized.GenerateWithSize(ctx, imagePrompt, dest, sizeHint)
			} else {
				genErr = generator.Generate(ctx, imagePrompt, dest)
			}
		} else {
			log.Printf("custom comic generator disabled, placeholder: taskId=%s panel=%d", taskID, panelNo)
			genErr = renderCustomPlaceholder(dest, comic.AspectRatio, panelNo, imagePrompt)
		}
		if genErr != nil {
			_ = s.store.MarkFailed(taskID, fmt.Sprintf("panel %d: %v", panelNo, genErr))
			return fmt.Errorf("panel %d: %w", panelNo, genErr)
		}

		url := s.localStore.CustomPublicURL(taskID, fmt.Sprintf("panel_%d.png", panelNo))
		if s.cos != nil && s.cos.Enabled() {
			cosKey := fmt.Sprintf("comics/custom/%s/panel_%d.png", taskID, panelNo)
			if cosURL, uploadErr := s.cos.UploadFile(ctx, cosKey, dest); uploadErr != nil {
				log.Printf("custom comic cos upload failed taskId=%s panel=%d: %v", taskID, panelNo, uploadErr)
			} else {
				url = cosURL
			}
		}

		results = append(results, model.PanelImageResult{
			PanelNo:     panelNo,
			URL:         url,
			Method:      panelImageMethod(generator != nil),
			ImagePrompt: imagePrompt,
		})
		if saveErr := s.store.SavePanelImages(taskID, results); saveErr != nil {
			log.Printf("custom comic save panels failed taskId=%s: %v", taskID, saveErr)
		}
	}

	return s.store.MarkCompleted(taskID, results)
}

func (s *CustomComicService) resolveGenerator(backend string) ImageGenerator {
	if backend == "" {
		backend = common.ImageBackendHunyuan
	}
	gen, ok := s.generators[backend]
	if !ok || gen == nil || !gen.Enabled() {
		return nil
	}
	return gen
}

type panelPromptList struct {
	Panels []struct {
		PanelNo     int    `json:"panelNo"`
		ImagePrompt string `json:"imagePrompt"`
	} `json:"panels"`
}

func (s *CustomComicService) splitPanelPrompts(ctx context.Context, userPrompt string, panelCount int, aspectRatio string) ([]string, error) {
	if s.llm == nil {
		return fallbackPanelPrompts(userPrompt, panelCount, aspectRatio), nil
	}
	aspectHint := aspectRatioPromptHint(aspectRatio)
	meta := fmt.Sprintf(`You are a comic storyboard artist. Split the user's idea into exactly %d sequential comic panels.
Return ONLY valid JSON:
{"panels":[{"panelNo":1,"imagePrompt":"..."},...]}

Rules for each imagePrompt:
- English keyword-style, under 180 characters
- Include: subject, action, scene, lighting, cartoon comic style
- Must include aspect framing: %s
- No text, letters, speech bubbles, watermark
- Panels should form a coherent short sequence from the same idea

User idea:
%s`, panelCount, aspectHint, userPrompt)

	content, err := llms.GenerateFromSinglePrompt(ctx, s.llm, meta)
	if err != nil {
		return nil, err
	}
	var parsed panelPromptList
	if err := llmjson.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Panels) < panelCount {
		return nil, fmt.Errorf("expected %d panels, got %d", panelCount, len(parsed.Panels))
	}
	out := make([]string, panelCount)
	for i := 0; i < panelCount; i++ {
		p := strings.TrimSpace(parsed.Panels[i].ImagePrompt)
		if p == "" {
			return nil, fmt.Errorf("empty imagePrompt at panel %d", i+1)
		}
		if !strings.Contains(strings.ToLower(p), "comic") {
			p = p + ", cartoon comic panel, " + aspectHint
		}
		out[i] = common.TruncateHunyuanPrompt(common.SanitizeHunyuanImagePrompt(p))
	}
	return out, nil
}

func fallbackPanelPrompts(userPrompt string, panelCount int, aspectRatio string) []string {
	hint := aspectRatioPromptHint(aspectRatio)
	out := make([]string, panelCount)
	for i := 0; i < panelCount; i++ {
		out[i] = common.TruncateHunyuanPrompt(fmt.Sprintf(
			"cartoon comic panel, %s, panel %d of %d sequence, %s, no text, no watermark",
			hint, i+1, panelCount, strings.TrimSpace(userPrompt),
		))
	}
	return out
}

func isValidAspectRatio(v string) bool {
	switch v {
	case model.AspectRatio1x1, model.AspectRatio16x9, model.AspectRatio9x16, model.AspectRatio2x3:
		return true
	default:
		return false
	}
}

func isValidImageBackend(v string) bool {
	switch v {
	case common.ImageBackendHunyuan, common.ImageBackendOpenAIImage1K, common.ImageBackendOpenAIImage4K:
		return true
	default:
		return false
	}
}

func aspectRatioPromptHint(aspectRatio string) string {
	switch aspectRatio {
	case model.AspectRatio1x1:
		return "square 1:1 composition"
	case model.AspectRatio9x16:
		return "vertical 9:16 portrait composition"
	case model.AspectRatio2x3:
		return "vertical 2:3 Xiaohongshu / RedNote feed composition"
	default:
		return "horizontal 16:9 cinematic widescreen"
	}
}

// aspectRatioToGeneratorSize 映射到各后端尺寸参数
// GPT: 1024x1024 / 1792x1024 / 1024x1792 / 1024x1536(2:3)
// 混元: 1024:1024 / 1920:1080 / 1080:1920 / 1024:1536(2:3)
func aspectRatioToGeneratorSize(aspectRatio, imageBackend string) string {
	isOpenAI := imageBackend == common.ImageBackendOpenAIImage1K || imageBackend == common.ImageBackendOpenAIImage4K
	switch aspectRatio {
	case model.AspectRatio1x1:
		if isOpenAI {
			return "1024x1024"
		}
		return "1024:1024"
	case model.AspectRatio9x16:
		if isOpenAI {
			return "1024x1792"
		}
		return "1080:1920"
	case model.AspectRatio2x3:
		if isOpenAI {
			return "1024x1536"
		}
		return "1024:1536"
	default:
		if isOpenAI {
			return "1792x1024"
		}
		return "1920:1080"
	}
}

func renderCustomPlaceholder(path, aspectRatio string, panelNo int, prompt string) error {
	w, h := 960, 540
	switch aspectRatio {
	case model.AspectRatio1x1:
		w, h = 768, 768
	case model.AspectRatio9x16:
		w, h = 540, 960
	case model.AspectRatio2x3:
		w, h = 640, 960
	}
	dc := gg.NewContext(w, h)
	dc.SetColor(color.RGBA{245, 240, 255, 255})
	dc.Clear()
	dc.SetColor(color.RGBA{120, 100, 220, 255})
	dc.DrawRectangle(16, 16, float64(w-32), float64(h-32))
	dc.SetLineWidth(3)
	dc.Stroke()
	dc.SetColor(color.Black)
	if err := dc.LoadFontFace("C:/Windows/Fonts/msyh.ttc", 22); err != nil {
		_ = dc.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 22)
	}
	dc.DrawStringAnchored(fmt.Sprintf("Panel %d", panelNo), float64(w/2), 48, 0.5, 0.5)
	y := 100.0
	for _, line := range wordWrap(prompt, 22) {
		dc.DrawStringAnchored(line, float64(w/2), y, 0.5, 0)
		y += 28
		if y > float64(h-40) {
			break
		}
	}
	return dc.SavePNG(path)
}
