import { useEffect, useState } from 'react'
import { get } from '../lib/api'
import StatusBadge from '../components/StatusBadge'

export default function Dashboard() {
  const [stats, setStats] = useState(null)
  const [recent, setRecent] = useState([])
  const [error, setError] = useState('')

  const load = () => {
    setError('')
    get('/api/health')
      .then((h) => {
        if (!h.db) setError(h.error || 'БД недоступна')
      })
      .catch(() => setError('API-сервер не відповідає — запусти fire-station.exe'))
    get('/api/stats').then(setStats).catch(() => {})
    get('/api/recent').then(setRecent).catch(() => {})
  }

  useEffect(load, [])

  return (
    <div className="page">
      <h1 className="page__title">Огляд</h1>

      {error && <div className="alert">{error}</div>}

      {stats && (
        <div className="cards">
          <div className="card">
            <div className="card__title">Викликів усього</div>
            <div className="card__value">{stats.TotalCalls}</div>
          </div>
          <div className="card">
            <div className="card__title">Викликів сьогодні</div>
            <div className="card__value">{stats.TodayCalls}</div>
          </div>
          <div className="card">
            <div className="card__title">Працівників активно</div>
            <div className="card__value">{stats.ActiveEmployees}</div>
          </div>
          <div className="card card--ok">
            <div className="card__title">Техніки в строю</div>
            <div className="card__value">{stats.EquipmentOK}</div>
          </div>
        </div>
      )}

      <h2 className="page__subtitle">Останні виклики</h2>
      {!recent.length && <p className="muted">(поки порожньо)</p>}
      <ul className="recent">
        {recent.map((rc, i) => (
          <li key={i} className="recent__item">
            <span className="recent__date">
              {new Date(rc.CallAt).toLocaleString('uk-UA', {
                day: '2-digit',
                month: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
              })}
            </span>
            <span className="recent__address">{rc.Address}</span>
            <StatusBadge status={rc.Status} />
          </li>
        ))}
      </ul>

      <button className="btn btn--ghost" onClick={load}>
        Оновити
      </button>
    </div>
  )
}
