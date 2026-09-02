import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import { get } from '../lib/api'

const NAMES = {
  tables: 'Таблиці',
  structure: 'Структура',
  'new-call': 'Новий виклик',
  reports: 'Звіти',
  ai: 'AI-асистент',
  settings: 'Налаштування',
}

export default function Topbar() {
  const router = useRouter()
  const [dbOK, setDbOK] = useState(null)

  useEffect(() => {
    const check = () =>
      get('/api/health')
        .then((h) => setDbOK(!!h.db))
        .catch(() => setDbOK(false))
    check()
    const t = setInterval(check, 15000)
    return () => clearInterval(t)
  }, [])

  const segments = router.pathname.split('/').filter(Boolean)
  const crumbs = ['АІС']
  segments.forEach((s, i) => {
    crumbs.push(i === segments.length - 1 ? NAMES[s] || router.query.name || s : NAMES[s] || s)
  })

  return (
    <header className="topbar">
      <div className="topbar__crumbs">
        {crumbs.map((c, i) => (
          <span key={i}>
            {i > 0 && <span className="topbar__sep">/</span>}
            <span className={i === crumbs.length - 1 ? 'topbar__cur' : ''}>{c}</span>
          </span>
        ))}
      </div>
      <div className="topbar__right">
        <span className={`dbstatus${dbOK === true ? ' dbstatus--ok' : ''}${dbOK === false ? ' dbstatus--err' : ''}`}>
          <span className="dbstatus__dot" />
          {dbOK === null ? 'перевірка…' : dbOK ? 'БД підключено' : 'БД недоступна'}
        </span>
        <button className="icon-btn" title="Оновити сторінку" onClick={() => router.reload()}>
          ⟳
        </button>
      </div>
    </header>
  )
}
