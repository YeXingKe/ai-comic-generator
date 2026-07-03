import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { TablePaginationConfig } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { mockArticles } from '@/constants/mockArticles'
import type { ArticleVO } from '@/types/api'
import './index.scss'

function formatTime(time?: string) {
  if (!time) return '--'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function statusTag(status?: string) {
  if (status === 'COMPLETED') return <Tag color="purple">已完成</Tag>
  if (status === 'PROCESSING') return <Tag color="blue">生成中</Tag>
  return <Tag>等待中</Tag>
}

export default function HistoryPage() {
  const navigate = useNavigate()
  const [pagination, setPagination] = useState({ pageNum: 1, pageSize: 10 })

  const columns = useMemo<ColumnsType<ArticleVO>>(
    () => [
      { title: '任务 ID', dataIndex: 'taskId', width: 120, ellipsis: true },
      { title: '主题', dataIndex: 'topic', width: 160, ellipsis: true, render: (v) => v || '--' },
      {
        title: '标题',
        dataIndex: 'mainTitle',
        ellipsis: true,
        render: (title: string | undefined, record) => title || record.topic || '未命名作品',
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
    total: mockArticles.length,
    showSizeChanger: true,
    showTotal: (total) => `共 ${total} 条`,
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
          dataSource={mockArticles}
          pagination={tablePagination}
          scroll={{ x: 960 }}
          className="page-shell__table"
        />
      </div>
    </div>
  )
}
