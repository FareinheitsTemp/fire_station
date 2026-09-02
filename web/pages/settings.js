import { useEffect, useState } from 'react'
import { get, put } from '../lib/api'

export default function Settings() {
  const [form, setForm] = useState({ db_path: '', font_path: '', ai_model: '', ai_key: '', ai_base_url: '' })
  const [hasKey, setHasKey] = useState(false)
  const [msg, setMsg] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    get('/api/config')
      .then((c) => {
        setForm({
          db_path: c.db_path,
          font_path: c.font_path,
          ai_model: c.ai_model,
          ai_base_url: c.ai_base_url || '',
          ai_key: '',
        })
        setHasKey(c.has_ai_key)
      })
      .catch((e) => setError(e.message))
  }, [])

  const set = (k) => (e) => setForm({ ...form, [k]: e.target.value })

  const submit = async (e) => {
    e.preventDefault()
    setMsg('')
    setError('')
    try {
      await put('/api/config', form)
      if (form.ai_key) setHasKey(true)
      setMsg('Налаштування збережено (якщо змінено шлях БД — перезапусти сервер)')
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">Налаштування</h1>
      {msg && <div className="alert alert--ok">{msg}</div>}
      {error && <div className="alert">{error}</div>}

      <form className="form" onSubmit={submit}>
        <label className="form__label">Шлях до БД (.accdb)</label>
        <input className="form__input" value={form.db_path} onChange={set('db_path')} />

        <label className="form__label">Шрифт для PDF (TTF)</label>
        <input className="form__input" value={form.font_path} onChange={set('font_path')} />

        <label className="form__label">API ключ Groq</label>
        <input
          className="form__input"
          type="password"
          value={form.ai_key}
          onChange={set('ai_key')}
          placeholder={hasKey ? '•••••• (збережено; порожньо — лишити)' : 'gsk_…'}
        />

        <label className="form__label">Base URL API</label>
        <input
          className="form__input"
          value={form.ai_base_url}
          onChange={set('ai_base_url')}
          placeholder="https://api.groq.com/openai/v1"
        />
        <p className="muted" style={{ margin: '2px 0 0', fontSize: 12.5 }}>
          Порожньо — дефолт Groq (https://api.groq.com/openai/v1). Для локальної моделі через Ollama:
          http://localhost:11434/v1
        </p>

        <label className="form__label">Модель ШІ</label>
        <input
          className="form__input"
          value={form.ai_model}
          onChange={set('ai_model')}
          placeholder="llama-3.3-70b-versatile"
        />
        <p className="muted" style={{ margin: '2px 0 0', fontSize: 12.5 }}>
          Production-моделі Groq: llama-3.3-70b-versatile · llama-3.1-8b-instant · openai/gpt-oss-120b ·
          openai/gpt-oss-20b
        </p>

        <button className="btn" type="submit">
          Зберегти
        </button>
      </form>
    </div>
  )
}
