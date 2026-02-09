import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import TradingChart from '../components/TradingChart'
import { getTradingPairs } from '../services/api'
import type { TradingPair } from '../services/api'
import './ChartPage.css'

export default function ChartPage() {
  const { pair } = useParams<{ pair?: string }>()
  const [selectedPair, setSelectedPair] = useState<string>(pair || 'BTCUSDT')
  const [availablePairs, setAvailablePairs] = useState<TradingPair[]>([])
  const [interval, setInterval] = useState<string>('1h')

  useEffect(() => {
    loadPairs()
  }, [])

  useEffect(() => {
    if (pair) {
      setSelectedPair(pair)
    }
  }, [pair])

  const loadPairs = async () => {
    try {
      const data = await getTradingPairs()
      setAvailablePairs(data)
    } catch (error) {
      console.error('Error loading pairs:', error)
    }
  }

  return (
    <div className="chart-page">
      <div className="chart-controls">
        <div className="pair-selector">
          <label>Select Pair:</label>
          <select 
            value={selectedPair} 
            onChange={(e) => setSelectedPair(e.target.value)}
          >
            {availablePairs.map(p => (
              <option key={p.id} value={p.symbol}>{p.symbol}</option>
            ))}
          </select>
        </div>
        <div className="interval-selector">
          <label>Interval:</label>
          <select value={interval} onChange={(e) => setInterval(e.target.value)}>
            <option value="1m">1 Minute</option>
            <option value="5m">5 Minutes</option>
            <option value="15m">15 Minutes</option>
            <option value="1h">1 Hour</option>
            <option value="4h">4 Hours</option>
            <option value="1d">1 Day</option>
          </select>
        </div>
      </div>
      <div className="chart-container">
        <TradingChart pair={selectedPair} interval={interval} />
      </div>
    </div>
  )
}

