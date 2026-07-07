package common

import (
	"fmt"
	"strconv"
	"strings"
)

// AI Prompt 模板常量（漫画生成六步流水线）

// ---------- 漫画风格常量（与 model.CreateComicRequest.style 对齐） ----------

const (
	ComicStyleCartoon   = "cartoon"   // 卡通
	ComicStyleRealistic = "realistic" // 写实
	ComicStyleChibi     = "chibi"     // Q 版
)

// ---------- 步骤 0：标题推荐 ----------

// TitleIdeationPrompt 标题推荐 Agent
// 占位符：{topic} {style} {descriptionSection} {stylePrompt}
const TitleIdeationPrompt = `你是一位擅长公众号传播的漫画标题策划，擅长根据选题生成多个吸引眼球的漫画标题方案。

根据以下选题，生成多个漫画标题推荐：
选题：{topic}
漫画风格：{style}
{descriptionSection}
{stylePrompt}

要求：
1. 生成 4 个风格各异但都与选题紧密相关的标题方案
2. 每个 title 不超过 20 字，适合四格/六格漫画，朗朗上口
3. subtitle 为 10-25 字的卖点或副标题，说明该标题的吸引力（可幽默、悬念、温情等）
4. 标题避免血腥、低俗、敏感政治内容
5. 四个标题角度要有差异（如：搞笑向、温情向、悬念向、热血向）

请直接返回 JSON 格式，不要有 markdown 代码块或其他说明文字：
{
  "options": [
    { "title": "标题1", "subtitle": "卖点说明1" },
    { "title": "标题2", "subtitle": "卖点说明2" },
    { "title": "标题3", "subtitle": "卖点说明3" },
    { "title": "标题4", "subtitle": "卖点说明4" }
  ]
}`

// ConfirmedTitleSection 用户已确认标题时插入故事构思 Prompt
const ConfirmedTitleSection = `

用户已确认的漫画标题：{confirmedTitle}
故事构思 JSON 中的 title 字段必须使用该标题，不得修改。`

// ---------- 步骤 1：故事构思 ----------

// StoryIdeationPrompt 故事构思 Agent（Agent1）
// 占位符：{topic} {style} {descriptionSection} {stylePrompt}
const StoryIdeationPrompt = `你是一位资深漫画编剧，擅长将简短选题扩展为结构完整、适合连载或单篇四格/六格漫画的故事方案。

根据以下选题，完成故事构思：
选题：{topic}
漫画风格：{style}
{descriptionSection}
{stylePrompt}

要求：
1. 故事要适合漫画表现，情节清晰、节奏紧凑，能在 4-6 格内讲完或留有续作空间
2. synopsis 用 80-150 字概括完整故事线，包含起承转合
3. theme 提炼一个核心主题（如成长、友情、逆袭、治愈）
4. tone 明确基调（搞笑、热血、治愈、悬疑、温情等），与选题和风格一致
5. title 给出 1 个吸引人的漫画标题，不超过 20 字，适合公众号传播
6. keyConflict 写清主角面临的核心矛盾或挑战
7. highlights 列出 2-4 个适合画成漫画画面的亮点情节（每句不超过 30 字）
8. 避免血腥、低俗、敏感政治内容；面向大众读者

请直接返回 JSON 格式，不要有 markdown 代码块或其他说明文字：
{
  "synopsis": "故事梗概，包含起因、发展、高潮与结局",
  "theme": "核心主题",
  "tone": "基调",
  "title": "漫画标题",
  "keyConflict": "核心冲突描述",
  "highlights": [
    "亮点情节 1",
    "亮点情节 2",
    "亮点情节 3"
  ]
}`

// StoryDescriptionSection 用户补充描述（动态插入 StoryIdeationPrompt / StoryboardScriptPrompt）
const StoryDescriptionSection = `

用户补充要求：{userDescription}
请在故事构思中充分体现用户的补充要求。`

// ---------- 步骤 2：角色设定 ----------

