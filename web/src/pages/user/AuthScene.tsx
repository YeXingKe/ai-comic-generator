import type { CSSProperties } from 'react'
import './AuthScene.scss'

const STAR_COUNT = 140
const DUST_GRADIENT_COUNT = 85

/** 确定性伪随机，保证每次渲染一致但视觉无序 */
function createRng(seed: number) {
  let state = seed >>> 0
  return () => {
    state = (state * 1664525 + 1013904223) >>> 0
    return state / 4294967296
  }
}

type StarSpec = {
  left: string
  top: string
  size: number
  opacity: number
  delay: string
  duration: string
  drift: string
  driftX: string
  driftY: string
  variant: 'normal' | 'bright' | 'dust' | 'purple' | 'cyan'
}

function buildStars(count: number, seed: number): StarSpec[] {
  const rand = createRng(seed)
  const stars: StarSpec[] = []

  for (let i = 0; i < count; i++) {
    const roll = rand()
    const left = rand() * 99 + 0.5
    const top = rand() * 56 + 1
    const size =
      roll > 0.92 ? 2.8 + rand() * 1.4 : roll > 0.72 ? 1.6 + rand() * 0.9 : 0.6 + rand() * 1.1
    const opacity = 0.18 + rand() * 0.72
    const colorRoll = rand()

    let variant: StarSpec['variant'] = 'normal'
    if (size >= 3) variant = 'bright'
    else if (size < 1) variant = 'dust'
    else if (colorRoll > 0.82) variant = 'purple'
    else if (colorRoll > 0.68) variant = 'cyan'

    stars.push({
      left: `${left.toFixed(2)}%`,
      top: `${top.toFixed(2)}%`,
      size,
      opacity,
      delay: `${(rand() * 6).toFixed(2)}s`,
      duration: `${(1.8 + rand() * 4.5).toFixed(2)}s`,
      drift: `${(8 + rand() * 14).toFixed(1)}s`,
      driftX: `${((rand() - 0.5) * 8).toFixed(1)}px`,
      driftY: `${((rand() - 0.5) * 6).toFixed(1)}px`,
      variant,
    })
  }

  return stars
}

function buildDustBackground(count: number, seed: number): string {
  const rand = createRng(seed)
  const layers: string[] = []

  for (let i = 0; i < count; i++) {
    const x = rand() * 100
    const y = rand() * 58
    const px = rand() > 0.75 ? 1.4 : 1
    const alpha = 0.12 + rand() * 0.38
    const rgb =
      rand() > 0.88 ? '196,181,253' : rand() > 0.76 ? '186,230,253' : '255,255,255'
    layers.push(
      `radial-gradient(${px}px ${px}px at ${x.toFixed(2)}% ${y.toFixed(2)}%, rgba(${rgb},${alpha.toFixed(2)}), transparent)`,
    )
  }

  return layers.join(',')
}

const STARS = buildStars(STAR_COUNT, 0x6e4a9f)
const DUST_BACKGROUND = buildDustBackground(DUST_GRADIENT_COUNT, 0x3c1f7a)

function starStyle(star: StarSpec): CSSProperties {
  return {
    left: star.left,
    top: star.top,
    width: star.size,
    height: star.size,
    opacity: star.opacity,
    ['--star-delay' as string]: star.delay,
    ['--star-dur' as string]: star.duration,
    ['--star-drift-dur' as string]: star.drift,
    ['--star-dx' as string]: star.driftX,
    ['--star-dy' as string]: star.driftY,
  }
}

const METEORS = [
  { top: '8%', left: '78%', delay: '0s', duration: '14s' },
  { top: '14%', left: '62%', delay: '6s', duration: '16s' },
  { top: '5%', left: '90%', delay: '11s', duration: '18s' },
  { top: '20%', left: '48%', delay: '4s', duration: '20s' },
]

