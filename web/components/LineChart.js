import { useState } from 'react'

// LineChart — простий SVG-графік з сіткою, точками й підказкою при наведенні.
export default function LineChart({ data, height = 190 }) {
  const W = 760
  const H = height
  const padL = 34
  const padR = 12
  const padT = 14
  const padB = 26
  const [hov, setHov] = useState(null)

  if (!data?.length) return null

  const max = Math.max(1, ...data.map((d) => d.value))
  const iw = W - padL - padR
  const ih = H - padT - padB
  const sx = (i) => padL + (i * iw) / Math.max(1, data.length - 1)
  const sy = (v) => padT + ih - (v / max) * ih
  const pts = data.map((d, i) => [sx(i), sy(d.value)])
  const line = pts.map((p, i) => `${i ? 'L' : 'M'}${p[0].toFixed(1)} ${p[1].toFixed(1)}`).join(' ')
  const area = `${line} L${sx(data.length - 1).toFixed(1)} ${padT + ih} L${padL} ${padT + ih} Z`
  const ticks = [0, Math.ceil(max / 2), max]
  const labelEvery = Math.ceil(data.length / 7)

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="line-chart">
      {ticks.map((t) => (
        <g key={t}>
          <line x1={padL} x2={W - padR} y1={sy(t)} y2={sy(t)} className="line-chart__grid" />
          <text x={padL - 7} y={sy(t) + 4} textAnchor="end" className="line-chart__tick">
            {t}
          </text>
        </g>
      ))}
      <path d={area} className="line-chart__area" />
      <path d={line} className="line-chart__line" />
      {pts.map((p, i) => (
        <g key={i}>
          <rect
            x={p[0] - iw / data.length / 2}
            y={padT}
            width={iw / data.length}
            height={ih}
            fill="transparent"
            onMouseEnter={() => setHov(i)}
            onMouseLeave={() => setHov(null)}
          />
          <circle cx={p[0]} cy={p[1]} r={hov === i ? 4.5 : 2.5} className="line-chart__dot" />
          {i % labelEvery === 0 && (
            <text x={p[0]} y={H - 8} textAnchor="middle" className="line-chart__tick">
              {data[i].label}
            </text>
          )}
        </g>
      ))}
      {hov != null && (
        <text x={pts[hov][0]} y={Math.max(12, pts[hov][1] - 10)} textAnchor="middle" className="line-chart__tip">
          {data[hov].label}: {data[hov].value}
        </text>
      )}
    </svg>
  )
}