// CharacterDesignPrompt 角色设定 Agent（Agent2）
// 占位符：{storyIdeation} {style} {stylePrompt}
const CharacterDesignPrompt = `你是一位专业漫画角色设计师，擅长为故事设计辨识度强、适合 AI 绘制的角色形象。

根据以下故事构思，设计漫画角色：
故事构思（JSON）：
{storyIdeation}
漫画风格：{style}
{stylePrompt}

要求：
1. 设计 2-4 个角色，必须包含 1 名主角（role 为 protagonist）
2. 可有 0-1 名反派（antagonist）和 1-2 名配角（supporting）
3. name 使用中文角色名，简短好记
4. appearance 详细描述外貌、服装、配色、标志性道具，便于后续分镜与生图保持一致（80-120 字）
5. personality 用 2-3 句话概括性格与说话习惯，便于撰写台词
6. avatarUrl 暂时留空字符串 ""（图片由后续步骤生成）
7. 角色外貌描述需与漫画风格一致（卡通/Q 版/写实）
8. 各角色之间要有视觉区分度（发型、服装、体型不要雷同）

请直接返回 JSON 数组格式，不要有 markdown 代码块或其他说明文字：
[
  {
    "name": "角色名",
    "role": "protagonist",
    "appearance": "外貌与服装详细描述",
    "personality": "性格与说话习惯",
    "avatarUrl": ""
  },
  {
    "name": "角色名",
    "role": "supporting",
    "appearance": "外貌与服装详细描述",
    "personality": "性格与说话习惯",
    "avatarUrl": ""
  }
]`

// ---------- 步骤 3：分镜脚本 ----------

// StoryboardScriptPrompt 分镜脚本 Agent（Agent3）
// 占位符：{storyIdeation} {characters} {style} {panelCount} {descriptionSection} {stylePrompt}
const StoryboardScriptPrompt = `你是一位经验丰富的漫画分镜师，擅长将故事拆解为适合公众号发布的四格/六格漫画脚本。

根据以下故事与角色，编写分镜脚本：
故事构思（JSON）：
{storyIdeation}
角色设定（JSON 数组）：
{characters}
漫画风格：{style}
分镜格数：{panelCount} 格（请严格输出该数量的 panels）
{descriptionSection}
{stylePrompt}

要求：
1. 共 {panelCount} 格，panelNo 从 1 连续编号到 {panelCount}
2. 第 1 格作为引子或场景建立，最后一格作为 punchline、反转或温情收尾
3. scene 描述画面内容、角色动作、表情、背景环境，具体到可作画（每格 50-100 字）
4. dialogue 为该格角色台词数组，每句不超过 20 字，符合角色 personality；无台词时返回空数组 []
5. narration 旁白，无则留空字符串 ""
6. camera 标明镜头：特写 / 中景 / 全景 / 俯视 / 仰视 / 过肩等
7. imagePrompt 必须为英文，描述该格画面供 AI 生图使用，包含：角色外貌关键词、动作、场景、光影、构图、风格词；必须注明 horizontal 16:9 cinematic widescreen comic panel（电影宽银幕比例单格，非竖版条漫）
8. 若该格有 dialogue 或 narration，imagePrompt 必须描述漫画式对白气泡：speech bubble / thought bubble 位于说话角色头顶或嘴边，bubble tail 指向该角色；有台词时在 Prompt 中注明 bubble contains Chinese text（不要把台词写进 scene 中文描述里重复一遍）
9. 角色外貌与 dialogue 须与角色设定一致，勿凭空新增未设定角色
10. pageCount 填写 1（单页四格/六格）或实际页数

请直接返回 JSON 格式，不要有 markdown 代码块或其他说明文字：
{
  "pageCount": 1,
  "panels": [
    {
      "panelNo": 1,
      "scene": "画面与动作描述",
      "dialogue": ["角色A：台词"],
      "narration": "",
      "camera": "中景",
      "imagePrompt": "English prompt for AI image generation, cartoon style, ..."
    },
    {
      "panelNo": 2,
      "scene": "画面与动作描述",
      "dialogue": [],
      "narration": "旁白文字",
      "camera": "特写",
      "imagePrompt": "English prompt..."
    }
  ]
}`

