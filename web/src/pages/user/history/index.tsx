import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TablePaginationConfig, ColumnsType } from 'antd/es/table'
import { Table, Tag, Button, Space, Modal, Tabs, message } from 'antd'
import dayjs from 'dayjs'
import { listComicPage, listCustomComicPage, publishComic } from '@/api/comic'
import type { ComicInfo, CustomComicInfo } from '@/types/api'
import { ADMIN_ROLE, useLoginUserStore } from '@/stores/loginUser'
import '@/styles/pageShell.css'

type HistoryMode = 'auto' | 'custom'

function formatTime(time?: string) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function autoStatusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  if (status === 'AWAITING_CONFIRM') return <Tag color="gold">待确认标题</Tag>
  if (status === 'TITLE_CONFIRMED') return <Tag color="cyan">待开始</Tag>
  if (status === 'AWAITING_STORYBOARD') return <Tag color="orange">待确认分镜</Tag>
  if (status === 'FAILED') return <Tag color="red">失败</Tag>
  return <Tag>等待中</Tag>
}

function customStatusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  if (status === 'FAILED') return <Tag color="red">失败</Tag>
  return <Tag>等待中</Tag>
}

function publishStatusTag(record: ComicInfo) {
  const status = record.publishResult?.status
  if (status === 'PUBLISHED') return <Tag color="green">已发布</Tag>
  if (status === 'DRAFT') return <Tag color="default">草稿</Tag>
  if (status === 'FAILED') return <Tag color="red">发布失败</Tag>
  return <Tag>未发布</Tag>
}

function creatorLabel(record: { userName?: string | null; userAccount?: string; userId: number }) {
  return record.userName || record.userAccount || String(record.userId) || '--'
}

const IMAGE_BACKEND_LABEL: Record<string, string> = {
  hunyuan: '混元',
  openai_image_1k: 'OpenAI 1K',
  openai_image_4k: 'OpenAI 4K',
}

