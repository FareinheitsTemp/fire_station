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
      <div className="tbl-cards">
        {tables.map((t) => (
          <Link key={t.name} href={`/tables/${t.name}`} className="tbl-card">
            <span className="tbl-card__icon" style={{ background: t.color }}>
              {t.label[0]}
            </span>
            <span className="tbl-card__body">
              <span className="tbl-card__label">{t.label}</span>
              <span className="tbl-card__name">
                {t.name} • {t.columns.length} полів
              </span>
            </span>
            <span className="tbl-card__arrow">→</span>
          </Link>
        ))}
      </div>
    </div>
  )
}
