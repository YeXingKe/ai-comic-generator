import dayjs from 'dayjs'
import type { ArticleVO } from '@/types/api'

export const mockArticles: ArticleVO[] = [
  {
    taskId: 'demo-001',
    topic: 'AI 漫画创作入门',
    mainTitle: '如何用 AI 快速生成精美漫画',
    status: 'COMPLETED',
    createTime: dayjs().subtract(1, 'day').toISOString(),
  },
  {
    taskId: 'demo-002',
    topic: '职场效率提升',
    mainTitle: '10 个提升工作效率的实用技巧',
    status: 'PROCESSING',
    createTime: dayjs().subtract(2, 'hour').toISOString(),
  },
  {
    taskId: 'demo-003',
    topic: '产品思维',
    mainTitle: '从 0 到 1 打造爆款产品的方法论',
    status: 'PENDING',
    createTime: dayjs().subtract(3, 'day').toISOString(),
  },
  {
    taskId: 'demo-004',
    topic: '哪吒闹海',
    mainTitle: '魔童降世：陈塘关的奇幻冒险',
    status: 'COMPLETED',
    createTime: dayjs().subtract(5, 'day').toISOString(),
  },
]
