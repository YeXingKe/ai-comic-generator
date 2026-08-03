package common

import (
	"strconv"
	"strings"
)

// AI Prompt Templates in English (for GPT models)

// ---------- Comic Style Constants (aligned with model.CreateComicRequest.style) ----------

// Reuse from prompt.go: ComicStyleCartoon, ComicStyleRealistic, ComicStyleChibi, ComicStyleAnimal

// ---------- Step 0: Title Ideation ----------

// TitleIdeationPromptEN - Title recommendation Agent
// Placeholders: {topic} {style} {descriptionSection} {stylePrompt}
const TitleIdeationPromptEN = `You are a creative comic title planner specialized in crafting attention-grabbing titles for social media comic posts.

Generate multiple comic title recommendations based on the following topic:
Topic: {topic}
Comic Style: {style}
{descriptionSection}
{stylePrompt}

Requirements:
1. Generate 4 distinct title options, each closely related to the topic
2. Each title should be no more than 20 words, suitable for 4-6 panel comics, catchy and memorable
3. Subtitle should be 10-25 words explaining the appeal of the title (can be humorous, suspenseful, heartwarming, etc.)
4. Avoid violent, vulgar, or politically sensitive content
5. Four titles should have different angles (e.g., humorous, heartwarming, suspenseful, inspiring)

Return JSON format directly, without markdown code blocks or other explanatory text:
{
  "options": [
    { "title": "Title 1", "subtitle": "Appeal description 1" },
    { "title": "Title 2", "subtitle": "Appeal description 2" },
    { "title": "Title 3", "subtitle": "Appeal description 3" },
    { "title": "Title 4", "subtitle": "Appeal description 4" }
  ]
}`

// ConfirmedTitleSectionEN - Section to insert when user confirms a title
const ConfirmedTitleSectionEN = `

User confirmed comic title: {confirmedTitle}
The title field in the story ideation JSON must use this exact title without modification.`

// ---------- Step 1: Story Ideation ----------

// StoryIdeationPromptEN - Story ideation Agent
// Placeholders: {topic} {style} {descriptionSection} {stylePrompt}
const StoryIdeationPromptEN = `You are an experienced comic scriptwriter skilled at expanding brief topics into complete story plans suitable for 4-6 panel comics.

Complete the story ideation based on the following topic:
Topic: {topic}
Comic Style: {style}
{descriptionSection}
{stylePrompt}

Requirements:
1. Story should be suitable for comic presentation, with clear plot and tight pacing, completable in 4-6 panels or leaving room for sequels
2. Synopsis should summarize the complete storyline in 80-150 words, including introduction, development, climax, and conclusion
3. Theme should extract a core theme (e.g., growth, friendship, comeback, healing)
4. Tone should specify the mood (comedy, action, healing, suspense, heartwarming, etc.), consistent with topic and style
5. Title should provide 1 attractive comic title, no more than 20 words, suitable for social media
6. keyConflict should clearly state the core conflict or challenge the protagonist faces
7. Highlights should list 2-4 highlight moments suitable for comic panels (each no more than 30 words)
8. Avoid violent, vulgar, or politically sensitive content; target general audience

Return JSON format directly, without markdown code blocks or other explanatory text:
{
  "synopsis": "Story synopsis including setup, development, climax, and conclusion",
  "theme": "Core theme",
  "tone": "Mood/tone",
  "title": "Comic title",
  "keyConflict": "Core conflict description",
  "highlights": [
    "Highlight moment 1",
    "Highlight moment 2",
    "Highlight moment 3"
  ]
}`

// StoryDescriptionSectionEN - User additional description (dynamically inserted)
const StoryDescriptionSectionEN = `

User additional requirements: {userDescription}
Please fully incorporate user requirements in the story ideation.`

// ---------- Step 2: Character Design ----------

