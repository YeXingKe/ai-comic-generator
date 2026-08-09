import { useEffect, useMemo, useState } from 'react'
import { Modal } from 'antd'
import {
  HeartFilled,
  LeftOutlined,
  MessageFilled,
  RightOutlined,
  StarFilled,
  FireFilled,
} from '@ant-design/icons'
import { buildViralCopy } from './viralCopy'
import './index.css'

export type XhsPreviewImage = {
  url: string
  panelNo: number
}

type XhsPhonePreviewProps = {
  open: boolean
  onClose: () => void
  images: XhsPreviewImage[]
  prompt?: string
  initialIndex?: number
}

function pad2(n: number) {
  return n < 10 ? `0${n}` : String(n)
}

export function XhsPhonePreview({ open, onClose, images, prompt = '', initialIndex = 0 }: XhsPhonePreviewProps) {
  const [index, setIndex] = useState(0)
  const [liked, setLiked] = useState(false)

  const copy = useMemo(() => buildViralCopy(prompt, images.length || 1), [prompt, images.length])
  const current = images[index]
  const total = images.length
  const now = useMemo(() => {
    const d = new Date()
    return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
  }, [open])

  useEffect(() => {
    if (!open) return
    const safe = Math.min(Math.max(initialIndex, 0), Math.max(images.length - 1, 0))
    setIndex(safe)
    setLiked(false)
  }, [open, initialIndex, images.length])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') setIndex((i) => Math.max(0, i - 1))
      if (e.key === 'ArrowRight') setIndex((i) => Math.min(total - 1, i + 1))
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, total])

  const goPrev = () => setIndex((i) => Math.max(0, i - 1))
  const goNext = () => setIndex((i) => Math.min(total - 1, i + 1))

  return (
    <Modal open={open} onCancel={onClose} footer={null} centered width={400} className="xhs-preview-modal" destroyOnClose>
      <div className="xhs-phone">
        <div className="xhs-phone__notch" />
        <div className="xhs-phone__screen">
          <div className="xhs-phone__status">
            <span>{now}</span>
            <span>5G ▮▮▮</span>
          </div>

          <div className="xhs-feed">
            <div className="xhs-carousel">
              {current ? (
                <img className="xhs-carousel__img" src={current.url} alt={`分镜 ${current.panelNo}`} />
              ) : (
                <div className="xhs-empty">暂无分镜，先生成后再预览小红书排版</div>
              )}

              {total > 1 && (
                <>
                  <button type="button" className="xhs-carousel__nav xhs-carousel__nav--prev" onClick={goPrev} disabled={index <= 0} aria-label="上一张">
                    <LeftOutlined />
                  </button>
                  <button type="button" className="xhs-carousel__nav xhs-carousel__nav--next" onClick={goNext} disabled={index >= total - 1} aria-label="下一张">
                    <RightOutlined />
                  </button>
                </>
              )}

              {total > 0 && (
                <div className="xhs-carousel__counter">
                  {index + 1}/{total}
                </div>
              )}

              {total > 1 && (
                <div className="xhs-carousel__dots">
                  {images.map((img, i) => (
                    <span key={img.panelNo} className={`xhs-carousel__dot${i === index ? ' is-active' : ''}`} />
                  ))}
                </div>
              )}

              <div className="xhs-side">
                <div className="xhs-side__avatar-wrap">
                  <div className="xhs-side__avatar">{copy.author.slice(0, 1)}</div>
                  <span className="xhs-side__follow">+</span>
                </div>
                <button type="button" className="xhs-side__action" onClick={() => setLiked((v) => !v)}>
                  <HeartFilled className={`xhs-side__action-icon${liked ? ' is-liked' : ''}`} />
                  <span>{liked ? bumpCount(copy.likes) : copy.likes}</span>
                </button>
                <div className="xhs-side__action">
                  <StarFilled className="xhs-side__action-icon" />
                  <span>{copy.collects}</span>
                </div>
                <div className="xhs-side__action">
                  <MessageFilled className="xhs-side__action-icon" />
                  <span>{copy.comments}</span>
                </div>
              </div>
            </div>

            <div className="xhs-caption">
              <div className="xhs-caption__badge">
                <FireFilled /> AI 漫画爆款模板
              </div>
              <h3 className="xhs-caption__title">{copy.title}</h3>
              <p className="xhs-caption__sub">{copy.subtitle}</p>
              <div className="xhs-caption__tags">
                {copy.tags.map((tag) => (
                  <span key={tag} className="xhs-caption__tag">
                    {tag}
                  </span>
                ))}
              </div>
              <div className="xhs-caption__author">
                <div className="xhs-caption__author-left">
                  <div className="xhs-caption__author-avatar">{copy.author.slice(0, 1)}</div>
                  <span className="xhs-caption__author-name">{copy.author}</span>
                </div>
                <span className="xhs-caption__cta">关注</span>
              </div>
            </div>
          </div>
        </div>
        <div className="xhs-phone__home" />
      </div>
    </Modal>
  )
}

function bumpCount(display: string): string {
  if (display.includes('万')) return display
  if (display.includes('千')) return display
  const n = Number(display)
  if (Number.isFinite(n)) return String(n + 1)
  return display
}

export { buildViralCopy } from './viralCopy'
export default XhsPhonePreview
