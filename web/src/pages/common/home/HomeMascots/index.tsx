import { useEffect, useState } from 'react'
import { ThunderboltOutlined } from '@ant-design/icons'
import { NailongCharacter, NiuniuCharacter } from '../HomeSceneCharacters'
import './index.css'
import '../HomeScence/index.css'

const NAILONG_LINES = ['灵感来了！', '今天画紫色冒险~', '分镜交给我！', '冲冲冲！']
const NIUNIU_LINES = ['脚本我来写~', '配色已就绪 ✦', '一起出爆款！', '等你开画啦']

export default function HomeMascots() {
  const [lineIndex, setLineIndex] = useState(0)

  useEffect(() => {
    const timer = window.setInterval(() => {
      setLineIndex((i) => (i + 1) % NAILONG_LINES.length)
    }, 3200)
    return () => window.clearInterval(timer)
  }, [])

  return (
    <div className="home-mascots" aria-hidden>
      <div className="home-mascots__glow" />

      <article className="home-mascots__item home-mascots__item--left">
        <div className="home-mascots__bubble" key={`n-${lineIndex}`}>
          {NAILONG_LINES[lineIndex]}
        </div>
        <div className="home-mascots__avatar home-mascots__avatar--nailong">
          <NailongCharacter />
        </div>
        <span className="home-mascots__name">奶龙</span>
      </article>

      <div className="home-mascots__center">
        <span className="home-mascots__spark">
          <ThunderboltOutlined />
        </span>
        <span className="home-mascots__tag">创作搭子</span>
      </div>

      <article className="home-mascots__item home-mascots__item--right">
        <div className="home-mascots__bubble" key={`u-${lineIndex}`}>
          {NIUNIU_LINES[lineIndex]}
        </div>
        <div className="home-mascots__avatar home-mascots__avatar--niuniu">
          <NiuniuCharacter />
        </div>
        <span className="home-mascots__name">牛牛</span>
      </article>
    </div>
  )
}
