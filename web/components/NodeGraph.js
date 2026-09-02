import { useRouter } from 'next/router'

// NodeGraph — динамічний SVG-граф структури БД.
// Центральна нода — файл БД; навколо — ноди-таблиці.
// Суцільна кольорова лінія = гілка ядра до ноди (колір ноди),
// пунктирна сіра = FK-зв'язок між таблицями. Клік по ноді — перехід до таблиці.
export default function NodeGraph({ tables }) {
  const router = useRouter()
  const W = 980
  const H = 640
  const cx = W / 2
  const cy = H / 2
  const R = 240

  const nodes = tables.map((t, i) => {
    const a = (2 * Math.PI * i) / tables.length - Math.PI / 2
    return { ...t, x: cx + R * Math.cos(a), y: cy + R * Math.sin(a) }
  })
  const byName = Object.fromEntries(nodes.map((n) => [n.name, n]))

  const fkEdges = []
  tables.forEach((t) =>
    t.columns.forEach((c) => {
      if (c.type === 'ref' && byName[c.ref] && byName[t.name]) {
        fkEdges.push({ from: byName[t.name], to: byName[c.ref], key: `${t.name}.${c.name}` })
      }
    })
  )

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="node-graph" role="img" aria-label="Структура бази даних">
      {nodes.map((n) => (
        <line key={`branch-${n.name}`} x1={cx} y1={cy} x2={n.x} y2={n.y} stroke={n.color} strokeWidth="2" opacity="0.7" />
      ))}
      {fkEdges.map((e) => (
        <line key={e.key} x1={e.from.x} y1={e.from.y} x2={e.to.x} y2={e.to.y} stroke="#94a3b8" strokeWidth="1.2" strokeDasharray="5 4" opacity="0.8" />
      ))}

      <g className="node-graph__core">
        <rect x={cx - 80} y={cy - 24} width="160" height="48" rx="12" />
        <text x={cx} y={cy + 5} textAnchor="middle">
          fire_station.accdb
        </text>
      </g>

      {nodes.map((n) => (
        <g key={n.name} className="node-graph__node" onClick={() => router.push(`/tables/${n.name}`)}>
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
    </svg>
  )
}
