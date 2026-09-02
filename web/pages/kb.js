import { useEffect, useState } from 'react'
import { del, get, post, put } from '../lib/api'
import RecordForm from '../components/RecordForm'
import Icon from '../components/Icon'

const prioClass = (p) => (p === 'високий' ? 'badge badge--err' : p === 'середній' ? 'badge badge--warn' : 'badge')

// База знань: правила реагування з таблиці kb_rules.
export default function Knowledge() {
  const [meta, setMeta] = useState(null)
  const [cols, setCols] = useState([])
  const [rows, setRows] = useState([])
  const [mode, setMode] = useState(null)
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')

  const load = () => {
    get('/api/meta')
      .then((list) => setMeta(list.find((t) => t.name === 'kb_rules') || null))
      .catch((e) => setError(e.message))
    get('/api/tables/kb_rules')
      .then((d) => {
        setCols(d.columns)
        setRows(d.rows)
      })
      .catch((e) => setError(e.message))
  }
  useEffect(load, [])

  const obj = (row) => Object.fromEntries(cols.map((c, i) => [c, row[i]]))

  const submit = async (values) => {
    setError('')
    setMsg('')
    try {
      if (mode === 'new') {
        await post('/api/tables/kb_rules/rows', { values })
        setMsg('Правило додано')
      } else {
        await put(`/api/tables/kb_rules/rows/${mode.id}`, { values })
        setMsg('Правило оновлено')
      }
      setMode(null)
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  const remove = async (id) => {
    if (!confirm(`Видалити правило #${id}?`)) return
    try {
      await del(`/api/tables/kb_rules/rows/${id}`)
      setMsg('Правило видалено')
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  const byCat = {}
  rows.forEach((r) => {
    const o = obj(r)
    const c = o.category || 'інше'
    if (!byCat[c]) byCat[c] = []
    byCat[c].push(o)
  })

  return (
    <div className="page">
      <h1 className="page__title">База знань</h1>
      <p className="muted">
        Правила реагування виду «якщо … → то …». AI-асистент враховує їх у своїх відповідях.
      </p>
      {msg && <div className="alert alert--ok">{msg}</div>}
      {error && <div className="alert">{error}</div>}

      {meta && (
        <div className="toolbar">
          <button className="btn" onClick={() => setMode('new')}>
            + Додати правило
          </button>
        </div>
      )}

      {Object.entries(byCat).map(([cat, rules]) => (
        <div key={cat}>
          <h2 className="page__subtitle">{cat}</h2>
          <div className="kb-grid">
            {rules.map((r) => (
              <div key={r.id} className="kb-card">
                <div className="kb-card__head">
                  <b>{r.topic}</b>
                  <span className={prioClass(r.priority)}>{r.priority || '—'}</span>
                </div>
                <p className="kb-card__line">
                  <span className="muted">Якщо:</span> {r.condition_text}
                </p>
                <p className="kb-card__line">
                  <span className="muted">То:</span> {r.recommendation}
                </p>
                <div className="kb-card__actions">
                  <button className="icon-btn" title="Редагувати" onClick={() => setMode(r)}>
                    <Icon name="edit" size={13} />
                  </button>{' '}
                  <button className="icon-btn icon-btn--danger" title="Видалити" onClick={() => remove(r.id)}>
                    <Icon name="trash" size={13} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}

      {meta && mode && (
        <RecordForm
          meta={meta}
          initial={mode === 'new' ? null : mode}
          onSubmit={submit}
          onCancel={() => setMode(null)}
        />
      )}
    </div>
  )
}
