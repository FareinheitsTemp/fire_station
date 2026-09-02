import Link from 'next/link'
import { useEffect, useState } from 'react'
import { get } from '../lib/api'

export default function Tables() {
  const [tables, setTables] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    get('/api/meta').then(setTables).catch((e) => setError(e.message))
  }, [])

  return (
    <div className="page">
      <h1 className="page__title">Таблиці бази даних</h1>
      {error && <div className="alert">{error}</div>}
      <div className="table-list">
        {tables.map((t) => (
          <Link key={t.name} href={`/tables/${t.name}`} className="table-list__item">
            <span className="dot" style={{ background: t.color }} />
            {t.label} <span className="muted">({t.name})</span>
          </Link>
        ))}
      </div>
    </div>
  )
}
