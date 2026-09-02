import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import { del, get, post, put } from '../../lib/api'
import DataTable from '../../components/DataTable'
import RecordForm from '../../components/RecordForm'

export default function TableView() {
  const router = useRouter()
  const { name } = router.query
  const [meta, setMeta] = useState(null)
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  const [mode, setMode] = useState(null) // null | 'new' | {row object}

  const load = () => {
    if (!name) return
    get('/api/meta')
      .then((list) => setMeta(list.find((t) => t.name === name) || null))
      .catch((e) => setError(e.message))
    get(`/api/tables/${name}`)
      .then(setData)
      .catch((e) => setError(e.message))
  }
  useEffect(load, [name])

  if (!name) return null

  const pkIdx = meta && data ? data.columns.indexOf(meta.pk) : -1
  const rowToObj = (row) => Object.fromEntries(data.columns.map((c, i) => [c, row[i]]))

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

      {meta && !mode && (
        <div className="toolbar">
          <button className="btn" onClick={() => setMode('new')}>
            + Додати запис
          </button>
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
        <DataTable
          columns={data.columns}
          rows={data.rows}
          actions={
            meta
              ? (row) => (
                  <>
                    <button className="icon-btn" title="Редагувати" onClick={() => setMode(rowToObj(row))}>
                      ✏️
                    </button>{' '}
                    <button className="icon-btn icon-btn--danger" title="Видалити" onClick={() => remove(row)}>
                      🗑
                    </button>
                  </>
                )
              : null
          }
        />
      )}
    </div>
  )
}
