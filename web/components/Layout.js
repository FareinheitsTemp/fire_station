import Link from 'next/link'
import { useRouter } from 'next/router'

const items = [
  ['/', 'Огляд'],
  ['/structure', 'Структура'],
  ['/tables', 'Таблиці'],
  ['/new-call', 'Новий виклик'],
  ['/reports', 'Звіти'],
  ['/ai', 'AI-асистент'],
  ['/settings', 'Налаштування'],
]

export default function Layout({ children }) {
  const { pathname } = useRouter()
  const isActive = (href) =>
    href === '/' ? pathname === '/' : pathname.startsWith(href)
  return (
    <div className="layout">
      <aside className="layout__sidebar">
        <div className="layout__logo">АІС «Пожежна частина»</div>
        <nav className="nav">
          {items.map(([href, label]) => (
            <Link
              key={href}
              href={href}
              className={`nav__link${isActive(href) ? ' nav__link--active' : ''}`}
            >
              {label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="layout__content">{children}</main>
    </div>
  )
}
