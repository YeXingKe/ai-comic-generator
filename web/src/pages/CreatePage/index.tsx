import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Input, message } from 'antd'
import { EditOutlined, RocketOutlined, PictureOutlined, OrderedListOutlined } from '@ant-design/icons'
import AppleCardList from '@/components/AppleCardList'
import './index.scss'

export default function CreatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [topic, setTopic] = useState(() => searchParams.get('topic') ?? '')

  const startCreate = () => {
    if (!topic.trim()) {
      message.warning('请输入创作主题')
      return
    }
    message.info('创作流程开发中，敬请期待')
    navigate('/history')
  }

  return (
    <div className="create-page">
      <div className="create-page__inner">
        <header className="create-page__header">
          <h1>开始创作</h1>
          <p>输入主题，AI 将为你生成漫画脚本与分镜</p>
        </header>

        <div className="create-page__composer">
          <Input.TextArea
            value={topic}
            onChange={(e) => setTopic(e.target.value)}
            placeholder="例如：哪吒与敖丙的东海对决，赛博朋克风格四格漫画..."
            autoSize={{ minRows: 4, maxRows: 8 }}
            className="create-page__textarea"
          />
          <Button type="primary" size="large" icon={<RocketOutlined />} onClick={startCreate}>
            生成漫画
          </Button>
        </div>

        <AppleCardList
          sections={[
            {
              title: '创作模式',
              items: [
                {
                  key: 'story',
                  icon: <EditOutlined />,
                  iconBg: 'linear-gradient(135deg, #8b5cf6, #7c3aed)',
                  title: '故事漫画',
                  description: '从主题到完整分镜脚本',
                },
                {
                  key: 'panel',
                  icon: <OrderedListOutlined />,
                  iconBg: 'linear-gradient(135deg, #6366f1, #4f46e5)',
                  title: '分镜模式',
                  description: '按格生成画面与对白',
                },
                {
                  key: 'image',
                  icon: <PictureOutlined />,
                  iconBg: 'linear-gradient(135deg, #ec4899, #db2777)',
                  title: '单图插画',
                  description: '快速生成封面或关键帧',
                },
              ],
            },
          ]}
        />
      </div>
    </div>
  )
}