// CharacterDesignPromptEN - Character design Agent
// Placeholders: {storyIdeation} {style} {stylePrompt}
const CharacterDesignPromptEN = `You are a professional comic character designer skilled at creating highly recognizable characters suitable for AI illustration.

Design comic characters based on the following story ideation:
Story Ideation (JSON):
{storyIdeation}
Comic Style: {style}
{stylePrompt}

Requirements:
1. Design 2-4 characters, must include 1 protagonist (role: protagonist)
2. May include 0-1 antagonist and 1-2 supporting characters
3. Name should be memorable and culturally appropriate
4. Appearance should detail physical features, clothing, color scheme, signature props for consistency across panels (80-120 words)
5. Personality should summarize character traits and speech patterns in 2-3 sentences for dialogue writing
6. avatarUrl should be empty string "" (image generated in later steps)
7. Character appearance must align with comic style (cartoon/chibi/realistic/anthropomorphic animal)
8. Characters should have visual distinction (different hairstyles, outfits, body types)

Return JSON array format directly, without markdown code blocks or other explanatory text:
[
  {
    "name": "Character name",
    "role": "protagonist",
    "appearance": "Detailed appearance and outfit description",
    "personality": "Personality traits and speech patterns",
    "avatarUrl": ""
  },
  {
    "name": "Character name",
    "role": "supporting",
    "appearance": "Detailed appearance and outfit description",
    "personality": "Personality traits and speech patterns",
    "avatarUrl": ""
  }
]`

// ---------- Step 3: Storyboard Script ----------

// StoryboardScriptPromptEN - Storyboard script Agent
// Placeholders: {storyIdeation} {characters} {style} {panelCount} {descriptionSection} {stylePrompt} {captionCompositionHint}
const StoryboardScriptPromptEN = `You are an experienced comic storyboard artist skilled at breaking down stories into 4-6 panel comic scripts suitable for social media.

Create storyboard script based on the following story and characters:
Story Ideation (JSON):
{storyIdeation}
Character Design (JSON array):
{characters}
Comic Style: {style}
Panel Count: {panelCount} panels (output exactly this number of panels)
{descriptionSection}
{stylePrompt}

Requirements:
1. Total {panelCount} panels, panelNo numbered consecutively from 1 to {panelCount}
2. Panel 1 should establish scene or introduce conflict, final panel should deliver punchline, twist, or heartwarming conclusion
3. Scene describes panel content, character actions, expressions, background environment, detailed enough for illustration (50-100 words per panel)
4. dialogue is array of character lines, each no more than 20 words, matching character personality; dialogueEn is corresponding English translation array, format exactly matching dialogue; when no dialogue, both return empty arrays []
5. narration is narrator text, empty string "" if none
6. camera specifies shot type: close-up / medium shot / wide shot / bird's eye / low angle / over-shoulder, etc.
7. imagePrompt must be English keyword-style short phrases (not long paragraphs), total length under 180 characters; include: character appearance keywords, actions, scene, lighting, style keywords; must specify horizontal 16:9 cinematic widescreen comic panel
8. imagePrompt describes only visual elements, **MUST NOT** include any text/dialogue/speech bubbles; {captionCompositionHint}
9. Character appearance and dialogue must match character design, do not add undesigned characters
10. pageCount fill 1 (single page 4-6 panels) or actual page count

Return JSON format directly, without markdown code blocks or other explanatory text:
{
  "pageCount": 1,
  "panels": [
    {
      "panelNo": 1,
      "scene": "Panel visual and action description",
      "dialogue": ["Character A: line"],
      "dialogueEn": ["Character A: line"],
      "narration": "",
      "camera": "medium shot",
      "imagePrompt": "English prompt for AI image generation, cartoon style, ..."
    },
    {
      "panelNo": 2,
      "scene": "Panel visual and action description",
      "dialogue": [],
      "dialogueEn": [],
      "narration": "Narrator text",
      "camera": "close-up",
      "imagePrompt": "English prompt..."
    }
  ]
}`