// AiModifyStoryboardPrompt AI 修改分镜（用户人工编辑时可选调用）
// 占位符：{storyIdeation} {currentStoryboard} {modifySuggestion}
const AiModifyStoryboardPrompt = `你是一位专业的漫画分镜师，擅长根据用户反馈优化分镜脚本。

当前故事构思：
{storyIdeation}

当前分镜脚本：
{currentStoryboard}

用户修改建议：
{modifySuggestion}

要求：
1. 根据用户建议调整分镜，保持 panelNo 从 1 连续编号
2. 保持 JSON 结构与字段名不变（panelNo, scene, dialogue, narration, camera, imagePrompt）
3. 若用户要求增删格数，同步调整 panels 数组长度与 pageCount
4. imagePrompt 保持英文，适合 AI 生图
5. 修改后剧情仍须与故事构思一致

请直接返回修改后的完整分镜 JSON，不要有 markdown 代码块或其他说明文字：
{
  "pageCount": 1,
  "panels": [
    {
      "panelNo": 1,
      "scene": "画面与动作描述",
      "dialogue": ["台词"],
      "narration": "",
      "camera": "中景",
      "imagePrompt": "English prompt..."
    }
  ]
}`

// ---------- 步骤 4：图片生成（Prompt 增强，供 image_service 使用） ----------

// PanelImageEnhancePrompt 在分镜 imagePrompt 基础上增强生图指令（含漫画对白气泡）
// 占位符：{style} {stylePrompt} {imagePrompt} {scene} {characters} {dialogue} {narration}
const PanelImageEnhancePrompt = `### 任务 ###
你是一位专业漫画分镜画师，负责将分镜信息转化为一条「可直接用于 AI 绘画模型」的英文 Prompt。
目标画面必须是**带对白气泡的漫画单格**（manga/comic panel with speech bubbles），气泡在角色头顶或嘴边，有指向角色的尾巴。

### 输入 ###
漫画风格：{style}
分镜画面描述：{scene}
角色参考：{characters}
本格台词（dialogue）：{dialogue}
本格旁白（narration）：{narration}
原始 Prompt：{imagePrompt}
{stylePrompt}

### 漫画气泡要求（必须遵守） ###
1. 画面为单格漫画构图（single comic panel），**横版 16:9 电影宽银幕比例**（horizontal cinematic 16:9 widescreen frame）
2. 若本格有台词：在**说话角色头顶上方**绘制 classic manga speech bubble（白色填充、黑色描边、圆角椭圆或圆角矩形），bubble tail 清晰指向该角色嘴部或头部
3. 气泡内必须显示台词原文，使用**清晰可读的中文汉字**（legible Chinese characters），文字居中，字号适中，不要乱码、不要英文替代
4. 若有多句台词：主说话者一个大气泡；另一角色可用较小气泡在画面另一侧，避免重叠遮挡面部
5. 若仅有旁白无台词：使用矩形 narration box（旁白框）通常在格子上/下边，内含中文旁白
6. 若既无台词也无旁白：仍保留漫画分格感，但不必强行加空气泡
7. 角色表情、口型与气泡情绪一致（惊喜/吐槽/愤怒等）

### 画面质量要求 ###
1. 输出为一条完整英文 Prompt，不要 JSON，不要解释，不要 markdown
2. 必须包含：horizontal 16:9 cinematic widescreen comic panel, clean line art, speech bubbles with Chinese text, 风格词、角色一致性、场景、光影、构图
3. 适合公众号条漫长图，角色肢体正常，避免模糊、畸形手指、多余 limbs
4. 气泡内中文须与输入 dialogue/narration **完全一致**，不得改写、不得遗漏
5. 长度控制在 120-220 个英文单词

### 输出 ###
直接返回最终英文 Prompt，不要有其他内容。`

