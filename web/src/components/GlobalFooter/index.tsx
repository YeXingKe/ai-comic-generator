import { Link } from 'react-router-dom'
import './index.scss'

export default function GlobalFooter() {
  return (
    <footer className="global-footer">
      <div className="footer-inner">
        <div className="footer-brand">
          <span className="footer-logo">✦ AI 漫画生成器</span>
          <p className="footer-desc">AI 驱动的漫画内容创作平台</p>
        </div>
        <div className="footer-links">
          <Link to="/">首页</Link>
          <Link to="/create">创作</Link>
          <Link to="/history">历史</Link>
          <Link to="/user/center">用户</Link>
          <Link to="/data">数据</Link>
        </div>
        <p className="footer-copy">© {new Date().getFullYear()} AI Comic Generator. All rights reserved.</p>
      </div>
    </footer>
  )
}