// AiModifyStoryboardPromptEN - AI modify storyboard (optional when user manually edits)
// Placeholders: {storyIdeation} {currentStoryboard} {modifySuggestion}
const AiModifyStoryboardPromptEN = `You are a professional comic storyboard artist skilled at optimizing storyboard scripts based on user feedback.

Current story ideation:
{storyIdeation}

Current storyboard script:
{currentStoryboard}

User modification suggestions:
{modifySuggestion}

Requirements:
1. Adjust storyboard based on user suggestions, maintain panelNo numbered consecutively from 1
2. Keep JSON structure and field names unchanged (panelNo, scene, dialogue, narration, camera, imagePrompt)
3. If user requests to add/remove panels, adjust panels array length and pageCount accordingly
4. imagePrompt must remain in English, suitable for AI image generation
5. Modified plot must still align with story ideation

Return complete modified storyboard JSON directly, without markdown code blocks or other explanatory text:
{
  "pageCount": 1,
  "panels": [
    {
      "panelNo": 1,
      "scene": "Panel visual and action description",
      "dialogue": ["line"],
      "narration": "",
      "camera": "medium shot",
      "imagePrompt": "English prompt..."
    }
  ]
}`

// ---------- Step 4: Image Generation (Prompt Enhancement, used by image_service) ----------

// PanelImageEnhancePromptEN - Enhance image prompt based on storyboard imagePrompt
// Placeholders: {style} {stylePrompt} {imagePrompt} {scene} {characters} {dialogue} {narration} {captionCompositionHint}
const PanelImageEnhancePromptEN = `### Task ###
You are a professional comic panel illustrator responsible for converting storyboard information into an English prompt for AI painting models.
**Important: Image generation models do not render text. Prompt must NOT include any dialogue text or require speech bubbles.**
Dialogue will be overlaid by subsequent programs based on caption mode. You only need to describe pure visuals (characters, actions, scene, composition).

### Input ###
Comic Style: {style}
Panel Scene Description: {scene}
Character Reference: {characters}
Panel Dialogue (for understanding emotion only, do NOT write into prompt): {dialogue}
Panel Narration (for understanding only, do NOT write into prompt): {narration}
Original Prompt: {imagePrompt}
{stylePrompt}

### Visual Requirements ###
1. single comic panel, horizontal 16:9 cinematic widescreen frame
2. Character expressions and actions match dialogue emotion, but **image must NOT contain any text, letters, numbers, watermarks, or gibberish**
3. {captionCompositionHint}
4. clean line art, style keywords, character consistency, scene, lighting

### Output Requirements ###
1. Output one English keyword prompt, no JSON, no markdown
2. **Total length must not exceed 200 characters**
3. Must include: no text, no letters, no watermark

### Output ###
Return final English prompt directly, no other content.`

// ---------- Step 5: Layout Composition (Caption assistance, optional LLM-generated image caption) ----------

// LayoutCaptionPromptEN - Generate brief image caption/intro for composed comic
// Placeholders: {title} {synopsis} {panelCount}
const LayoutCaptionPromptEN = `You are a social media comic editor skilled at writing engaging comic intro copy.

Comic Title: {title}
Story Synopsis: {synopsis}
Panel Count: {panelCount}

Requirements:
1. Output 1 intro paragraph, 50-80 words, conversational with suspense or humor, suitable for article opening
2. Do not use clichés like "this article" or "the editor"
3. Do not spoil the ending

Return plain text intro directly, no title, JSON or other formatting.`

// ---------- Step 6: WeChat Publication (Article summary) ----------

