import { useRouter } from 'next/router'
import { useEffect, useMemo, useRef, useState } from 'react'
import { get, put } from '../lib/api'
import Icon from './Icon'

const CATS = {
  core: 'Ядро викликів',
  staff: 'Персонал',
  equipment: 'Техніка',
  refs: 'Довідники',
  knowledge: 'База знань',
}
const CAT_COLORS = { core: '#dc2626', staff: '#16a34a', equipment: '#2563eb', refs: '#9333ea', knowledge: '#65a30d' }
const TYPE_SHORT = { text: 'text', number: 'num', date: 'date', bool: 'bool', select: 'enum' }

const NW = 200
const HEAD = 30
const ROWH = 16

// NodeGraph — інтерактивна ERD-мапа структури БД.
// Зум (коліщатко/кнопки), панорама (фон), перетягування нод і ядра,
// вигин FK-ліній за ручки, фільтри гілок, мінімапа.
// Клік (без зсуву) відкриває таблицю; утримання+рух — лише перетягування.
// Усі позиції зберігаються у БД (graph_layouts).
export default function NodeGraph({ tables }) {
  const router = useRouter()
  const W = 1100
  const H = 700
  const cx = W / 2
  const cy = H / 2
  const R = 280
  const svgRef = useRef(null)

  const base = useMemo(
    () =>
      tables.map((t, i) => {
        const a = (2 * Math.PI * i) / tables.length - Math.PI / 2
        return { ...t, bx: cx + R * Math.cos(a), by: cy + R * Math.sin(a) }
      }),
    [tables]
  )

  const [saved, setSaved] = useState({}) // з БД
  const [offsets, setOffsets] = useState({}) // поточне перетягування (ще не збережено)
  const [view, setView] = useState({ x: 0, y: 0, k: 1 })
  const [panStart, setPanStart] = useState(null)
  const [drag, setDrag] = useState(null)
  const [bendDrag, setBendDrag] = useState(null)
  const [hover, setHover] = useState(null)
  const [cats, setCats] = useState(() => new Set(Object.keys(CATS)))
  const moved = useRef(0)

  useEffect(() => {
    get('/api/layout').then(setSaved).catch(() => {})
  }, [])

  const off = (key) => offsets[key] ?? saved[key] ?? { dx: 0, dy: 0 }
  const nodeH = (t) => HEAD + 10 + t.columns.length * ROWH

  const nodes = base
    .filter((n) => cats.has(n.category))
    .map((n) => {
      const o = off(n.name)
      return { ...n, x: n.bx + o.dx, y: n.by + o.dy, h: nodeH(n) }
    })
  const byName = Object.fromEntries(nodes.map((n) => [n.name, n]))
  const coreO = off('core')
  const core = { x: cx + coreO.dx, y: cy + coreO.dy }

  const fk = []
  nodes.forEach((t) =>
    t.columns.forEach((c) => {
      if (c.type === 'ref' && byName[c.ref]) {
        const key = `${t.name}.${c.name}`
        fk.push({ key, from: byName[t.name], to: byName[c.ref], label: c.name })
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
  const dim = (n) => (linked && !linked.has(n) ? 0.12 : 1)

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

  const persist = (key) => {
    const o = off(key)
    setSaved((s) => ({ ...s, [key]: o }))
    setOffsets((os) => {
      const n = { ...os }
      delete n[key]
      return n
    })
    put('/api/layout', { node: key, dx: o.dx, dy: o.dy }).catch(() => {})
  }

  const onBgDown = (e) => {
    const s = toScreen(e)
    moved.current = 0
    setPanStart({ sx: s.x, sy: s.y, ox: view.x, oy: view.y })
  }
  const onMove = (e) => {
    const active = drag || bendDrag
    if (active) {
      const p = toWorld(e)
      moved.current++
      setOffsets((o) => ({
        ...o,
        [active.key]: { dx: active.odx + p.x - active.px, dy: active.ody + p.y - active.py },
      }))
    } else if (panStart) {
      const s = toScreen(e)
      setView((v) => ({ ...v, x: panStart.ox + s.x - panStart.sx, y: panStart.oy + s.y - panStart.sy }))
    }
  }
  const onUp = () => {
    if (drag && moved.current > 2) persist(drag.key)
    else if (drag) setOffsets((o) => ({ ...o }))
    if (bendDrag) persist(bendDrag.key)
    setPanStart(null)
    setDrag(null)
    setBendDrag(null)
  }

  const startDrag = (key) => (e) => {
    e.stopPropagation()
    const p = toWorld(e)
    const o = off(key)
    moved.current = 0
    setDrag({ key, px: p.x, py: p.y, odx: o.dx, ody: o.dy })
  }
  const startBend = (key) => (e) => {
    e.stopPropagation()
    const p = toWorld(e)
    const o = off(`edge:${key}`)
    setBendDrag({ key: `edge:${key}`, px: p.x, py: p.y, odx: o.dx, ody: o.dy })
  }

  const toggleCat = (k) =>
    setCats((s) => {
      const n = new Set(s)
      if (n.has(k)) n.delete(k)
      else n.add(k)
      return n
    })

  const typeLabel = (c) => (c.type === 'ref' ? `→ ${c.ref}` : TYPE_SHORT[c.type] || c.type)

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
          onMouseDown={onBgDown}
          onMouseMove={onMove}
          onMouseUp={onUp}
          onMouseLeave={onUp}
        >
          <g transform={`translate(${view.x} ${view.y}) scale(${view.k})`}>
            {nodes.map((n) => (
              <line
                key={`b-${n.name}`}
                x1={core.x}
                y1={core.y}
                x2={n.x}
                y2={n.y}
                stroke={n.color}
                strokeWidth="2"
                opacity={0.55 * dim(n.name)}
              />
            ))}

            {fk.map((e) => {
              const b = off(`edge:${e.key}`)
              const mx = (e.from.x + e.to.x) / 2 + b.dx
              const my = (e.from.y + e.to.y) / 2 + b.dy
              const op = 0.85 * Math.min(dim(e.from.name), dim(e.to.name))
              return (
                <g key={e.key} opacity={op}>
                  <path
                    d={`M ${e.from.x} ${e.from.y} Q ${mx} ${my} ${e.to.x} ${e.to.y}`}
                    fill="none"
                    stroke="#94a3b8"
                    strokeWidth="1.3"
                    strokeDasharray="5 4"
                  />
                  <circle cx={mx} cy={my} r="5" className="edge-handle" onMouseDown={startBend(e.key)} />
                  <text x={mx + 8} y={my - 6} className="node-graph__edgename">
                    {e.label}
                  </text>
                </g>
              )
            })}

            <g className="node-graph__core" onMouseDown={startDrag('core')}>
              <rect x={core.x - 82} y={core.y - 24} width="164" height="48" rx="12" />
              <text x={core.x} y={core.y + 5} textAnchor="middle">
                fire_station.accdb
              </text>
            </g>

            {nodes.map((n) => (
              <g
                key={n.name}
                className="node-graph__node"
                opacity={dim(n.name)}
                onMouseDown={startDrag(n.name)}
                onMouseEnter={() => setHover(n.name)}
                onMouseLeave={() => setHover(null)}
                onClick={() => {
                  if (moved.current < 5) router.push(`/tables/${n.name}`)
                }}
              >
                <rect x={n.x - NW / 2} y={n.y - n.h / 2} width={NW} height={n.h} rx="10" style={{ stroke: n.color }} />
                <path
                  d={`M ${n.x - NW / 2 + 10} ${n.y - n.h / 2} h ${NW - 20} a 10 10 0 0 1 10 10 v ${HEAD - 10} h ${-NW} v ${-(HEAD - 10)} a 10 10 0 0 1 10 -10 z`}
                  fill={n.color}
                  stroke="none"
                />
                <text x={n.x} y={n.y - n.h / 2 + HEAD / 2 + 4} textAnchor="middle" className="node-graph__headtext">
                  {n.label}
                </text>
                {n.columns.map((c, i) => (
                  <g key={c.name}>
                    <text x={n.x - NW / 2 + 12} y={n.y - n.h / 2 + HEAD + 16 + i * ROWH} className="node-graph__col">
                      {c.name}
                    </text>
                    <text x={n.x + NW / 2 - 12} y={n.y - n.h / 2 + HEAD + 16 + i * ROWH} textAnchor="end" className="node-graph__type">
                      {typeLabel(c)}
                    </text>
                  </g>
                ))}
              </g>
            ))}
          </g>
        </svg>

        <div className="node-graph__controls">
          <button className="icon-btn" title="Збільшити" onClick={() => setView((v) => ({ ...v, k: clampK(v.k * 1.2) }))}>
            +
          </button>
          <button className="icon-btn" title="Зменшити" onClick={() => setView((v) => ({ ...v, k: clampK(v.k / 1.2) }))}>
            −
          </button>
          <button
            className="icon-btn"
            title="Скинути вигляд і розкладку"
            onClick={() => {
              setView({ x: 0, y: 0, k: 1 })
              setOffsets({})
            }}
          >
            <Icon name="refresh" size={13} />
          </button>
        </div>

        <svg viewBox={`0 0 ${W} ${H}`} className="node-graph__minimap" style={{ width: 150, height: 150 * (H / W) }}>
          <rect x="0" y="0" width={W} height={H} className="node-graph__minimap-bg" />
          <circle cx={core.x} cy={core.y} r="18" fill="#111827" />
          {nodes.map((n) => (
            <rect key={n.name} x={n.x - 30} y={n.y - 16} width="60" height="32" rx="6" fill={n.color} opacity="0.8" />
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
        Коліщатко або +/− — масштаб • тягни фон — панорама • тягни й тримай ноду чи ядро — переставити •
        точка на пунктирній лінії — вигин зв'язку • клік по ноді (без руху) — відкрити таблицю •
        розкладка зберігається у БД
      </p>
    </div>
  )
}
