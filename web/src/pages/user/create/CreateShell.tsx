import { useNavigate } from 'react-router-dom'
import { Tabs } from 'antd'
import { EditOutlined, PictureOutlined, ThunderboltOutlined } from '@ant-design/icons'
import './CreateShell.css'

export type CreateMode = 'auto' | 'custom'

const MODE_SUBTITLE: Record<CreateMode, string> = {
  auto: 'AI 智能体流水线：从主题与标题推荐，到分镜脚本、成稿与排版合成，一键生成完整漫画',
  custom: '自由配置画幅、模型与提示词，一次生成多格分镜，支持小红书排版与打包下载',
}

type CreateShellProps = {
  mode: CreateMode
  children: React.ReactNode
}

export default function CreateShell({ mode, children }: CreateShellProps) {
  const navigate = useNavigate()

  const handleModeChange = (key: string) => {
    navigate(key === 'auto' ? '/create' : '/create/custom')
  }

  return (
    <div className="create-shell">
      <div className="create-shell__top">
        <header className="create-shell__header">
          <h1>
            🎨 漫画工坊
          </h1>
          <p>{MODE_SUBTITLE[mode]}</p>
        </header>

        <Tabs
          centered
          activeKey={mode}
          onChange={handleModeChange}
          items={[
            { key: 'auto', label: '自动化创作' },
            { key: 'custom', label: '自定义创作' },
          ]}
          className="create-shell__tabs"
        />
      </div>

      <div className="create-shell__body">{children}</div>
    </div>
  )
}
