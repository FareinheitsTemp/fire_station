import { useState } from 'react'
import { post } from '../lib/api'
import Icon from '../components/Icon'

// Самотест системи: 7 перевірок зв'язки БД + БЗ + ШІ.
export default function SelfTest() {
  const [busy, setBusy] = useState(false)
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  const run = async () => {
    setBusy(true)
    setError('')
    setData(null)
    try {
      setData(await post('/api/selftest', {}))
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">
        <Icon name="flask" size={22} /> Тестування системи
      </h1>
      <p className="muted">
        Проганяє перевірки всієї зв'язки: підключення до БД, читання кожної таблиці, повний
        CRUD-цикл, запис/читання розкладки мапи, наявність шрифту для PDF, ключ/endpoint ШІ і
        реальну відповідь моделі.
      </p>

      {error && <div className="alert">{error}</div>}

      <div className="toolbar">
        <button className="btn" onClick={run} disabled={busy}>
          {busy ? 'Тестую…' : '▶ Запустити самотест'}
        </button>
      </div>

      {data && (
        <>
          <div className={`alert ${data.ok ? 'alert--ok' : ''}`}>
            {data.ok ? 'Усі перевірки пройдено' : 'Є провалені перевірки — деталі нижче'}
          </div>
          <div className="grid-wrap">
            <table className="grid">
              <thead>
                <tr>
                  <th>Перевірка</th>
                  <th>Статус</th>
                  <th>Деталь</th>
                  <th>Час</th>
                </tr>
              </thead>
              <tbody>
                {data.results.map((r, i) => (
                  <tr key={i}>
                    <td>{r.name}</td>
                    <td>
                      <span className={`badge ${r.ok ? 'badge--ok' : 'badge--err'}`}>
                        {r.ok ? 'пройдено' : 'провалено'}
                      </span>
                    </td>
                    <td style={{ whiteSpace: 'normal' }}>{r.detail}</td>
                    <td className="muted">{r.ms} мс</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
