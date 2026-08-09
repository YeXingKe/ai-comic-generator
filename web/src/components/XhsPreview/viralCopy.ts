/** 小红书 AI 漫画爆款文案模板（前端本地生成，不调接口） */

const TITLE_HOOKS = [
  '谁懂啊！AI一键画出的漫画直接封神',
  '这条AI漫画我能连刷三遍不腻',
  '分镜拉满！AI漫画像在看电影',
  '熬夜赶稿？AI 分镜救我狗命',
  '小红书封神漫：AI 也能讲故事了',
  '收藏！AI 漫画爆款构图公式',
  '同一提示词，AI 画出六格情绪线',
  '看完只想说：这也能是 AI？',
]

const SUBLINES = [
  '从构图到情绪递进，一屏看懂爆款节奏',
  '竖版 2:3 专为小红书信息流优化',
  '角色一致性 + 分镜叙事，完播率拉满',
  '适合职场梗 / 情感线 / 反转剧情',
]

const TAGS = ['#AI漫画', '#小红书爆款', '#分镜设计', '#AI绘画', '#竖版内容', '#漫画创作']

function hashSeed(text: string): number {
  let h = 0
  for (let i = 0; i < text.length; i++) {
    h = (h * 31 + text.charCodeAt(i)) >>> 0
  }
  return h
}

export type ViralCopy = {
  title: string
  subtitle: string
  tags: string[]
  likes: string
  collects: string
  comments: string
  author: string
}

/** 根据提示词生成稳定的爆款文案与互动数据 */
export function buildViralCopy(prompt: string, panelCount: number): ViralCopy {
  const seed = hashSeed(prompt || 'ai-comic')
  const title = TITLE_HOOKS[seed % TITLE_HOOKS.length]
  const subtitle = SUBLINES[seed % SUBLINES.length]
  const tagCount = 3 + (seed % 2)
  const tags = Array.from({ length: tagCount }, (_, i) => TAGS[(seed + i) % TAGS.length])
  const likesNum = 1200 + (seed % 8800)
  const collectsNum = Math.floor(likesNum * (0.28 + (seed % 20) / 100))
  const commentsNum = 36 + (seed % 420)

  const topic = prompt.trim().slice(0, 18)
  const author = topic ? `漫仔_${(seed % 900) + 100}` : 'AI漫研所'

  return {
    title: topic ? `${title}｜${topic}${prompt.trim().length > 18 ? '…' : ''}` : title,
    subtitle: `${subtitle} · ${panelCount} 格连播`,
    tags,
    likes: formatCount(likesNum),
    collects: formatCount(collectsNum),
    comments: formatCount(commentsNum),
    author,
  }
}

function formatCount(n: number): string {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  if (n >= 1000) return `${(n / 1000).toFixed(1)}千`
  return String(n)
}
