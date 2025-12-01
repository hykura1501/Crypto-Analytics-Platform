import { ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
import './Layout.css'

interface LayoutProps {
  children: ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation()

  return (
    <div className="layout">
      <nav className="navbar">
        <div className="navbar-brand">
          <h1>📊 Crypto Analytics</h1>
        </div>
        <div className="navbar-menu">
          <Link 
            to="/" 
            className={location.pathname === '/' ? 'active' : ''}
          >
            Dashboard
          </Link>
          <Link 
            to="/chart" 
            className={location.pathname.startsWith('/chart') ? 'active' : ''}
          >
            Charts
          </Link>
          <Link 
            to="/news" 
            className={location.pathname === '/news' ? 'active' : ''}
          >
            News
          </Link>
          <Link 
            to="/account" 
            className={location.pathname === '/account' ? 'active' : ''}
          >
            Account
          </Link>
        </div>
      </nav>
      <main className="main-content">
        {children}
      </main>
    </div>
  )
}

