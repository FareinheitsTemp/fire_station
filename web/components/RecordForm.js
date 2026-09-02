import { useEffect, useState } from 'react'
import { get } from '../lib/api'

// «02.01.2006 15:04» → «2006-01-02» для <input type="date">
function toDateInput(v) {
  if (!v) return ''
  const m = String(v).match(/(\d{2})\.(\d{2})\.(\d{4})/)
  if (m) return `${m[3]}-${m[2]}-${m[1]}`
  return String(v).slice(0, 10)
}

export default function RecordForm({ meta, initial, onSubmit, onCancel }) {
  const [values, setValues] = useState({})
  const [refs, setRefs] = useState({})

  useEffect(() => {
    const v = {}
    meta.columns.forEach((c) => {
      let val = initial ? initial[c.name] ?? '' : ''
      if (c.type === 'date') val = toDateInput(val)
      if (c.type === 'bool') val = val === '1' || val === '-1' || val === 'true' ? '1' : '0'
      v[c.name] = val
    })
    setValues(v)
    meta.columns
      .filter((c) => c.type === 'ref')
      .forEach((c) => {
        get(`/api/ref/${c.ref}`)
          .then((list) => setRefs((r) => ({ ...r, [c.name]: list })))
          .catch(() => {})
      })
  }, [meta, initial])

  const set = (k) => (e) => setValues({ ...values, [k]: e.target.value })

  const field = (c) => {
    switch (c.type) {
      case 'select':
        return (
          <select className="form__input" value={values[c.name] ?? ''} onChange={set(c.name)}>
            <option value="">—</option>
            {c.options.map((o) => (
              <option key={o} value={o}>{o}</option>
            ))}
          </select>
        )
      case 'ref':
        return (
          <select className="form__input" value={values[c.name] ?? ''} onChange={set(c.name)}>
            <option value="">—</option>
            {(refs[c.name] || []).map((r) => (
              <option key={r.id} value={r.id}>{r.label || `#${r.id}`}</option>
            ))}
          </select>
        )
      case 'bool':
        return (
          <select className="form__input" value={values[c.name] ?? '0'} onChange={set(c.name)}>
            <option value="1">Так</option>
            <option value="0">Ні</option>
          </select>
        )
      case 'number':
        return <input className="form__input" type="number" step="any" value={values[c.name] ?? ''} onChange={set(c.name)} />
      case 'date':
        return <input className="form__input" type="date" value={values[c.name] ?? ''} onChange={set(c.name)} />
      default:
        return <input className="form__input" value={values[c.name] ?? ''} onChange={set(c.name)} />
    }
  }

  return (
    <form
      className="form form--panel"
      onSubmit={(e) => {
        e.preventDefault()
        onSubmit(values)
      }}
    >
      {meta.columns.map((c) => (
        <div key={c.name}>
          <label className="form__label">
            {c.label}
            {c.required ? ' *' : ''}
          </label>
          {field(c)}
        </div>
      ))}
      <div className="form__actions">
        <button className="btn" type="submit">
          {initial ? 'Зберегти зміни' : 'Додати'}
        </button>
        <button className="btn btn--ghost" type="button" onClick={onCancel}>
          Скасувати
        </button>
      </div>
    </form>
  )
}
