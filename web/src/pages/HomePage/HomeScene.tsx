import type { CSSProperties } from 'react'
import './HomeScene.scss'

const STAR_COUNT = 100

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
  variant: 'normal' | 'bright' | 'purple'
}

function buildStars(count: number, seed: number): StarSpec[] {
  const rand = createRng(seed)
  const stars: StarSpec[] = []

  for (let i = 0; i < count; i++) {
    const roll = rand()
    const size = roll > 0.9 ? 2.4 + rand() * 1.2 : roll > 0.7 ? 1.4 + rand() * 0.8 : 0.6 + rand() * 0.9
    let variant: StarSpec['variant'] = 'normal'
    if (size >= 2.5) variant = 'bright'
    else if (rand() > 0.82) variant = 'purple'

    stars.push({
      left: `${(rand() * 99 + 0.5).toFixed(2)}%`,
      top: `${(rand() * 99 + 0.5).toFixed(2)}%`,
      size,
      opacity: 0.2 + rand() * 0.75,
      delay: `${(rand() * 5).toFixed(2)}s`,
      duration: `${(2 + rand() * 4).toFixed(2)}s`,
      variant,
    })
  }

  return stars
}

const STARS = buildStars(STAR_COUNT, 0x7a3cf1)

function starStyle(star: StarSpec): CSSProperties {
  return {
    left: star.left,
    top: star.top,
    width: star.size,
    height: star.size,
    opacity: star.opacity,
    ['--star-delay' as string]: star.delay,
    ['--star-dur' as string]: star.duration,
  }
}

/** 左下 · 奶龙 — 经典圆滚滚 Q 版，面向右侧 */
function NailongCharacter() {
  return (
    <svg viewBox="0 0 400 480" fill="none" xmlns="http://www.w3.org/2000/svg">
      <ellipse cx="200" cy="456" rx="108" ry="15" fill="rgba(88,28,135,0.32)" />

      <g className="home-scene__nailong-tail">
        <path
          d="M92 310c-28 36-38 82-24 120 16 44 58 68 96 52 22-10 38-30 44-54"
          fill="#84cc16"
        />
        <path d="M104 328c-16 22-22 52-14 78 12 36 46 54 76 42" fill="#65a30d" />
      </g>

      <g className="home-scene__nailong-wing">
        <ellipse cx="108" cy="292" rx="28" ry="36" fill="#bef264" opacity="0.85" />
        <ellipse cx="108" cy="292" rx="18" ry="24" fill="#d9f99d" opacity="0.6" />
      </g>

      <g className="home-scene__nailong-body">
        <ellipse cx="200" cy="328" rx="96" ry="108" fill="#fde047" />
        <ellipse cx="200" cy="318" rx="82" ry="92" fill="#fef08a" />
        <ellipse cx="200" cy="248" rx="86" ry="80" fill="#fde047" />
        <ellipse cx="200" cy="242" rx="76" ry="70" fill="#fef08a" />

        <ellipse cx="168" cy="392" rx="22" ry="16" fill="#fde047" />
        <ellipse cx="232" cy="392" rx="22" ry="16" fill="#fde047" />
        <ellipse cx="168" cy="402" rx="16" ry="10" fill="#eab308" />
        <ellipse cx="232" cy="402" rx="16" ry="10" fill="#eab308" />
      </g>

      <g className="home-scene__nailong-face">
        <path d="M158 168 L148 138 L168 162 Z" fill="#ca8a04" />
        <path d="M242 168 L252 138 L232 162 Z" fill="#ca8a04" />
        <circle cx="200" cy="168" r="10" fill="#a855f7" />
        <path d="M192 162 Q200 152 208 162 Q200 170 192 162" fill="#9333ea" />

        <ellipse cx="168" cy="232" rx="22" ry="26" fill="#fff" />
        <ellipse cx="232" cy="232" rx="22" ry="26" fill="#fff" />
        <circle cx="168" cy="236" r="13" fill="#1e293b" />
        <circle cx="232" cy="236" r="13" fill="#1e293b" />
        <circle cx="163" cy="228" r="5" fill="#fff" />
        <circle cx="227" cy="228" r="5" fill="#fff" />

        <ellipse cx="142" cy="252" rx="14" ry="8" fill="#fda4af" opacity="0.65" />
        <ellipse cx="258" cy="252" rx="14" ry="8" fill="#fda4af" opacity="0.65" />

        <path
          d="M178 268 Q200 284 222 268"
          stroke="#ca8a04"
          strokeWidth="4.5"
          strokeLinecap="round"
          fill="none"
        />
        <ellipse cx="192" cy="262" rx="4" ry="3" fill="#eab308" opacity="0.45" />
        <ellipse cx="208" cy="262" rx="4" ry="3" fill="#eab308" opacity="0.45" />
      </g>

      <g className="home-scene__nailong-wave">
        <ellipse cx="298" cy="300" rx="20" ry="18" fill="#fde047" />
        <ellipse cx="298" cy="300" rx="14" ry="12" fill="#fef08a" />
        <path
          d="M286 288c8-14 22-18 32-10"
          stroke="#fde047"
          strokeWidth="16"
          strokeLinecap="round"
        />
      </g>
    </svg>
  )
}

