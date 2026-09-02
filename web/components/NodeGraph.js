import { useRouter } from 'next/router'
import { useEffect, useMemo, useRef, useState } from 'react'

const CATS = {
  core: 'Ядро викликів',
  staff: 'Персонал',
  equipment: 'Техніка',
  refs: 'Довідники',
}
const CAT_COLORS = { core: '#dc2626', staff: '#16a34a', equipment: '#2563eb', refs: '#9333ea' }

// NodeGraph — інтерактивна мапа структури БД:
// зум (коліщатко/кнопки), панорама (тягнути фон), перетягування нод,
// фільтри гілок за категоріями, підсвітка зв'язків при наведенні, мінімапа.
export default function NodeGraph({ tables }) {
  const router = useRouter()
  const W = 1100
  const H = 660
  const cx = W / 2
  const cy = H / 2
  const R = 260
  const svgRef = useRef(null)

  const base = useMemo(
    () =>
      tables.map((t, i) => {
        const a = (2 * Math.PI * i) / tables.length - Math.PI / 2
        return { ...t, bx: cx + R * Math.cos(a), by: cy + R * Math.sin(a) }
      }),
    [tables]
  )

  const [offsets, setOffsets] = useState({})
  const [view, setView] = useState({ x: 0, y: 0, k: 1 })
  const [panStart, setPanStart] = useState(null)
  const [drag, setDrag] = useState(null)
  const [hover, setHover] = useState(null)
  const [cats, setCats] = useState(() => new Set(Object.keys(CATS)))

  const nodes = base
    .filter((n) => cats.has(n.category))
    .map((n) => ({ ...n, x: n.bx + (offsets[n.name]?.dx || 0), y: n.by + (offsets[n.name]?.dy || 0) }))
  const byName = Object.fromEntries(nodes.map((n) => [n.name, n]))

  const fk = []
  nodes.forEach((t) =>
    t.columns.forEach((c) => {
      if (c.type === 'ref' && byName[c.ref]) {
        fk.push({ key: `${t.name}.${c.name}`, from: byName[t.name], to: byName[c.ref] })
      }
    })
  )

  const linked = hover ? new Set([hover]) : null
  if (linked) {
    fk.forEach((e) => {
      if (e.from.name === hover) linked.add(e.to.name)
      if (e.to.name === hover) linked.add(e.from.name)
    })
  }
  const dim = (name) => (linked && !linked.has(name) ? 0.1 : 1)

  const clampK = (k) => Math.min(2.5, Math.max(0.4, k))

  useEffect(() => {
    const el = svgRef.current
    if (!el) return
    const h = (e) => {
      e.preventDefault()
      setView((v) => ({ ...v, k: clampK(v.k * (e.deltaY < 0 ? 1.12 : 0.89)) }))
    }
    el.addEventListener('wheel', h, { passive: false })
    return () => el.removeEventListener('wheel', h)
  }, [])

  const toScreen = (e) => {
    const r = svgRef.current.getBoundingClientRect()
    return { x: (e.clientX - r.left) * (W / r.width), y: (e.clientY - r.top) * (H / r.height) }
  }
  const toWorld = (e) => {
    const s = toScreen(e)
    return { x: (s.x - view.x) / view.k, y: (s.y - view.y) / view.k }
  }

  const onMouseDown = (e) => {
    const s = toScreen(e)
    setPanStart({ sx: s.x, sy: s.y, ox: view.x, oy: view.y })
  }
  const onMouseMove = (e) => {
    if (drag) {
      const p = toWorld(e)
      setOffsets((o) => ({
        ...o,
        [drag.name]: { dx: drag.odx + p.x - drag.px, dy: drag.ody + p.y - drag.py },
      }))
    } else if (panStart) {
      const s = toScreen(e)
      setView((v) => ({ ...v, x: panStart.ox + s.x - panStart.sx, y: panStart.oy + s.y - panStart.sy }))
    }
  }
  const endGesture = () => {
    setPanStart(null)
    setDrag(null)
  }

  const nodeDown = (n) => (e) => {
    e.stopPropagation()
    const p = toWorld(e)
    setDrag({ name: n.name, px: p.x, py: p.y, odx: offsets[n.name]?.dx || 0, ody: offsets[n.name]?.dy || 0 })
  }

  const toggleCat = (k) =>
    setCats((s) => {
      const n = new Set(s)
      if (n.has(k)) n.delete(k)
      else n.add(k)
      return n
    })

  return (
    <div>
      <div className="node-filters">
        {Object.entries(CATS).map(([k, label]) => (
          <button
            key={k}
            className={`fchip${cats.has(k) ? ' fchip--on' : ''}`}
            style={cats.has(k) ? { borderColor: CAT_COLORS[k], color: CAT_COLORS[k] } : {}}
            onClick={() => toggleCat(k)}
          >
            {label} ({base.filter((b) => b.category === k).length})
          </button>
        ))}
      </div>

      <div className="node-stage">
        <svg
          ref={svgRef}
          viewBox={`0 0 ${W} ${H}`}
          className="node-graph"
          onMouseDown={onMouseDown}
          onMouseMove={onMouseMove}
          onMouseUp={endGesture}
          onMouseLeave={endGesture}
        >
          <g transform={`translate(${view.x} ${view.y}) scale(${view.k})`}>
            {nodes.map((n) => (
              <line
                key={`b-${n.name}`}
                x1={cx}
                y1={cy}
                x2={n.x}
                y2={n.y}
                stroke={n.color}
                strokeWidth="2"
                opacity={0.7 * dim(n.name)}
              />
            ))}
            {fk.map((e) => (
              <line
                key={e.key}
                x1={e.from.x}
                y1={e.from.y}
                x2={e.to.x}
                y2={e.to.y}
                stroke="#94a3b8"
                strokeWidth="1.2"
                strokeDasharray="5 4"
                opacity={0.8 * Math.min(dim(e.from.name), dim(e.to.name))}
              />
            ))}

            <g className="node-graph__core">
              <rect x={cx - 80} y={cy - 24} width="160" height="48" rx="12" />
              <text x={cx} y={cy + 5} textAnchor="middle">
                fire_station.accdb
              </text>
            </g>

            {nodes.map((n) => (
              <g
                key={n.name}
                className="node-graph__node"
                opacity={dim(n.name)}
                onMouseDown={nodeDown(n)}
                onMouseEnter={() => setHover(n.name)}
                onMouseLeave={() => setHover(null)}
                onClick={() => !drag && router.push(`/tables/${n.name}`)}
              >
                <rect x={n.x - 82} y={n.y - 26} width="164" height="52" rx="12" style={{ stroke: n.color }} />
                <rect x={n.x - 82} y={n.y - 26} width="7" height="52" rx="3" fill={n.color} stroke="none" />
                <text x={n.x + 4} y={n.y - 3} textAnchor="middle" className="node-graph__label">
                  {n.label}
                </text>
                <text x={n.x + 4} y={n.y + 14} textAnchor="middle" className="node-graph__name">
                  {n.name}
                </text>
              </g>
            ))}
          </g>
        </svg>

        <div className="node-graph__controls">
          <button className="icon-btn" onClick={() => setView((v) => ({ ...v, k: clampK(v.k * 1.2) }))}>
            +
          </button>
          <button className="icon-btn" onClick={() => setView((v) => ({ ...v, k: clampK(v.k / 1.2) }))}>
            −
          </button>
          <button className="icon-btn" onClick={() => { setView({ x: 0, y: 0, k: 1 }); setOffsets({}) }}>
            ⟲
          </button>
        </div>

        <svg viewBox={`0 0 ${W} ${H}`} className="node-graph__minimap" style={{ width: 150, height: 150 * (H / W) }}>
          <rect x="0" y="0" width={W} height={H} className="node-graph__minimap-bg" />
          {nodes.map((n) => (
            <rect key={n.name} x={n.x - 30} y={n.y - 14} width="60" height="28" rx="6" fill={n.color} opacity="0.8" />
          ))}
          <rect
            x={-view.x / view.k}
            y={-view.y / view.k}
            width={W / view.k}
            height={H / view.k}
            className="node-graph__viewport"
          />
        </svg>
      </div>

      <p className="muted">
        Коліщатко або +/− — масштаб • тягни фон — панорама • тягни ноду — переставити • клік по ноді —
        відкрити таблицю • суцільні лінії — гілки ядра (колір ноди), пунктир — FK-зв'язки
      </p>
    </div>
  )
}
