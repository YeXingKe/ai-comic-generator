package common

import (
	"strconv"
	"strings"
)

// AI Prompt 模板常量（漫画生成六步流水线）

// ---------- 漫画风格常量（与 model.CreateComicRequest.style 对齐） ----------

const (
	ComicStyleCartoon   = "cartoon"   // 卡通
	ComicStyleRealistic = "realistic" // 写实
	ComicStyleChibi     = "chibi"     // Q 版
	ComicStyleAnimal    = "animal"    // 动物拟人条漫（扁平插画、顶栏字幕）
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
7. 角色外貌描述需与漫画风格一致（卡通/Q 版/写实/动物拟人）
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
7. imagePrompt 必须为英文关键词式短句（非长段落），总长不超过 180 个英文字符；包含：角色外貌关键词、动作、场景、光影、风格词；必须注明 horizontal 16:9 cinematic widescreen comic panel
8. imagePrompt 只描述纯画面（角色外貌、动作、场景、光影、构图），**禁止** speech bubble、文字、中文/英文台词；画面上方留白，中文台词由程序后续在顶部叠加字幕
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

// PanelImageEnhancePrompt 在分镜 imagePrompt 基础上增强生图指令（画面不含文字，台词由程序叠加顶栏字幕）
// 占位符：{style} {stylePrompt} {imagePrompt} {scene} {characters} {dialogue} {narration}
const PanelImageEnhancePrompt = `### 任务 ###
你是一位专业漫画分镜画师，负责将分镜信息转化为一条「可直接用于 AI 绘画模型」的英文 Prompt。
**重要：生图模型不渲染文字。Prompt 中禁止出现任何中文/英文台词、禁止要求绘制 speech bubble 或文字框。**
台词与旁白将由后续程序在画面顶部居中叠加中文黑字字幕（无气泡），你只需描述纯画面（角色、动作、场景、构图）。

### 输入 ###
漫画风格：{style}
分镜画面描述：{scene}
角色参考：{characters}
本格台词（dialogue，仅作理解情绪，勿写入 Prompt）：{dialogue}
本格旁白（narration，仅作理解，勿写入 Prompt）：{narration}
原始 Prompt：{imagePrompt}
{stylePrompt}

### 画面要求 ###
1. single comic panel, horizontal 16:9 cinematic widescreen frame
2. 角色表情、动作与台词情绪一致，但**画面中不得出现任何文字、字母、数字、水印、乱码**
3. 画面上方约 20% 区域构图留白，便于后续叠加顶栏字幕（leave top area clear, no text, no speech bubble）
4. clean line art, 风格词、角色一致性、场景、光影

### 输出要求 ###
1. 输出一条英文 keyword Prompt，不要 JSON，不要 markdown
2. **总长度不得超过 200 个字符**
3. 必须包含：no text, no letters, no watermark

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
- 分镜 imagePrompt 中需包含 cartoon style, clean line art, vibrant colors, leave top area clear, no text 等风格词`

const StyleRealisticPrompt = `

**重要：请使用写实漫画风格进行创作与描述**
- 人物比例接近真实，细节丰富（服饰纹理、光影）
- 场景透视准确，氛围感强，偏电影分镜质感
- 台词克制，适合剧情向、悬疑或都市题材
- 分镜 imagePrompt 中需包含 realistic comic style, detailed shading, cinematic lighting, leave top area clear, no text 等风格词`

const StyleChibiPrompt = `

**重要：请使用 Q 版（Chibi）风格进行创作与描述**
- 二头身或三头身比例，头大身小，可爱化
- 表情极度夸张，适合萌系、搞笑、轻松治愈
- 道具与场景可可爱化变形，色彩粉嫩或高饱和
- 分镜 imagePrompt 中需包含 chibi style, super deformed, cute, kawaii, 2-3 head ratio, leave top area clear, no text 等风格词`

const StyleAnimalPrompt = `

**重要：请使用「动物拟人条漫」风格进行创作与描述（参考公众号治愈/幽默动物漫画，如拟人青蛙家庭日常）**
- 角色全部为拟人化动物（青蛙、猫、兔、熊、狐狸、狗等），直立行走、有手有脚、表情拟人；**禁止出现人类角色**
- 造型：粗黑描边、扁平 2D 填色、低饱和或明快纯色；背景极简（如一色墙面 + 地面线、简单家具）
- 场景偏日常：卧室、客厅、办公室、草地等，适合四格温情/轻幽默/家庭互动
- 角色 appearance 必须写明：物种 + 肤色/毛色 + 体型 + 1-2 件标志性服饰或道具（如绿色盘扣上衣、波点被子）
- 分镜布局：每格仅 1 条顶栏字幕（dialogue 合并为一句 20 字内，或写在 narration）；**禁止** speech bubble、框内文字
- scene 须明确哪种动物、在做什么动作与表情；构图留出画面上方约 20% 给程序叠加顶栏黑字
- 分镜 imagePrompt 必须包含：anthropomorphic animal, flat 2D illustration, bold black outlines, solid colors, minimal background, horizontal 16:9, leave top area clear, no text`

// GetComicStylePrompt 根据漫画风格返回附加 Prompt 片段（空或未知风格返回空字符串）
func GetComicStylePrompt(style string) string {
	switch style {
	case ComicStyleCartoon:
		return StyleCartoonPrompt
	case ComicStyleRealistic:
		return StyleRealisticPrompt
	case ComicStyleChibi:
		return StyleChibiPrompt
	case ComicStyleAnimal:
		return StyleAnimalPrompt
	default:
		return StyleCartoonPrompt // 未知风格默认卡通
	}
}

// GetComicStyleDisplayName 返回风格中文名（日志/展示用）
func GetComicStyleDisplayName(style string) string {
	switch style {
	case ComicStyleCartoon:
		return "卡通"
	case ComicStyleRealistic:
		return "写实"
	case ComicStyleChibi:
		return "Q版"
	case ComicStyleAnimal:
		return "动物漫"
	default:
		return style
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

// BuildPanelImageEnhancePrompt 组装混元生图 Prompt（画面纯绘，顶栏字幕由程序叠加）
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

// BuildDirectPanelImagePrompt 不经 LLM，直接拼装纯画面英文 Prompt（顶栏字幕由程序叠加）
func BuildDirectPanelImagePrompt(style, scene, characters, imagePrompt, dialogue, narration string) string {
	_ = dialogue
	_ = narration
	parts := make([]string, 0, 14)
	switch style {
	case ComicStyleAnimal:
		parts = append(parts,
			"anthropomorphic animal comic",
			"flat 2D illustration",
			"bold black outlines",
			"solid pastel colors",
			"minimal background",
		)
	case "":
		parts = append(parts, "cartoon comic style")
	default:
		parts = append(parts, style+" comic style")
	}
	parts = append(parts,
		"single manga comic panel",
		"horizontal 16:9 aspect ratio",
		"clean line art",
		"vibrant colors",
		"leave top area clear",
		"no text",
		"no letters",
		"no speech bubble",
		"no caption box",
		"no watermark",
	)
	if imagePrompt != "" {
		parts = append(parts, TruncateRunes(SanitizeHunyuanImagePrompt(imagePrompt), 80))
	}
	if scene != "" {
		parts = append(parts, "scene: "+TruncateRunes(scene, 40))
	}
	if characters != "" {
		parts = append(parts, "characters: "+TruncateRunes(characters, 40))
	}
	parts = append(parts, "no deformed hands, high quality illustration")
	return TruncateHunyuanPrompt(strings.Join(parts, ", "))
}

func truncateRunes(s string, max int) string {
	return TruncateRunes(s, max)
}