const FISH = [
  {
    path: 'M -60 300 Q 320 240, 640 280 Q 960 320, 1320 260 Q 1680 210, 2040 290',
    duration: 24,
    delay: 0,
    scale: 1,
    color: '#c4b5fd',
    tailSpeed: 0.38,
  },
  {
    path: 'M 2040 340 Q 1600 380, 1200 310 Q 800 260, 400 320 Q 0 360, -80 300',
    duration: 28,
    delay: 5,
    scale: 0.78,
    color: '#a5b4fc',
    tailSpeed: 0.42,
  },
  {
    path: 'M -80 220 Q 280 280, 560 240 Q 840 200, 1120 250 Q 1400 300, 1800 230 Q 2100 190, 2060 260',
    duration: 22,
    delay: 10,
    scale: 0.62,
    color: '#818cf8',
    tailSpeed: 0.35,
  },
  {
    path: 'M 2060 360 Q 1700 320, 1300 370 Q 900 400, 500 340 Q 100 290, -60 350',
    duration: 26,
    delay: 3,
    scale: 0.88,
    color: '#7dd3fc',
    tailSpeed: 0.4,
  },
  {
    path: 'M 400 380 Q 700 320, 1000 360 Q 1300 400, 1600 350 Q 1900 300, 2040 380',
    duration: 20,
    delay: 14,
    scale: 0.52,
    color: '#6366f1',
    tailSpeed: 0.32,
  },
  {
    path: 'M 1600 280 Q 1200 240, 800 290 Q 400 340, 100 280 Q -40 240, -70 310',
    duration: 23,
    delay: 8,
    scale: 0.72,
    color: '#e9d5ff',
    tailSpeed: 0.36,
  },
]

function SkyDecor() {
  return (
    <>
      <div
        className="auth-scene__stars auth-scene__stars--dense"
        style={{ backgroundImage: DUST_BACKGROUND }}
      />
      <div className="auth-scene__stars-layer">
        {STARS.map((star, i) => (
          <span
            key={i}
            className={`auth-scene__star auth-scene__star--${star.variant}`}
            style={starStyle(star)}
          />
        ))}
      </div>
      <div className="auth-scene__meteors">
        {METEORS.map((m, i) => (
          <span
            key={i}
            className="auth-scene__meteor"
            style={{
              top: m.top,
              left: m.left,
              animationDelay: m.delay,
              animationDuration: m.duration,
            }}
          />
        ))}
      </div>
    </>
  )
}

function FishIcon({ color }: { color: string }) {
  return (
    <svg viewBox="0 0 56 28" fill="none" className="auth-scene__fish-svg">
      <g className="auth-scene__fish-body">
        <ellipse cx="20" cy="14" rx="15" ry="9" fill={color} opacity="0.92" />
        <ellipse cx="20" cy="14" rx="11" ry="6.5" fill={color} opacity="0.55" />
        <g className="auth-scene__fish-fin auth-scene__fish-fin--dorsal">
          <path
            d="M14 6 C18 4, 24 4, 28 7"
            stroke={color}
            strokeWidth="2.5"
            strokeLinecap="round"
            fill="none"
            opacity="0.7"
          />
        </g>
        <g className="auth-scene__fish-fin auth-scene__fish-fin--pectoral">
          <path
            d="M16 16 C12 20, 10 22, 8 24"
            stroke={color}
            strokeWidth="2"
            strokeLinecap="round"
            fill="none"
            opacity="0.65"
          />
        </g>
        <circle cx="11" cy="13" r="2" fill="#1e1b4b" opacity="0.55" />
        <circle cx="10.5" cy="12.5" r="0.8" fill="#fff" opacity="0.8" />
      </g>
      <g className="auth-scene__fish-tail">
        <path d="M32 14 L44 5 L48 14 L44 23 Z" fill={color} opacity="0.88" />
        <path d="M32 14 L44 5 L44 23 Z" fill={color} opacity="0.45" />
      </g>
    </svg>
  )
}

function SeaFish() {
  return (
    <div className="auth-scene__fish-school">
      {FISH.map((fish, i) => (
        <div
          key={i}
          className="auth-scene__fish"
          style={{
            ['--fish-path' as string]: `path('${fish.path}')`,
            ['--tail-speed' as string]: `${fish.tailSpeed}s`,
            animationDuration: `${fish.duration}s`,
            animationDelay: `${fish.delay}s`,
          }}
        >
          <div
            className="auth-scene__fish-scaler"
            style={{ transform: `scale(${fish.scale})` }}
          >
            <FishIcon color={fish.color} />
          </div>
        </div>
      ))}
    </div>
  )
}

function OceanWaves() {
  const wavePath =
    'M0,64 C120,96 240,32 360,64 C480,96 600,32 720,64 C840,96 960,32 1080,64 C1200,96 1320,32 1440,64 L1440,200 L0,200 Z'

  return (
    <div className="auth-scene__ocean">
      {[1, 2, 3, 4].map((layer) => (
        <div key={layer} className={`auth-scene__wave-track auth-scene__wave-track--${layer}`}>
          <svg viewBox="0 0 1440 200" preserveAspectRatio="none" className="auth-scene__wave-svg">
            <path d={wavePath} />
          </svg>
          <svg viewBox="0 0 1440 200" preserveAspectRatio="none" className="auth-scene__wave-svg">
            <path d={wavePath} />
          </svg>
        </div>
      ))}
      <div className="auth-scene__ocean-depth" />
      <SeaFish />
      <div className="auth-scene__spray auth-scene__spray--left" />
      <div className="auth-scene__spray auth-scene__spray--right" />
      <div className="auth-scene__spray auth-scene__spray--center" />
    </div>
  )
}

