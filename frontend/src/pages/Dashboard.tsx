import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getTradingPairs, getCurrentPrice } from '../services/api'
import type { TradingPair } from '../services/api'
import './Dashboard.css'

export default function Dashboard() {
  const [pairs, setPairs] = useState<TradingPair[]>([])
  const [prices, setPrices] = useState<Record<string, number>>({})

  useEffect(() => {
    loadPairs()
  }, [])

  useEffect(() => {
    const interval = setInterval(() => {
      pairs.forEach(pair => {
        loadPrice(pair.symbol)
      })
    }, 5000)

    return () => clearInterval(interval)
  }, [pairs])

  const loadPairs = async () => {
    try {
      const data = await getTradingPairs()
      setPairs(data)
      // Load initial prices
      data.forEach(pair => loadPrice(pair.symbol))
    } catch (error) {
      console.error('Error loading pairs:', error)
    }
  }

  const loadPrice = async (pair: string) => {
    try {
      const data = await getCurrentPrice(pair)
      setPrices(prev => ({ ...prev, [pair]: parseFloat(data.price) }))
    } catch (error) {
      console.error(`Error loading price for ${pair}:`, error)
    }
  }

  return (
    <div className="dashboard">
      <h1>Dashboard</h1>
      <div className="pairs-grid">
        {pairs.map(pair => (
          <Link key={pair.id} to={`/chart/${pair.symbol}`} className="pair-card">
            <div className="pair-header">
              <h3>{pair.symbol}</h3>
              <span className="pair-assets">{pair.base_asset}/{pair.quote_asset}</span>
            </div>
            <div className="pair-price">
              {prices[pair.symbol] ? (
                <>${prices[pair.symbol].toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</>
              ) : (
                <>Loading...</>
              )}
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}

