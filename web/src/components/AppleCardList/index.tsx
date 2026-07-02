import type { ReactNode } from 'react'
import { RightOutlined } from '@ant-design/icons'
import './index.scss'

export interface AppleCardItem {
  key: string
  icon: ReactNode
  iconBg?: string
  title: string
  description?: string
  extra?: ReactNode
  onClick?: () => void
  danger?: boolean
}

export interface AppleCardSection {
  title?: string
  footer?: string
  items: AppleCardItem[]
}

interface AppleCardListProps {
  sections: AppleCardSection[]
}

export default function AppleCardList({ sections }: AppleCardListProps) {
  return (
    <div className="apple-card-list">
      {sections.map((section) => (
        <div key={section.title ?? section.items[0]?.key} className="apple-card-list__group">
          {section.title ? (
            <p className="apple-card-list__section-title">{section.title}</p>
          ) : null}
          <div className="apple-card-list__card">
            {section.items.map((item, index) => {
              const Tag = item.onClick ? 'button' : 'div'
              return (
                <Tag
                  key={item.key}
                  type={item.onClick ? 'button' : undefined}
                  className={`apple-card-list__row ${index === section.items.length - 1 ? 'apple-card-list__row--last' : ''} ${item.danger ? 'apple-card-list__row--danger' : ''}`}
                  onClick={item.onClick}
                >
                  <span
                    className="apple-card-list__icon"
                    style={{ background: item.iconBg ?? 'var(--gradient-primary)' }}
                  >
                    {item.icon}
                  </span>
                  <span className="apple-card-list__content">
                    <span className="apple-card-list__title">{item.title}</span>
                    {item.description ? (
                      <span className="apple-card-list__desc">{item.description}</span>
                    ) : null}
                  </span>
                  {item.extra ? (
                    <span className="apple-card-list__extra">{item.extra}</span>
                  ) : null}
                  {item.onClick ? (
                    <RightOutlined className="apple-card-list__chevron" />
                  ) : null}
                </Tag>
              )
            })}
          </div>
          {section.footer ? (
            <p className="apple-card-list__section-footer">{section.footer}</p>
          ) : null}
        </div>
      ))}
    </div>
  )
}