function BattleIllustration() {
  return (
    <svg
      className="auth-scene__battle"
      viewBox="0 0 900 480"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id="skin" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#ffe8d6" />
          <stop offset="100%" stopColor="#e8b896" />
        </linearGradient>
        <linearGradient id="skinCold" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#f0f9ff" />
          <stop offset="100%" stopColor="#bae6fd" />
        </linearGradient>
        <linearGradient id="nezhaTop" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#ef4444" />
          <stop offset="100%" stopColor="#b91c1c" />
        </linearGradient>
        <linearGradient id="aobingRobe" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#f8fafc" />
          <stop offset="55%" stopColor="#e2e8f0" />
          <stop offset="100%" stopColor="#94a3b8" />
        </linearGradient>
        <linearGradient id="fireTip" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#fde047" />
          <stop offset="50%" stopColor="#f97316" />
          <stop offset="100%" stopColor="#ef4444" />
        </linearGradient>
        <linearGradient id="iceTip" x1="1" y1="0" x2="0" y2="0">
          <stop offset="0%" stopColor="#e0f2fe" />
          <stop offset="50%" stopColor="#38bdf8" />
          <stop offset="100%" stopColor="#0284c7" />
        </linearGradient>
        <linearGradient id="wheel" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#fef08a" />
          <stop offset="100%" stopColor="#ea580c" />
        </linearGradient>
        <radialGradient id="clashGlow" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#fff" stopOpacity="0.95" />
          <stop offset="35%" stopColor="#c4b5fd" stopOpacity="0.7" />
          <stop offset="70%" stopColor="#f97316" stopOpacity="0.35" />
          <stop offset="100%" stopColor="transparent" stopOpacity="0" />
        </radialGradient>
        <filter id="softGlow" x="-40%" y="-40%" width="180%" height="180%">
          <feGaussianBlur stdDeviation="4" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <filter id="iceGlow" x="-30%" y="-30%" width="160%" height="160%">
          <feGaussianBlur stdDeviation="3" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>

      {/* 交锋光效 */}
      <g className="auth-scene__clash">
        <circle cx="450" cy="218" r="72" fill="url(#clashGlow)" opacity="0.85" />
        <circle cx="450" cy="218" r="28" fill="#fff" opacity="0.55" />
        {[0, 45, 90, 135].map((deg) => (
          <line
            key={deg}
            x1="450"
            y1="218"
            x2="450"
            y2="168"
            stroke="#fde68a"
            strokeWidth="3"
            strokeLinecap="round"
            opacity="0.75"
            transform={`rotate(${deg} 450 218)`}
          />
        ))}
        <path
          d="M430 198 L450 178 L470 198 L450 238 Z"
          fill="#fff"
          opacity="0.5"
          filter="url(#softGlow)"
        />
      </g>

      {/* 哪吒 */}
      <g className="auth-scene__nezha">
        {/* 混天绫 */}
        <path
          className="auth-scene__silk auth-scene__silk--a"
          d="M210 130 C160 120, 120 150, 95 195 C130 175, 170 145, 210 155"
          stroke="#dc2626"
          strokeWidth="9"
          strokeLinecap="round"
          fill="none"
          opacity="0.9"
        />
        <path
          className="auth-scene__silk auth-scene__silk--b"
          d="M225 125 C270 105, 310 125, 340 165 C305 150, 265 130, 225 145"
          stroke="#ef4444"
          strokeWidth="8"
          strokeLinecap="round"
          fill="none"
        />

        {/* 风火轮 */}
        <g className="auth-scene__wheel auth-scene__wheel--nezha">
          <circle cx="168" cy="378" r="30" fill="url(#wheel)" filter="url(#softGlow)" />
          <circle cx="168" cy="378" r="16" fill="#fef9c3" opacity="0.65" />
          <g className="auth-scene__wheel-spokes" style={{ transformOrigin: '168px 378px' }}>
            {[0, 60, 120].map((d) => (
              <line key={d} x1="168" y1="378" x2="168" y2="350" stroke="#fff" strokeWidth="2.5" opacity="0.8" transform={`rotate(${d} 168 378)`} />
            ))}
          </g>
          <ellipse className="auth-scene__flame" cx="168" cy="398" rx="22" ry="10" fill="#fb923c" opacity="0.75" />
        </g>

        {/* 腿 - 踢击姿态 */}
        <path d="M195 300 L175 355" stroke="#4c1d95" strokeWidth="18" strokeLinecap="round" />
        <path d="M220 295 L248 340 L270 318" stroke="url(#skin)" strokeWidth="16" strokeLinecap="round" strokeLinejoin="round" fill="none" />

        {/* 躯干 */}
        <path d="M178 248 C195 235, 230 238, 245 258 L238 310 C220 322, 188 318, 172 298 Z" fill="#6d28d9" />
        <path d="M185 252 C200 242, 225 244, 235 258 L230 300 C218 308, 192 305, 180 292 Z" fill="url(#nezhaTop)" />
        <ellipse cx="208" cy="268" rx="8" ry="10" fill="#fbbf24" opacity="0.85" />

        {/* 左臂收，右臂刺枪 */}
        <path d="M178 255 C155 248, 138 230, 130 210" stroke="url(#skin)" strokeWidth="15" strokeLinecap="round" />
        <path d="M238 258 C275 245, 310 220, 335 195" stroke="url(#skin)" strokeWidth="14" strokeLinecap="round" />

        {/* 火尖枪 */}
        <g className="auth-scene__weapon auth-scene__weapon--fire">
          <line x1="335" y1="195" x2="430" y2="215" stroke="#cbd5e1" strokeWidth="5" strokeLinecap="round" />
          <path d="M430 215 L448 205 L442 225 L458 218 L438 232 Z" fill="url(#fireTip)" filter="url(#softGlow)" />
        </g>

        {/* 头 */}
        <circle cx="205" cy="195" r="36" fill="url(#skin)" />
        {/* 双丸子头 */}
        <circle cx="178" cy="162" r="16" fill="#0f172a" />
        <circle cx="232" cy="162" r="16" fill="#0f172a" />
        <circle cx="178" cy="162" r="10" fill="#1e293b" />
        <circle cx="232" cy="162" r="10" fill="#1e293b" />
        <path d="M172 158 C178 145, 188 142, 195 150" stroke="#ef4444" strokeWidth="4" strokeLinecap="round" fill="none" />
        <path d="M238 158 C232 145, 222 142, 215 150" stroke="#ef4444" strokeWidth="4" strokeLinecap="round" fill="none" />
        {/* 刘海 */}
        <path d="M172 178 C185 168, 210 165, 225 172 C218 182, 192 186, 172 182 Z" fill="#0f172a" />
        {/* 五官 */}
        <path d="M188 192 L198 188" stroke="#7c2d12" strokeWidth="2.5" strokeLinecap="round" />
        <path d="M218 192 L228 188" stroke="#7c2d12" strokeWidth="2.5" strokeLinecap="round" />
        <ellipse cx="194" cy="200" rx="5" ry="6" fill="#1e1b4b" />
        <ellipse cx="220" cy="200" rx="5" ry="6" fill="#1e1b4b" />
        <circle cx="196" cy="198" r="1.8" fill="#fff" />
        <circle cx="222" cy="198" r="1.8" fill="#fff" />
        <path d="M198 214 Q207 220 216 214" stroke="#c2410c" strokeWidth="2" fill="none" />
        {/* 怒眉 */}
        <path d="M186 188 L198 192" stroke="#1e1b4b" strokeWidth="2.5" strokeLinecap="round" />
        <path d="M230 188 L218 192" stroke="#1e1b4b" strokeWidth="2.5" strokeLinecap="round" />
      </g>

      {/* 敖丙 */}
      <g className="auth-scene__aobing">
        {/* 龙尾幻影 */}
        <path
          className="auth-scene__tail"
          d="M710 340 C760 360, 790 330, 820 350 C795 345, 755 355, 720 348"
          stroke="#38bdf8"
          strokeWidth="6"
          strokeLinecap="round"
          fill="none"
          opacity="0.55"
        />

        {/* 腿 */}
        <path d="M655 305 L640 365" stroke="#334155" strokeWidth="16" strokeLinecap="round" />
        <path d="M685 300 L705 358" stroke="#334155" strokeWidth="16" strokeLinecap="round" />

        {/* 白袍 */}
        <path d="M648 250 C670 235, 710 238, 728 258 L735 315 C720 330, 665 328, 645 305 Z" fill="url(#aobingRobe)" />
        <path d="M655 255 C672 245, 700 247, 715 260 L720 305 C708 315, 670 312, 652 298 Z" fill="#f1f5f9" opacity="0.85" />
        {/* 龙鳞肩甲 */}
        <path d="M648 252 L638 272 L652 278 L662 258 Z" fill="#7dd3fc" opacity="0.8" />
        <path d="M720 255 L732 275 L718 282 L708 262 Z" fill="#7dd3fc" opacity="0.8" />
        <path d="M640 268 C650 262, 660 268, 655 278" stroke="#0ea5e9" strokeWidth="1.5" fill="none" />
        <path d="M725 270 C715 264, 705 270, 710 280" stroke="#0ea5e9" strokeWidth="1.5" fill="none" />

        {/* 臂 - 格挡 */}
        <path d="M652 262 C625 255, 598 240, 575 218" stroke="url(#skinCold)" strokeWidth="14" strokeLinecap="round" />
        <path d="M718 260 C745 248, 768 228, 778 205" stroke="url(#skinCold)" strokeWidth="13" strokeLinecap="round" />

        {/* 冰凌枪 */}
        <g className="auth-scene__weapon auth-scene__weapon--ice">
          <line x1="575" y1="218" x2="468" y2="215" stroke="#e2e8f0" strokeWidth="5" strokeLinecap="round" />
          <path d="M468 215 L450 205 L456 225 L440 218 L460 232 Z" fill="url(#iceTip)" filter="url(#iceGlow)" />
          <polygon points="452,210 458,200 464,210 458,220" fill="#bae6fd" opacity="0.8" />
        </g>

        {/* 头 */}
        <circle cx="688" cy="198" r="34" fill="url(#skinCold)" />
        {/* 龙角 */}
        <path d="M668 168 L662 145 L674 162 Z" fill="#e0f2fe" stroke="#38bdf8" strokeWidth="1" />
        <path d="M708 168 L714 143 L700 160 Z" fill="#e0f2fe" stroke="#38bdf8" strokeWidth="1" />
        {/* 长发 */}
        <path
          className="auth-scene__hair"
          d="M658 178 C640 200, 628 240, 620 280 C635 250, 650 210, 668 190"
          fill="#1e3a5f"
        />
        <path
          className="auth-scene__hair auth-scene__hair--b"
          d="M718 178 C738 200, 752 245, 760 290 C742 255, 728 215, 712 192"
          fill="#0f2744"
        />
        <path d="M662 175 C680 162, 710 162, 728 178 C715 188, 675 188, 662 175 Z" fill="#1e3a5f" />
        {/* 五官 - 冷峻 */}
        <path d="M672 192 L682 190" stroke="#64748b" strokeWidth="2" strokeLinecap="round" />
        <path d="M698 192 L708 190" stroke="#64748b" strokeWidth="2" strokeLinecap="round" />
        <ellipse cx="678" cy="200" rx="4.5" ry="5.5" fill="#0c4a6e" />
        <ellipse cx="702" cy="200" rx="4.5" ry="5.5" fill="#0c4a6e" />
        <circle cx="679" cy="198" r="1.5" fill="#e0f2fe" />
        <circle cx="703" cy="198" r="1.5" fill="#e0f2fe" />
        <path d="M680 212 Q690 216 700 212" stroke="#64748b" strokeWidth="1.5" fill="none" />

        {/* 冰晶护盾 */}
        <path
          className="auth-scene__shield"
          d="M620 230 C600 250, 595 290, 615 320 C635 295, 640 255, 625 235 Z"
          fill="#38bdf8"
          opacity="0.25"
          stroke="#7dd3fc"
          strokeWidth="2"
        />
      </g>

      {/* 飞溅粒子 */}
      <g className="auth-scene__sparks">
        <circle cx="420" cy="200" r="4" fill="#fde047" />
        <circle cx="480" cy="228" r="3" fill="#38bdf8" />
        <circle cx="445" cy="245" r="3.5" fill="#f97316" />
        <circle cx="465" cy="190" r="2.5" fill="#e0f2fe" />
        <circle cx="435" cy="235" r="2" fill="#c4b5fd" />
      </g>
    </svg>
  )
}

export default function AuthScene() {
  return (
    <div className="auth-scene" aria-hidden="true">
      <div className="auth-scene__sky">
        <SkyDecor />
        <div className="auth-scene__moon" />
        <div className="auth-scene__glow auth-scene__glow--left" />
        <div className="auth-scene__glow auth-scene__glow--right" />
        <div className="auth-scene__glow auth-scene__glow--center" />
      </div>

      <div className="auth-scene__battle-wrap">
        <BattleIllustration />
        <p className="auth-scene__title">哪吒 · 敖丙</p>
        <p className="auth-scene__subtitle">魔丸灵珠，踏浪争锋</p>
      </div>

      <OceanWaves />
    </div>
  )
}