/** 右下 · 牛牛 — 同款 Q 版比例，面向左侧 */
function NiuniuCharacter() {
  return (
    <svg viewBox="0 0 400 480" fill="none" xmlns="http://www.w3.org/2000/svg">
      <ellipse cx="200" cy="456" rx="108" ry="15" fill="rgba(88,28,135,0.32)" />

      <g className="home-scene__niuniu-tail">
        <path
          d="M308 340c20 28 24 62 12 92-14 34-48 52-78 40"
          stroke="#c4b5fd"
          strokeWidth="14"
          strokeLinecap="round"
          fill="none"
        />
        <circle cx="318" cy="338" r="10" fill="#fff7ed" stroke="#ddd6fe" strokeWidth="3" />
      </g>

      <g className="home-scene__niuniu-body">
        <ellipse cx="200" cy="330" rx="94" ry="106" fill="#fff7ed" />
        <ellipse cx="200" cy="320" rx="80" ry="90" fill="#fffbeb" />
        <ellipse cx="200" cy="250" rx="88" ry="82" fill="#fff7ed" />
        <ellipse cx="200" cy="244" rx="78" ry="72" fill="#fffbeb" />

        <circle cx="168" cy="310" r="14" fill="#c4b5fd" opacity="0.55" />
        <circle cx="228" cy="328" r="11" fill="#a78bfa" opacity="0.5" />
        <circle cx="186" cy="348" r="9" fill="#ddd6fe" opacity="0.45" />

        <ellipse cx="166" cy="394" rx="22" ry="16" fill="#fff7ed" />
        <ellipse cx="234" cy="394" rx="22" ry="16" fill="#fff7ed" />
        <ellipse cx="166" cy="404" rx="16" ry="10" fill="#fdba74" />
        <ellipse cx="234" cy="404" rx="16" ry="10" fill="#fdba74" />
      </g>

      <g className="home-scene__niuniu-hoodie">
        <path
          d="M128 290c8-42 36-68 72-68s64 26 72 68v36c0 18-32 32-72 32s-72-14-72-32v-36z"
          fill="#8b5cf6"
          opacity="0.92"
        />
        <path
          d="M156 278c12-22 28-34 44-34s32 12 44 34"
          stroke="#c4b5fd"
          strokeWidth="6"
          strokeLinecap="round"
          fill="none"
        />
      </g>

      <g className="home-scene__niuniu-face">
        <ellipse cx="148" cy="218" rx="22" ry="30" fill="#fff7ed" stroke="#fed7aa" strokeWidth="3" />
        <ellipse cx="252" cy="218" rx="22" ry="30" fill="#fff7ed" stroke="#fed7aa" strokeWidth="3" />
        <ellipse cx="148" cy="228" rx="10" ry="14" fill="#fda4af" opacity="0.45" />
        <ellipse cx="252" cy="228" rx="10" ry="14" fill="#fda4af" opacity="0.45" />

        <path d="M172 188 L164 158 L180 182 Z" fill="#fde68a" />
        <path d="M228 188 L236 158 L220 182 Z" fill="#fde68a" />
        <ellipse cx="178" cy="186" rx="8" ry="10" fill="#fef3c7" />
        <ellipse cx="222" cy="186" rx="8" ry="10" fill="#fef3c7" />

        <ellipse cx="200" cy="262" rx="36" ry="28" fill="#fda4af" />
        <ellipse cx="200" cy="258" rx="30" ry="22" fill="#fecdd3" />
        <ellipse cx="192" cy="262" rx="5" ry="4" fill="#fb7185" opacity="0.55" />
        <ellipse cx="208" cy="262" rx="5" ry="4" fill="#fb7185" opacity="0.55" />

        <ellipse cx="172" cy="238" rx="20" ry="24" fill="#fff" />
        <ellipse cx="228" cy="238" rx="20" ry="24" fill="#fff" />
        <circle cx="172" cy="242" r="12" fill="#1e293b" />
        <circle cx="228" cy="242" r="12" fill="#1e293b" />
        <circle cx="167" cy="236" r="4.5" fill="#fff" />
        <circle cx="223" cy="236" r="4.5" fill="#fff" />

        <ellipse cx="146" cy="258" rx="13" ry="7" fill="#fda4af" opacity="0.6" />
        <ellipse cx="254" cy="258" rx="13" ry="7" fill="#fda4af" opacity="0.6" />

        <path
          d="M184 278 Q200 288 216 278"
          stroke="#e11d48"
          strokeWidth="3.5"
          strokeLinecap="round"
          fill="none"
        />
      </g>

      <g className="home-scene__niuniu-wave">
        <ellipse cx="102" cy="302" rx="20" ry="18" fill="#fff7ed" />
        <ellipse cx="102" cy="302" rx="14" ry="12" fill="#fffbeb" />
        <path
          d="M114 290c-8-14-22-18-32-10"
          stroke="#fff7ed"
          strokeWidth="16"
          strokeLinecap="round"
        />
      </g>
    </svg>
  )
}

export default function HomeScene() {
  return (
    <div className="home-scene" aria-hidden>
      <div className="home-scene__base" />
      <div className="home-scene__aurora home-scene__aurora--1" />
      <div className="home-scene__aurora home-scene__aurora--2" />
      <div className="home-scene__aurora home-scene__aurora--3" />
      <div className="home-scene__grid" />
      <div className="home-scene__vignette" />

      <div className="home-scene__stars">
        {STARS.map((star, i) => (
          <span
            key={i}
            className={`home-scene__star home-scene__star--${star.variant}`}
            style={starStyle(star)}
          />
        ))}
      </div>

      <div className="home-scene__char home-scene__char--nailong">
        <NailongCharacter />
      </div>
      <div className="home-scene__char home-scene__char--niuniu">
        <NiuniuCharacter />
      </div>
    </div>
  )
}