export default function HistoryPage() {
  const navigate = useNavigate()
  const loginUser = useLoginUserStore((s) => s.loginUser)
  const isAdmin = loginUser.userRole === ADMIN_ROLE
  const [mode, setMode] = useState<HistoryMode>('auto')
  const [pagination, setPagination] = useState({ pageNum: 1, pageSize: 10 })
  const [total, setTotal] = useState(0)
  const [autoRecords, setAutoRecords] = useState<ComicInfo[]>([])
  const [customRecords, setCustomRecords] = useState<CustomComicInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [publishingTaskId, setPublishingTaskId] = useState<string | null>(null)

  const fetchList = useCallback(async () => {
    setLoading(true)
    try {
      if (mode === 'auto') {
        const res = await listComicPage({
          pageNum: pagination.pageNum,
          pageSize: pagination.pageSize,
        })
        if (res.code === 0 && res.data) {
          setAutoRecords(res.data.records)
          setTotal(res.data.total)
        } else {
          message.error(res.message || '获取自动化创作历史失败')
        }
      } else {
        const res = await listCustomComicPage({
          pageNum: pagination.pageNum,
          pageSize: pagination.pageSize,
        })
        if (res.code === 0 && res.data) {
          setCustomRecords(res.data.records)
          setTotal(res.data.total)
        } else {
          message.error(res.message || '获取自定义创作历史失败')
        }
      }
    } catch {
      message.error('获取创作历史失败')
    } finally {
      setLoading(false)
    }
  }, [mode, pagination.pageNum, pagination.pageSize])

  useEffect(() => {
    void fetchList()
  }, [fetchList])

  const handleModeChange = (key: string) => {
    setMode(key as HistoryMode)
    setPagination({ pageNum: 1, pageSize: pagination.pageSize })
    setTotal(0)
  }

  const handlePublish = useCallback(
    (record: ComicInfo) => {
      const title = record.title || record.storyIdeation?.title || record.topic || '未命名作品'
      const republish = record.publishResult?.status === 'PUBLISHED'
      Modal.confirm({
        title: republish ? '重新发布到公众号？' : '发布到公众号？',
        content: `确认将《${title}》上传至微信公众号素材库。`,
        okText: republish ? '重新发布' : '发布',
        cancelText: '取消',
        onOk: async () => {
          setPublishingTaskId(record.taskId)
          try {
            const res = await publishComic({ taskId: record.taskId })
            if (res.code !== 0) {
              message.error(res.message || '发布失败')
              return
            }
            const status = res.data?.status
            if (status === 'PUBLISHED') {
              message.success('发布成功')
            } else if (status === 'DRAFT') {
              message.success('已标记为草稿（公众号未启用）')
            } else {
              message.warning(res.message || '发布未成功')
            }
            await fetchList()
          } catch {
            message.error('发布失败')
          } finally {
            setPublishingTaskId(null)
          }
        },
      })
    },
    [fetchList],
  )

  const autoColumns = useMemo<ColumnsType<ComicInfo>>(() => {
    const cols: ColumnsType<ComicInfo> = [
      { title: '任务 ID', dataIndex: 'taskId', width: 120, ellipsis: true },
      { title: '主题', dataIndex: 'topic', width: 160, ellipsis: true, render: (v) => v || '--' },
      {
        title: '标题',
        dataIndex: 'title',
        ellipsis: true,
        render: (title: string | null | undefined, record) => title || record.storyIdeation?.title || record.topic || '未命名作品',
      },
    ]
    if (isAdmin) {
      cols.push({
        title: '创建者',
        key: 'creator',
        width: 140,
        ellipsis: true,
        render: (_: unknown, record) => creatorLabel(record),
      })
    }
    cols.push(
      {
        title: '状态',
        dataIndex: 'status',
        width: 100,
        render: (status: string | undefined) => autoStatusTag(status),
      },
      {
        title: '发布',
        key: 'publishStatus',
        width: 100,
        render: (_: unknown, record) => publishStatusTag(record),
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        width: 180,
        render: (time: string | undefined) => formatTime(time),
      },
      {
        title: '操作',
        key: 'action',
        width: 160,
        fixed: 'right',
        render: (_: unknown, record) => {
          const canPublish = record.status === 'COMPLETED'
          const published = record.publishResult?.status === 'PUBLISHED'
          const canContinue =
            record.status === 'AWAITING_CONFIRM' ||
            record.status === 'TITLE_CONFIRMED' ||
            record.status === 'AWAITING_STORYBOARD' ||
            record.status === 'FAILED' ||
            record.status === 'PROCESSING' ||
            record.status === 'PENDING'
          return (
            <Space size={0}>
              {canContinue ? (
                <Button type="link" size="small" onClick={() => navigate(`/create?taskId=${encodeURIComponent(record.taskId)}`)}>
                  {record.status === 'FAILED' ? '继续编辑' : '继续创作'}
                </Button>
              ) : (
                <>
                  <Button type="link" size="small" onClick={() => navigate(`/comic/${record.taskId}`)}>
                    查看
                  </Button>
                  <Button type="link" size="small" onClick={() => navigate(`/create?taskId=${encodeURIComponent(record.taskId)}`)}>
                    编辑
                  </Button>
                </>
              )}
              {canPublish && (
                <Button type="link" size="small" loading={publishingTaskId === record.taskId} onClick={() => handlePublish(record)}>
                  {published ? '重新发布' : '发布'}
                </Button>
              )}
            </Space>
          )
        },
      },
    )
    return cols
  }, [navigate, handlePublish, publishingTaskId, isAdmin])

  const customColumns = useMemo<ColumnsType<CustomComicInfo>>(() => {
    const cols: ColumnsType<CustomComicInfo> = [
      { title: '任务 ID', dataIndex: 'taskId', width: 120, ellipsis: true },
      {
        title: '提示词',
        dataIndex: 'prompt',
        ellipsis: true,
        render: (v: string) => v || '--',
      },
      {
        title: '画幅',
        dataIndex: 'aspectRatio',
        width: 80,
      },
      {
        title: '模型',
        dataIndex: 'imageBackend',
        width: 110,
        render: (v: string) => IMAGE_BACKEND_LABEL[v] || v || '--',
      },
      {
        title: '格数',
        dataIndex: 'panelCount',
        width: 70,
      },
    ]
    if (isAdmin) {
      cols.push({
        title: '创建者',
        key: 'creator',
        width: 140,
        ellipsis: true,
        render: (_: unknown, record) => creatorLabel(record),
      })
    }
    cols.push(
      {
        title: '状态',
        dataIndex: 'status',
        width: 100,
        render: (status: string | undefined) => customStatusTag(status),
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        width: 180,
        render: (time: string | undefined) => formatTime(time),
      },
      {
        title: '操作',
        key: 'action',
        width: 100,
        fixed: 'right',
        render: (_: unknown, record) => (
          <Button type="link" size="small" onClick={() => navigate(`/create/custom?taskId=${encodeURIComponent(record.taskId)}`)}>
            自定义创作
          </Button>
        ),
      },
    )
    return cols
  }, [navigate, isAdmin])

  const tablePagination: TablePaginationConfig = {
    current: pagination.pageNum,
    pageSize: pagination.pageSize,
    total,
    showSizeChanger: true,
    showTotal: (n) => `共 ${n} 条`,
    onChange: (page, pageSize) => setPagination({ pageNum: page, pageSize }),
  }

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <h1>创作历史</h1>
          <p>
            {isAdmin
              ? '按创作模式查看全部用户记录；自动化作品完成后可发布至公众号'
              : '按创作模式查看你的记录；自动化作品完成后可发布至公众号'}
          </p>
        </header>

        <Tabs
          activeKey={mode}
          onChange={handleModeChange}
          items={[
            { key: 'auto', label: '自动化创作' },
            { key: 'custom', label: '自定义创作' },
          ]}
          style={{ marginBottom: 8 }}
        />

        {mode === 'auto' ? (
          <Table
            rowKey="taskId"
            columns={autoColumns}
            dataSource={autoRecords}
            loading={loading}
            pagination={tablePagination}
            scroll={{ x: isAdmin ? 1200 : 1060 }}
            className="page-shell__table"
          />
        ) : (
          <Table
            rowKey="taskId"
            columns={customColumns}
            dataSource={customRecords}
            loading={loading}
            pagination={tablePagination}
            scroll={{ x: isAdmin ? 1100 : 960 }}
            className="page-shell__table"
          />
        )}
      </div>
    </div>
  )
}
