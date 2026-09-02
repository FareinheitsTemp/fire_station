import Link from 'next/link'
import { useEffect, useState } from 'react'
import { get } from '../lib/api'

export default function Tables() {
  const [names, setNames] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    get('/api/tables').then(setNames).catch((e) => setError(e.message))
  }, [])

  return (
    <div className="page">
      <h1 className="page__title">Таблиці бази даних</h1>
      {error && <div className="alert">{error}</div>}
      <div className="table-list">
        {names.map((n) => (
          <Link key={n} href={`/tables/${n}`} className="table-list__item">
            {n}
          </Link>
        ))}
      </div>
    </div>
  )
}
