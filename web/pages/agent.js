import { useState } from 'react'
import { post } from '../lib/api'
import Icon from '../components/Icon'

// Авто-мод: автономний агент аналізує стан системи й за потреби
// сам створює нові правила в базі знань.
export default function Agent() {
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState('')

  const run = async () => {
    setBusy(true)
    setError('')
    setResult(null)
    try {
      const res = await post('/api/agent/analyze', {})
      setResult(res)
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">
        <Icon name="bot" size={22} /> Агент — авто-мод
      </h1>
      <p className="muted">
        Агент автономно аналізує знімок системи: статистику, останні виклики, чинні правила бази знань
        та останні помилки API. Якщо щось не так або якогось правила бракує — сам створює нові
        правила в базі знань і формулює висновок.
      </p>

      {error && <div className="alert">{error}</div>}

      <div className="toolbar">
        <button className="btn" onClick={run} disabled={busy}>
          {busy ? 'Аналізую систему…' : '▶ Запустити автономний аналіз'}
        </button>
      </div>

      {result && (
        <>
          <div className="panel">
            <h2 className="page__subtitle" style={{ marginTop: 0 }}>
              Висновок агента
            </h2>
            <p>{result.conclusion}</p>
            <p className="muted">
              Нових правил створено: {result.rules_added} • усього правил у БЗ: {result.rules_total}
            </p>
          </div>

          {result.new_rules?.length > 0 && (
            <>
              <h2 className="page__subtitle">Нові правила (записані у базу знань)</h2>
              <div className="kb-grid">
                {result.new_rules.map((r, i) => (
                  <div key={i} className="kb-card">
                    <div className="kb-card__head">
                      <b>{r.topic}</b>
                      <span className="badge badge--warn">{r.priority}</span>
                    </div>
                    <p className="kb-card__line">
                      <span className="muted">Якщо:</span> {r.condition}
                    </p>
                    <p className="kb-card__line">
                      <span className="muted">То:</span> {r.recommendation}
                    </p>
                    <p className="muted">Категорія: {r.category}</p>
                  </div>
                ))}
              </div>
            </>
          )}
        </>
      )}
    </div>
  )
}
