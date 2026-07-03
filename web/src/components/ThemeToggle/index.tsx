import type { ReactNode } from 'react'
import { BulbOutlined, StarOutlined } from '@ant-design/icons'
import { Tooltip } from 'antd'
import { useThemeStore, type AppTheme } from '@/stores/theme'
import './index.scss'

const options: { value: AppTheme; label: string; icon: ReactNode; tip: string }[] = [
  {
    value: 'classic',
    label: '经典',
    icon: <BulbOutlined />,
    tip: '浅色经典主题（上一版风格）',
  },
  {
    value: 'immersive',
    label: '沉浸',
    icon: <StarOutlined />,
    tip: '深色沉浸主题（当前最新）',
  },
]

export default function ThemeToggle() {
  const { theme, setTheme } = useThemeStore()

  return (
    <div className="theme-toggle" role="group" aria-label="主题切换">
      {options.map((option) => (
        <Tooltip key={option.value} title={option.tip}>
          <button
            type="button"
            className={`theme-toggle__btn${theme === option.value ? ' theme-toggle__btn--active' : ''}`}
            onClick={() => setTheme(option.value)}
            aria-pressed={theme === option.value}
          >
            {option.icon}
            <span className="theme-toggle__label">{option.label}</span>
          </button>
        </Tooltip>
      ))}
    </div>
  )
}
