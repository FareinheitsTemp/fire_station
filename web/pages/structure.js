import { useEffect, useState } from 'react'
import { get } from '../lib/api'
import NodeGraph from '../components/NodeGraph'

export default function Structure() {
  const [tables, setTables] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    get('/api/meta').then(setTables).catch((e) => setError(e.message))
  }, [])

  return (
    <div className="page">
      <h1 className="page__title">Структура бази даних</h1>
      <p className="muted">
        Кожна таблиця — нода, підключена до спільного ядра (файл БД). Суцільні кольорові лінії —
        гілки ядра (колір відповідає ноді), пунктирні — зв'язки між таблицями (FK). Клік по ноді
        відкриває таблицю.
      </p>
      {error && <div className="alert">{error}</div>}
      {tables && <NodeGraph tables={tables} />}
    </div>
  )
}
