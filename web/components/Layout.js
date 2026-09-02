import Sidebar from './Sidebar'
import Topbar from './Topbar'

export default function Layout({ children }) {
  return (
    <div className="layout">
      <Sidebar />
      <div className="layout__main">
        <Topbar />
        <main className="layout__content">{children}</main>
      </div>
    </div>
  )
}
