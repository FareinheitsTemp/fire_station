import { useEffect, useRef, useState } from 'react'
import { post } from '../lib/api'
import Icon from './Icon'

// ChatWidget — плаваюче чат-вікно з агентом АІС (доступне на всіх сторінках).
export default function ChatWidget() {
  const [open, setOpen] = useState(false)
  const [msgs, setMsgs] = useState([])
  const [input, setInput] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const bodyRef = useRef(null)

  useEffect(() => {
    if (bodyRef.current) bodyRef.current.scrollTop = bodyRef.current.scrollHeight
  }, [msgs, busy])

  const send = async (e) => {
    e.preventDefault()
    const text = input.trim()
    if (!text || busy) return
    const next = [...msgs, { role: 'user', content: text }]
    setMsgs(next)
    setInput('')
    setBusy(true)
    setError('')
    try {
      const res = await post('/api/chat', { messages: next })
      setMsgs([...next, { role: 'assistant', content: res.reply }])
    } catch (err) {
      setError(err.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <button className="chat-fab" title="Чат з агентом" onClick={() => setOpen(!open)}>
        <Icon name="chat" size={20} />
      </button>
      {open && (
        <div className="chat-panel">
          <div className="chat-panel__head">
            <span>
              <Icon name="bot" size={15} /> Агент АІС
            </span>
            <button className="icon-btn" onClick={() => setOpen(false)}>
              ×
            </button>
          </div>
          <div className="chat-panel__body" ref={bodyRef}>
            {!msgs.length && (
              <p className="muted">
                Я в курсі стану частини: статистика, останні виклики, правила бази знань. Питай —
                відповім.
              </p>
            )}
            {msgs.map((m, i) => (
              <div key={i} className={`chat-msg chat-msg--${m.role}`}>
                {m.content}
              </div>
            ))}
            {busy && <div className="chat-msg chat-msg--assistant">Друкує…</div>}
            {error && <div className="alert">{error}</div>}
          </div>
          <form className="chat-panel__form" onSubmit={send}>
            <input
              className="form__input"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Повідомлення…"
            />
            <button className="btn" type="submit" disabled={busy} title="Надіслати">
              <Icon name="send" size={14} />
            </button>
          </form>
        </div>
      )}
    </>
  )
}