// ---------- 步骤 5：排版合成（文案辅助，可选 LLM 生成图注） ----------

// LayoutCaptionPrompt 为排版后的漫画生成简短图注/导读（公众号图文导语）
// 占位符：{title} {synopsis} {panelCount}
const LayoutCaptionPrompt = `你是一位公众号漫画编辑，擅长撰写吸引读者点击的漫画导读文案。

漫画标题：{title}
故事梗概：{synopsis}
分镜格数：{panelCount}

要求：
1. 输出 1 段导读，50-80 字，口语化、有悬念或笑点，适合放在公众号文章开头
2. 不要使用「本文」「小编」等套话
3. 不要剧透结局

请直接返回纯文本导读，不要有标题、JSON 或其他格式。`

// ---------- 步骤 6：公众号发布（图文摘要） ----------

// WechatPublishCopyPrompt 生成公众号群发所需的标题与摘要
// 占位符：{title} {synopsis} {theme} {tone}
const WechatPublishCopyPrompt = `你是一位微信公众号运营专家，擅长为漫画内容撰写传播性强的标题与摘要。

漫画标题：{title}
故事梗概：{synopsis}
主题：{theme}
基调：{tone}

要求：
1. 输出 JSON，包含 title（群发标题，不超过 32 字）、digest（摘要，不超过 80 字）、tags（3-5 个标签，字符串数组）
2. 标题要有吸引力，可适当使用疑问、反差、数字，但不做标题党
3. 摘要概括故事看点，引导读者点开长图，不剧透结局
4. tags 与漫画主题相关，便于分类

请直接返回 JSON 格式，不要有 markdown 代码块或其他说明文字：
{
  "title": "公众号文章标题",
  "digest": "文章摘要",
  "tags": ["标签1", "标签2", "标签3"]
}`

// ---------- 漫画风格附加 Prompt（运行时按 style 字段拼接） ----------

const StyleCartoonPrompt = `

**重要：请使用卡通漫画风格进行创作与描述**
- 造型圆润、线条简洁、色彩明快
- 表情适度夸张，适合四格搞笑或温情向
- 背景可适度简化，突出角色与动作
- 分镜 imagePrompt 中需包含 cartoon style, clean line art, vibrant colors, speech bubble above character head 等风格词`

const StyleRealisticPrompt = `

**重要：请使用写实漫画风格进行创作与描述**
- 人物比例接近真实，细节丰富（服饰纹理、光影）
- 场景透视准确，氛围感强，偏电影分镜质感
- 台词克制，适合剧情向、悬疑或都市题材
- 分镜 imagePrompt 中需包含 realistic comic style, detailed shading, cinematic lighting, speech bubble 等风格词`

const StyleChibiPrompt = `

**重要：请使用 Q 版（Chibi）风格进行创作与描述**
- 二头身或三头身比例，头大身小，可爱化
- 表情极度夸张，适合萌系、搞笑、轻松治愈
- 道具与场景可可爱化变形，色彩粉嫩或高饱和
- 分镜 imagePrompt 中需包含 chibi style, super deformed, cute, kawaii, 2-3 head ratio, cute speech bubble 等风格词`

// GetComicStylePrompt 根据漫画风格返回附加 Prompt 片段（空或未知风格返回空字符串）
func GetComicStylePrompt(style string) string {
	switch style {
	case ComicStyleCartoon:
		return StyleCartoonPrompt
	case ComicStyleRealistic:
		return StyleRealisticPrompt
	case ComicStyleChibi:
		return StyleChibiPrompt
	default:
		return StyleCartoonPrompt // 未知风格默认卡通
	}
}

// BuildDescriptionSection 根据用户补充描述构建可插入 Prompt 的片段
func BuildDescriptionSection(userDescription string) string {
	if userDescription == "" {
		return ""
	}
	return strings.ReplaceAll(StoryDescriptionSection, "{userDescription}", userDescription)
}

