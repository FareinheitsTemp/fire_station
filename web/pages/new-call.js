import { useEffect, useState } from 'react'
import { get, post } from '../lib/api'

export default function NewCall() {
  const [types, setTypes] = useState([])
  const [form, setForm] = useState({
    address: '',
    district: '',
    fire_type_id: '',
    caller_name: '',
    caller_phone: '',
    description: '',
  })
  const [msg, setMsg] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    get('/api/fire-types').then(setTypes).catch(() => {})
  }, [])

  const set = (k) => (e) => setForm({ ...form, [k]: e.target.value })

  const submit = async (e) => {
    e.preventDefault()
    setMsg('')
    setError('')
    try {
      const res = await post('/api/calls', { ...form, fire_type_id: Number(form.fire_type_id) || 0 })
      setMsg(`Виклик №${res.id} зареєстровано`)
      setForm({ address: '', district: '', fire_type_id: '', caller_name: '', caller_phone: '', description: '' })
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div className="page">
      <h1 className="page__title">Новий виклик</h1>
      {msg && <div className="alert alert--ok">{msg}</div>}
      {error && <div className="alert">{error}</div>}

      <form className="form" onSubmit={submit}>
        <label className="form__label">Адреса виклику *</label>
        <input className="form__input" value={form.address} onChange={set('address')} required />

        <label className="form__label">Район</label>
        <input className="form__input" value={form.district} onChange={set('district')} />

        <label className="form__label">Тип пожежі</label>
        <select className="form__input" value={form.fire_type_id} onChange={set('fire_type_id')}>
          <option value="">— оберіть тип —</option>
          {types.map((t) => (
            <option key={t.ID} value={t.ID}>
              {t.Name}
            </option>
          ))}
        </select>

        <label className="form__label">Заявник (ПІБ)</label>
        <input className="form__input" value={form.caller_name} onChange={set('caller_name')} />

        <label className="form__label">Телефон заявника</label>
        <input className="form__input" value={form.caller_phone} onChange={set('caller_phone')} />

        <label className="form__label">Опис ситуації</label>
        <textarea className="form__input" rows="3" value={form.description} onChange={set('description')} />

        <button className="btn" type="submit">
          Зареєструвати виклик
        </button>
      </form>
    </div>
  )
}
