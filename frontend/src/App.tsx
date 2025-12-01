import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import ChartPage from './pages/ChartPage'
import NewsPage from './pages/NewsPage'
import AccountPage from './pages/AccountPage'

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/chart/:pair?" element={<ChartPage />} />
          <Route path="/news" element={<NewsPage />} />
          <Route path="/account" element={<AccountPage />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App