// BuildTitleIdeationPrompt 组装标题推荐完整 Prompt
func BuildTitleIdeationPrompt(topic, style, userDescription string) string {
	prompt := strings.ReplaceAll(TitleIdeationPrompt, "{topic}", topic)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", BuildDescriptionSection(userDescription))
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePrompt(style))
	return prompt
}

// BuildStoryIdeationPrompt 组装故事构思完整 Prompt
func BuildStoryIdeationPrompt(topic, style, userDescription, confirmedTitle string) string {
	prompt := strings.ReplaceAll(StoryIdeationPrompt, "{topic}", topic)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	descSection := BuildDescriptionSection(userDescription)
	if confirmedTitle != "" {
		descSection += strings.ReplaceAll(ConfirmedTitleSection, "{confirmedTitle}", confirmedTitle)
	}
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", descSection)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePrompt(style))
	return prompt
}

// BuildCharacterDesignPrompt 组装角色设定完整 Prompt
func BuildCharacterDesignPrompt(storyIdeationJSON, style string) string {
	prompt := strings.ReplaceAll(CharacterDesignPrompt, "{storyIdeation}", storyIdeationJSON)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePrompt(style))
	return prompt
}

// BuildStoryboardScriptPrompt 组装分镜脚本完整 Prompt
func BuildStoryboardScriptPrompt(storyIdeationJSON, charactersJSON, style, userDescription string, panelCount int) string {
	if panelCount <= 0 {
		panelCount = 4
	}
	countStr := strconv.Itoa(panelCount)
	prompt := strings.ReplaceAll(StoryboardScriptPrompt, "{storyIdeation}", storyIdeationJSON)
	prompt = strings.ReplaceAll(prompt, "{characters}", charactersJSON)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{panelCount}", countStr)
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", BuildDescriptionSection(userDescription))
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePrompt(style))
	return prompt
}

// BuildPanelImageEnhancePrompt 组装混元生图 Prompt（含对白气泡与中文台词）
func BuildPanelImageEnhancePrompt(style, scene, characters, imagePrompt, dialogue, narration string) string {
	if dialogue == "" {
		dialogue = "（无）"
	}
	if narration == "" {
		narration = "（无）"
	}
	prompt := strings.ReplaceAll(PanelImageEnhancePrompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{scene}", scene)
	prompt = strings.ReplaceAll(prompt, "{characters}", characters)
	prompt = strings.ReplaceAll(prompt, "{imagePrompt}", imagePrompt)
	prompt = strings.ReplaceAll(prompt, "{dialogue}", dialogue)
	prompt = strings.ReplaceAll(prompt, "{narration}", narration)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePrompt(style))
	return prompt
}

// BuildDirectPanelImagePrompt 不经 LLM，直接拼装含对白气泡的英文生图 Prompt（降级/兜底）
func BuildDirectPanelImagePrompt(style, scene, characters, imagePrompt, dialogue, narration string) string {
	parts := make([]string, 0, 12)
	if style != "" {
		parts = append(parts, style+" comic style")
	} else {
		parts = append(parts, "cartoon comic style")
	}
	parts = append(parts,
		"single manga comic panel",
		"horizontal 16:9 aspect ratio",
		"clean line art",
		"vibrant colors",
		"professional webtoon panel",
	)
	if imagePrompt != "" {
		parts = append(parts, imagePrompt)
	}
	if scene != "" {
		parts = append(parts, "scene: "+scene)
	}
	if characters != "" {
		parts = append(parts, "characters: "+characters)
	}
	if dialogue != "" {
		parts = append(parts,
			"classic manga speech bubble above speaking character head",
			"white bubble with black outline",
			"bubble tail pointing to character mouth",
			fmt.Sprintf(`legible Chinese text inside bubble: "%s"`, dialogue),
		)
	} else if narration != "" {
		parts = append(parts,
			"rectangular narration box at top of panel",
			fmt.Sprintf(`Chinese narration text: "%s"`, narration),
		)
	}
	parts = append(parts, "no deformed hands, no blurry face, high quality illustration")
	return strings.Join(parts, ", ")
}
