import Link from 'next/link'
import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import { get } from '../lib/api'
import Icon from './Icon'

const SERVICES = [
  ['/new-call', 'Новий виклик', 'call'],
  ['/reports', 'Звіти', 'report'],
  ['/ai', 'AI-асистент', 'ai'],
  ['/agent', 'Агент (авто-мод)', 'bot'],
  ['/settings', 'Налаштування', 'settings'],
]

const GROUPS = [
  { key: 'core', title: 'Ядро викликів' },
  { key: 'staff', title: 'Персонал' },
  { key: 'equipment', title: 'Техніка' },
  { key: 'refs', title: 'Довідники' },
]

export default function Sidebar() {
  const { pathname } = useRouter()
  const [tables, setTables] = useState([])
  const [open, setOpen] = useState({ core: true, staff: true, equipment: true, refs: true })

  useEffect(() => {
    get('/api/meta').then(setTables).catch(() => {})
  }, [])

  const isActive = (href) => (href === '/' ? pathname === '/' : pathname.startsWith(href))
  const toggle = (k) => setOpen((o) => ({ ...o, [k]: !o[k] }))

  return (
    <aside className="sidebar">
      <Link href="/" className="sidebar__logo">
        <Icon name="flame" size={17} /> Пожежна частина
      </Link>

      <nav className="sidebar__nav">
        <div className="sgroup">
          <div className="sgroup__title">Дані</div>
          <Link href="/" className={`slink${isActive('/') ? ' slink--active' : ''}`}>
            <Icon name="dashboard" /> Огляд
          </Link>
          <Link href="/structure" className={`slink${isActive('/structure') ? ' slink--active' : ''}`}>
            <Icon name="graph" /> Структура
          </Link>
          <Link href="/tables" className={`slink${pathname === '/tables' ? ' slink--active' : ''}`}>
            <Icon name="table" /> Усі таблиці
          </Link>
          <Link href="/kb" className={`slink${isActive('/kb') ? ' slink--active' : ''}`}>
            <Icon name="kb" /> База знань
          </Link>
        </div>

        {GROUPS.map((g) => {
          const items = tables.filter((t) => t.category === g.key)
          if (!items.length) return null
          return (
            <div className="sgroup" key={g.key}>
              <button className="sgroup__title sgroup__title--btn" onClick={() => toggle(g.key)}>
                <span>{open[g.key] ? '▾' : '▸'}</span> {g.title}
              </button>
              {open[g.key] &&
                items.map((t) => (
                  <Link
                    key={t.name}
                    href={`/tables/${t.name}`}
                    className={`slink${pathname === `/tables/${t.name}` ? ' slink--active' : ''}`}
                  >
                    <span className="slink__dot" style={{ background: t.color }} />
                    {t.label}
                  </Link>
                ))}
            </div>
          )
        })}

        <div className="sgroup">
          <div className="sgroup__title">Сервіси</div>
          {SERVICES.map(([href, label, icon]) => (
            <Link key={href} href={href} className={`slink${isActive(href) ? ' slink--active' : ''}`}>
              <Icon name={icon} /> {label}
            </Link>
          ))}
        </div>
      </nav>
    </aside>
  )
}
