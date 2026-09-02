import StatusBadge from './StatusBadge'
import Icon from './Icon'

const STATUS = new Set(['новий', 'в роботі', 'завершений', 'в строю', 'ремонт', 'списано', 'справно', 'несправно'])

// TableGrid — таблиця в стилі Supabase: липкий заголовок, zebra,
// PK-чип з копіюванням, FK-чипи кольору цільової гілки, статуси-бейджі.
export default function TableGrid({ meta, refColors = {}, columns, rows, onEdit, onDelete }) {
  const cm = Object.fromEntries((meta?.columns || []).map((c) => [c.name, c]))

  const cell = (col, val) => {
    const c = cm[col]
    if (meta && col === meta.pk) {
      return (
        <button className="chip chip--pk" title="Скопіювати ID" onClick={() => navigator.clipboard?.writeText(String(val))}>
          {val} <Icon name="copy" size={11} />
        </button>
      )
    }
    if (val === '' || val == null) return <span className="null">—</span>
    if (c?.type === 'ref') {
      return (
        <span className="chip chip--ref" style={{ borderColor: refColors[c.ref], color: refColors[c.ref] }}>
          {val}
        </span>
      )
    }
    if (STATUS.has(val)) return <StatusBadge status={val} />
    return String(val)
  }

  if (!rows.length) return <p className="muted">(порожньо)</p>

  return (
    <div className="grid-wrap">
      <table className="grid">
        <thead>
          <tr>
            {columns.map((c) => (
              <th key={c}>
                {cm[c]?.label || c}
                <span className="grid__colname">{c}</span>
              </th>
            ))}
            {(onEdit || onDelete) && <th className="grid__actcol">Дії</th>}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i}>
              {row.map((v, j) => (
                <td key={j}>{cell(columns[j], v)}</td>
              ))}
              {(onEdit || onDelete) && (
                <td className="grid__actcol">
                  {onEdit && (
                    <button className="icon-btn" title="Редагувати" onClick={() => onEdit(row)}>
                      <Icon name="edit" size={13} />
                    </button>
                  )}{' '}
                  {onDelete && (
                    <button className="icon-btn icon-btn--danger" title="Видалити" onClick={() => onDelete(row)}>
                      <Icon name="trash" size={13} />
                    </button>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
