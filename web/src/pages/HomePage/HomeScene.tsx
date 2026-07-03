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

const FLOATING_ORBS = [
  { left: '8%', top: '18%', size: 120, color: 'rgba(167, 139, 250, 0.18)', delay: '0s', duration: '14s' },
  { left: '82%', top: '22%', size: 90, color: 'rgba(244, 114, 182, 0.14)', delay: '-3s', duration: '16s' },
  { left: '72%', top: '62%', size: 70, color: 'rgba(96, 165, 250, 0.12)', delay: '-6s', duration: '12s' },
  { left: '14%', top: '58%', size: 56, color: 'rgba(250, 204, 21, 0.1)', delay: '-2s', duration: '18s' },
] as const

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

export default function HomeScene() {
  return (
    <div className="home-scene" aria-hidden>
      <div className="home-scene__base" />
      <div className="home-scene__aurora home-scene__aurora--1" />
      <div className="home-scene__aurora home-scene__aurora--2" />
      <div className="home-scene__aurora home-scene__aurora--3" />
      <div className="home-scene__grid" />
      <div className="home-scene__hero-glow" />

      <div className="home-scene__stars">
        {STARS.map((star, i) => (
          <span
            key={i}
            className={`home-scene__star home-scene__star--${star.variant}`}
            style={starStyle(star)}
          />
        ))}
      </div>

      <div className="home-scene__orbs">
        {FLOATING_ORBS.map((orb, i) => (
          <span
            key={i}
            className="home-scene__orb"
            style={{
              left: orb.left,
              top: orb.top,
              width: orb.size,
              height: orb.size,
              background: orb.color,
              ['--orb-delay' as string]: orb.delay,
              ['--orb-dur' as string]: orb.duration,
            }}
          />
        ))}
      </div>

      <div className="home-scene__vignette" />
    </div>
  )
}