// WechatPublishCopyPromptEN - Generate title and summary for WeChat publication
// Placeholders: {title} {synopsis} {theme} {tone}
const WechatPublishCopyPromptEN = `You are a WeChat official account operations expert skilled at writing viral titles and summaries for comic content.

Comic Title: {title}
Story Synopsis: {synopsis}
Theme: {theme}
Tone: {tone}

Requirements:
1. Output JSON including title (publication title, no more than 32 words), digest (summary, no more than 80 words), tags (3-5 tags, string array)
2. Title should be attractive, may use questions, contrasts, numbers appropriately, but avoid clickbait
3. Digest summarizes story highlights, guides readers to open full comic, without spoiling ending
4. Tags should relate to comic theme for categorization

Return JSON format directly, without markdown code blocks or other explanatory text:
{
  "title": "Publication article title",
  "digest": "Article summary",
  "tags": ["tag1", "tag2", "tag3"]
}`

// ---------- Comic Style Additional Prompts (runtime spliced by style field) ----------

const StyleCartoonPromptEN = `

**Important: Please use cartoon comic style for creation and description**
- Rounded shapes, clean lines, bright colors
- Moderately exaggerated expressions, suitable for 4-panel comedy or heartwarming
- Background can be simplified, highlighting characters and actions
- Storyboard imagePrompt must include: cartoon style, clean line art, vibrant colors, leave top area clear, no text`

const StyleRealisticPromptEN = `

**Important: Please use realistic comic style for creation and description**
- Character proportions close to reality, rich details (clothing textures, lighting)
- Accurate scene perspective, strong atmosphere, cinematic panel quality
- Restrained dialogue, suitable for drama, suspense, or urban themes
- Storyboard imagePrompt must include: realistic comic style, detailed shading, cinematic lighting, leave top area clear, no text`

const StyleChibiPromptEN = `

**Important: Please use Chibi (super deformed) style for creation and description**
- 2-3 head-to-body ratio, large head small body, cute aesthetic
- Extremely exaggerated expressions, suitable for moe, comedy, light healing
- Props and scenes can be cutely deformed, pastel or highly saturated colors
- Storyboard imagePrompt must include: chibi style, super deformed, cute, kawaii, 2-3 head ratio, leave top area clear, no text`

const StyleAnimalPromptEN = `

**Important: Please use "anthropomorphic animal comic strip" style for creation and description (reference social media healing/humorous animal comics like anthropomorphic frog family daily life)**
- All characters are anthropomorphic animals (frogs, cats, rabbits, bears, foxes, dogs, etc.), standing upright with hands, feet, and humanized expressions; **NO human characters allowed**
- Visual style: bold black outlines, flat 2D coloring, low saturation or bright solid colors; minimal background (e.g., single-color wall + ground line, simple furniture)
- Scenes lean toward daily life: bedroom, living room, office, grass, suitable for 4-panel heartwarming/light humor/family interaction
- Character appearance must specify: species + skin/fur color + body type + 1-2 signature clothing items or props (e.g., green frog-closure top, polka dot blanket)
- Panel layout: each panel has only 1 top caption bar (merge dialogue into one sentence under 20 words, or write in narration); **PROHIBIT** speech bubbles or in-frame text
- Scene must clearly specify which animal, doing what action and expression; composition leaves top ~20% for program to overlay top caption bar with black text
- Storyboard imagePrompt must include: anthropomorphic animal, flat 2D illustration, bold black outlines, solid colors, minimal background, horizontal 16:9, leave top area clear, no text`

// GetComicStylePromptEN - Return additional prompt segment based on comic style (empty/unknown style returns empty string)
func GetComicStylePromptEN(style string) string {
	switch style {
	case ComicStyleCartoon:
		return StyleCartoonPromptEN
	case ComicStyleRealistic:
		return StyleRealisticPromptEN
	case ComicStyleChibi:
		return StyleChibiPromptEN
	case ComicStyleAnimal:
		return StyleAnimalPromptEN
	default:
		return StyleCartoonPromptEN // Unknown style defaults to cartoon
	}
}

// BuildDescriptionSectionEN - Build insertable prompt segment from user additional description
func BuildDescriptionSectionEN(userDescription string) string {
	if userDescription == "" {
		return ""
	}
	return strings.ReplaceAll(StoryDescriptionSectionEN, "{userDescription}", userDescription)
}

