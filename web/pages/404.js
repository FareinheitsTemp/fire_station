import Link from 'next/link'

export default function NotFound() {
  return (
    <div className="notfound">
      <div className="notfound__code">404</div>
      <h1 className="notfound__title">Такої сторінки немає</h1>
      <p className="muted">Схоже, ця гілка не підключена до жодної ноди.</p>
      <Link href="/" className="btn">
        ← На сторінку огляду
      </Link>
    </div>
  )
}
