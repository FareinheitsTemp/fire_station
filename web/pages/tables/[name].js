import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import { del, get, post, put } from '../../lib/api'
import TableGrid from '../../components/TableGrid'
import RecordForm from '../../components/RecordForm'

export default function TableView() {
  const router = useRouter()
  const { name } = router.query
  const [meta, setMeta] = useState(null)
  const [refColors, setRefColors] = useState({})
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  const [mode, setMode] = useState(null) // null | 'new' | {row object}
  const [q, setQ] = useState('')

  const load = () => {
    if (!name) return
    get('/api/meta')
      .then((list) => {
        setMeta(list.find((t) => t.name === name) || null)
        setRefColors(Object.fromEntries(list.map((t) => [t.name, t.color])))
      })
      .catch((e) => setError(e.message))
    get(`/api/tables/${name}`)
      .then(setData)
      .catch((e) => setError(e.message))
  }
  useEffect(load, [name])

  if (!name) return null

  const pkIdx = meta && data ? data.columns.indexOf(meta.pk) : -1
  const rowToObj = (row) => Object.fromEntries(data.columns.map((c, i) => [c, row[i]]))
  const filtered = data
    ? data.rows.filter((r) => !q || r.some((cell) => String(cell).toLowerCase().includes(q.toLowerCase())))
    : []

  const submit = async (values) => {
    setError('')
    setMsg('')
    try {
      if (mode === 'new') {
        await post(`/api/tables/${name}/rows`, { values })
        setMsg('Запис додано')
      } else {
        await put(`/api/tables/${name}/rows/${mode[meta.pk]}`, { values })
        setMsg('Запис оновлено')
      }
      setMode(null)
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  const remove = async (row) => {
    if (pkIdx < 0) {
      setError('Не знайдено первинний ключ таблиці')
      return
    }
    const id = row[pkIdx]
    if (!confirm(`Видалити запис #${id}?`)) return
    setError('')
    setMsg('')
    try {
      await del(`/api/tables/${name}/rows/${id}`)
      setMsg('Запис видалено')
      load()
    } catch (e) {
      setError(e.message)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">
        {meta ? meta.label : name} <span className="muted">({name})</span>
      </h1>
      {msg && <div className="alert alert--ok">{msg}</div>}
      {error && <div className="alert">{error}</div>}

      {meta && (
        <div className="toolbar">
          <button className="btn" onClick={() => setMode('new')}>
            + Додати запис
          </button>
          <input
            className="form__input toolbar__search"
            placeholder="Пошук по таблиці…"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
          <span className="muted toolbar__count">
            Записів: {filtered.length}
            {q && data ? ` з ${data.rows.length}` : ''}
          </span>
        </div>
      )}

      {meta && mode && (
        <RecordForm
          meta={meta}
          initial={mode === 'new' ? null : mode}
          onSubmit={submit}
          onCancel={() => setMode(null)}
        />
      )}

      {data && (
        <TableGrid
          meta={meta}
          refColors={refColors}
          columns={data.columns}
          rows={filtered}
          onEdit={meta ? (row) => setMode(rowToObj(row)) : null}
          onDelete={meta ? remove : null}
        />
      )}
    </div>
  )
}
