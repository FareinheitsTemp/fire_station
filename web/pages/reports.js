import { useState } from 'react'
import { post } from '../lib/api'

function monthDefaults() {
  const now = new Date()
  const first = new Date(now.getFullYear(), now.getMonth(), 1)
  const fmt = (d) => d.toISOString().slice(0, 10)
  return [fmt(first), fmt(now)]
}

export default function Reports() {
  const [defFrom, defTo] = monthDefaults()
  const [from, setFrom] = useState(defFrom)
  const [to, setTo] = useState(defTo)
  const [file, setFile] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (e) => {
    e.preventDefault()
    setFile('')
    setError('')
    setBusy(true)
    try {
      const res = await post('/api/reports/calls', { from, to })
      setFile(res.file)
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">Звіт «Виклики за період» (PDF)</h1>
      {error && <div className="alert">{error}</div>}

      <form className="form" onSubmit={submit}>
        <label className="form__label">З дати</label>
        <input className="form__input" type="date" value={from} onChange={(e) => setFrom(e.target.value)} required />

        <label className="form__label">По дату</label>
        <input className="form__input" type="date" value={to} onChange={(e) => setTo(e.target.value)} required />

        <button className="btn" type="submit" disabled={busy}>
          {busy ? 'Генерую...' : 'Згенерувати PDF'}
        </button>
      </form>

      {file && (
        <div className="alert alert--ok">
          Готово:{' '}
          <a className="link" href={`/api/reports/file/${file}`} download>
            Скачати {file}
          </a>
        </div>
      )}
    </div>
  )
}
