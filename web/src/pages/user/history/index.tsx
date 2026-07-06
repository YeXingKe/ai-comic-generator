import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TablePaginationConfig } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { listComicPage } from '@/api/comic'
import type { ComicInfo } from '@/types/api'
import '@/styles/pageShell.css'

function formatTime(time?: string) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function statusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  if (status === 'FAILED') return <Tag color="red">失败</Tag>
  return <Tag>等待中</Tag>
}

export default function HistoryPage() {
  const navigate = useNavigate()
  const [pagination, setPagination] = useState({ pageNum: 1, pageSize: 10 })
  const [total, setTotal] = useState(0)
  const [records, setRecords] = useState<ComicInfo[]>([])
  const [loading, setLoading] = useState(false)

  const fetchList = useCallback(async () => {
    setLoading(true)
    try {
      const res = await listComicPage({
        pageNum: pagination.pageNum,
        pageSize: pagination.pageSize,
      })
      if (res.code === 0 && res.data) {
        setRecords(res.data.records)
        setTotal(res.data.total)
      } else {
        message.error(res.message || '获取创作历史失败')
      }
    } catch {
      message.error('获取创作历史失败')
    } finally {
      setLoading(false)
    }
  }, [pagination.pageNum, pagination.pageSize])

  useEffect(() => {
    void fetchList()
  }, [fetchList])

  const columns = useMemo<ColumnsType<ComicInfo>>(
    () => [
      { title: '任务 ID', dataIndex: 'taskId', width: 120, ellipsis: true },
      { title: '主题', dataIndex: 'topic', width: 160, ellipsis: true, render: (v) => v || '--' },
      {
        title: '标题',
        dataIndex: 'title',
        ellipsis: true,
        render: (title: string | null | undefined, record) =>
          title || record.storyIdeation?.title || record.topic || '未命名作品',
      },
      {
        title: '状态',
        dataIndex: 'status',
        width: 100,
        render: (status: string | undefined) => statusTag(status),
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
        width: 88,
        fixed: 'right',
        render: (_: unknown, record) => (
          <Button type="link" size="small" onClick={() => navigate(`/article/${record.taskId}`)}>
            查看
          </Button>
        ),
      },
    ],
    [navigate],
  )

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
          <p>查看与管理你的全部漫画创作记录</p>
        </header>

        <Table
          rowKey="taskId"
          columns={columns}
          dataSource={records}
          loading={loading}
          pagination={tablePagination}
          scroll={{ x: 960 }}
          className="page-shell__table"
        />
      </div>
    </div>
  )
}
