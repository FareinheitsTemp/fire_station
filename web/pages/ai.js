import { useState } from 'react'
import { post } from '../lib/api'
import DataTable from '../components/DataTable'

export default function AI() {
  const [question, setQuestion] = useState('')
  const [result, setResult] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setResult(null)
    setError('')
    setBusy(true)
    try {
      const res = await post('/api/ai', { question })
      setResult(res)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">AI-асистент</h1>
      <p className="muted">Запит українською → SQL (Access) → таблиця результатів. Лише читання (SELECT).</p>
      {error && <div className="alert">{error}</div>}

      <form className="form" onSubmit={submit}>
        <label className="form__label">Питання</label>
        <input
          className="form__input"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          placeholder="скільки викликів по районах цього місяця"
          required
        />
        <button className="btn" type="submit" disabled={busy}>
          {busy ? 'Думаю...' : 'Запитати'}
        </button>
      </form>

      {result && (
        <>
          <h2 className="page__subtitle">Згенерований SQL</h2>
          <pre className="code">{result.sql}</pre>
          <h2 className="page__subtitle">Результат</h2>
          <DataTable columns={result.columns} rows={result.rows} />
        </>
      )}
    </div>
  )
}
