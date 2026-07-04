import { useEffect, useMemo } from 'react'
import { useLoginUserStore } from '@/stores/loginUser'
import { buildProfileTableColumns } from './profileTableColumns'
import '@/styles/pageShell.css'

export default function UserCenterPage() {
  const { loginUser, fetchLoginUser } = useLoginUserStore()

  useEffect(() => {
    void fetchLoginUser()
  }, [fetchLoginUser])

  const columns = useMemo(() => buildProfileTableColumns(), [])

  return (
    <div className="page-shell">
      <div className="page-shell__inner">
        <header className="page-shell__header">
          <h1>用户中心</h1>
          <p>查看当前账号信息</p>
        </header>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={[loginUser]}
          pagination={false}
          scroll={{ x: 960 }}
          className="page-shell__table"
        />
      </div>
    </div>
  )
}
