package common

import (
	"strconv"
	"strings"
)

// PromptBuilder 根据语言配置选择对应的 Prompt 构建函数
type PromptBuilder struct {
	lang string // "zh" 或 "en"
}

// NewPromptBuilder 创建 Prompt 构建器
func NewPromptBuilder(lang string) *PromptBuilder {
	if lang != "zh" && lang != "en" {
		lang = "zh" // 默认中文
	}
	return &PromptBuilder{lang: lang}
}

// BuildTitleIdeation 构建标题推荐 Prompt
func (pb *PromptBuilder) BuildTitleIdeation(topic, style, userDescription string) string {
	if pb.lang == "en" {
		return BuildTitleIdeationPromptEN(topic, style, userDescription)
	}
	return BuildTitleIdeationPrompt(topic, style, userDescription)
}

// BuildStoryIdeation 构建故事构思 Prompt
func (pb *PromptBuilder) BuildStoryIdeation(topic, style, userDescription, confirmedTitle string) string {
	if pb.lang == "en" {
		return BuildStoryIdeationPromptEN(topic, style, userDescription, confirmedTitle)
	}
	return BuildStoryIdeationPrompt(topic, style, userDescription, confirmedTitle)
}

// BuildCharacterDesign 构建角色设定 Prompt
func (pb *PromptBuilder) BuildCharacterDesign(storyIdeationJSON, style string) string {
	if pb.lang == "en" {
		return BuildCharacterDesignPromptEN(storyIdeationJSON, style)
	}
	return BuildCharacterDesignPrompt(storyIdeationJSON, style)
}

// BuildStoryboardScript 构建分镜脚本 Prompt
func (pb *PromptBuilder) BuildStoryboardScript(storyIdeationJSON, charactersJSON, style, userDescription, captionTextMode string, panelCount int) string {
	if pb.lang == "en" {
		return BuildStoryboardScriptPromptEN(storyIdeationJSON, charactersJSON, style, userDescription, captionTextMode, panelCount)
	}
	return BuildStoryboardScriptPrompt(storyIdeationJSON, charactersJSON, style, userDescription, captionTextMode, panelCount)
}

// BuildPanelImageEnhance 构建生图增强 Prompt
func (pb *PromptBuilder) BuildPanelImageEnhance(style, scene, characters, imagePrompt, dialogue, narration, captionTextMode string) string {
	if pb.lang == "en" {
		return BuildPanelImageEnhancePromptEN(style, scene, characters, imagePrompt, dialogue, narration, captionTextMode)
	}
	return BuildPanelImageEnhancePrompt(style, scene, characters, imagePrompt, dialogue, narration, captionTextMode)
}

// BuildLayoutCaption 构建排版图注 Prompt
func (pb *PromptBuilder) BuildLayoutCaption(title, synopsis string, panelCount int) string {
	if pb.lang == "en" {
		return BuildLayoutCaptionPromptEN(title, synopsis, panelCount)
	}
	return BuildLayoutCaptionPrompt(title, synopsis, panelCount)
}

// BuildWechatPublishCopy 构建公众号发布文案 Prompt
func (pb *PromptBuilder) BuildWechatPublishCopy(title, synopsis, theme, tone string) string {
	if pb.lang == "en" {
		return BuildWechatPublishCopyPromptEN(title, synopsis, theme, tone)
	}
	return BuildWechatPublishCopyPrompt(title, synopsis, theme, tone)
}

// GetComicStylePrompt 获取风格附加 Prompt
func (pb *PromptBuilder) GetComicStylePrompt(style string) string {
	if pb.lang == "en" {
		return GetComicStylePromptEN(style)
	}
	return GetComicStylePrompt(style)
}

// BuildLayoutCaptionPrompt 为排版后的漫画生成简短图注/导读（公众号图文导语）
// 占位符：{title} {synopsis} {panelCount}
func BuildLayoutCaptionPrompt(title, synopsis string, panelCount int) string {
	prompt := strings.ReplaceAll(LayoutCaptionPrompt, "{title}", title)
	prompt = strings.ReplaceAll(prompt, "{synopsis}", synopsis)
	prompt = strings.ReplaceAll(prompt, "{panelCount}", strconv.Itoa(panelCount))
	return prompt
}

// BuildLayoutCaptionPromptEN - Assemble complete layout caption prompt
func BuildLayoutCaptionPromptEN(title, synopsis string, panelCount int) string {
	prompt := strings.ReplaceAll(LayoutCaptionPromptEN, "{title}", title)
	prompt = strings.ReplaceAll(prompt, "{synopsis}", synopsis)
	prompt = strings.ReplaceAll(prompt, "{panelCount}", strconv.Itoa(panelCount))
	return prompt
}

// BuildWechatPublishCopyPrompt 生成公众号群发所需的标题与摘要
// 占位符：{title} {synopsis} {theme} {tone}
func BuildWechatPublishCopyPrompt(title, synopsis, theme, tone string) string {
	prompt := strings.ReplaceAll(WechatPublishCopyPrompt, "{title}", title)
	prompt = strings.ReplaceAll(prompt, "{synopsis}", synopsis)
	prompt = strings.ReplaceAll(prompt, "{theme}", theme)
	prompt = strings.ReplaceAll(prompt, "{tone}", tone)
	return prompt
}

// BuildWechatPublishCopyPromptEN - Assemble complete WeChat publish copy prompt
func BuildWechatPublishCopyPromptEN(title, synopsis, theme, tone string) string {
	prompt := strings.ReplaceAll(WechatPublishCopyPromptEN, "{title}", title)
	prompt = strings.ReplaceAll(prompt, "{synopsis}", synopsis)
	prompt = strings.ReplaceAll(prompt, "{theme}", theme)
	prompt = strings.ReplaceAll(prompt, "{tone}", tone)
	return prompt
}