// BuildTitleIdeationPromptEN - Assemble complete title ideation prompt
func BuildTitleIdeationPromptEN(topic, style, userDescription string) string {
	prompt := strings.ReplaceAll(TitleIdeationPromptEN, "{topic}", topic)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", BuildDescriptionSectionEN(userDescription))
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePromptEN(style))
	return prompt
}

// BuildStoryIdeationPromptEN - Assemble complete story ideation prompt
func BuildStoryIdeationPromptEN(topic, style, userDescription, confirmedTitle string) string {
	prompt := strings.ReplaceAll(StoryIdeationPromptEN, "{topic}", topic)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	descSection := BuildDescriptionSectionEN(userDescription)
	if confirmedTitle != "" {
		descSection += strings.ReplaceAll(ConfirmedTitleSectionEN, "{confirmedTitle}", confirmedTitle)
	}
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", descSection)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePromptEN(style))
	return prompt
}

// BuildCharacterDesignPromptEN - Assemble complete character design prompt
func BuildCharacterDesignPromptEN(storyIdeationJSON, style string) string {
	prompt := strings.ReplaceAll(CharacterDesignPromptEN, "{storyIdeation}", storyIdeationJSON)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePromptEN(style))
	return prompt
}

// BuildStoryboardScriptPromptEN - Assemble complete storyboard script prompt
func BuildStoryboardScriptPromptEN(storyIdeationJSON, charactersJSON, style, userDescription, captionTextMode string, panelCount int) string {
	if panelCount <= 0 {
		panelCount = 4
	}
	countStr := strconv.Itoa(panelCount)
	prompt := strings.ReplaceAll(StoryboardScriptPromptEN, "{storyIdeation}", storyIdeationJSON)
	prompt = strings.ReplaceAll(prompt, "{characters}", charactersJSON)
	prompt = strings.ReplaceAll(prompt, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{panelCount}", countStr)
	prompt = strings.ReplaceAll(prompt, "{descriptionSection}", BuildDescriptionSectionEN(userDescription))
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePromptEN(style))
	prompt = strings.ReplaceAll(prompt, "{captionCompositionHint}", getCaptionCompositionHintEN(captionTextMode))
	return prompt
}

// BuildPanelImageEnhancePromptEN - Assemble image generation prompt (pure visual, caption mode determined by captionTextMode)
func BuildPanelImageEnhancePromptEN(style, scene, characters, imagePrompt, dialogue, narration, captionTextMode string) string {
	if dialogue == "" {
		dialogue = "(none)"
	}
	if narration == "" {
		narration = "(none)"
	}
	prompt := strings.ReplaceAll(PanelImageEnhancePromptEN, "{style}", style)
	prompt = strings.ReplaceAll(prompt, "{scene}", scene)
	prompt = strings.ReplaceAll(prompt, "{characters}", characters)
	prompt = strings.ReplaceAll(prompt, "{imagePrompt}", imagePrompt)
	prompt = strings.ReplaceAll(prompt, "{dialogue}", dialogue)
	prompt = strings.ReplaceAll(prompt, "{narration}", narration)
	prompt = strings.ReplaceAll(prompt, "{stylePrompt}", GetComicStylePromptEN(style))
	prompt = strings.ReplaceAll(prompt, "{captionCompositionHint}", getCaptionCompositionHintEN(captionTextMode))
	return prompt
}

// getCaptionCompositionHintEN - Return image composition requirements based on caption mode
func getCaptionCompositionHintEN(captionTextMode string) string {
	switch captionTextMode {
	case captionTextModeTop:
		return "Leave top ~20% of frame clear for overlay caption bar (leave top area clear, no text in top 20%)"
	case captionTextModeBubble:
		return "Normal composition, leave space near characters for speech bubble overlay (normal composition, leave space near characters for speech bubble overlay)"
	default: // none
		return "Full frame composition, no space needed for captions (full frame composition, no caption overlay needed)"
	}
}
