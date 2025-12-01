import { useEffect, useRef, useState } from 'react'
import { createChart, IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts'
import { useWebSocket } from '../hooks/useWebSocket'
import { getKlines } from '../services/api'
import './TradingChart.css'

interface TradingChartProps {
  pair: string
  interval?: string
}

export default function TradingChart({ pair, interval = '1h' }: TradingChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const [currentPrice, setCurrentPrice] = useState<number | null>(null)

  // WebSocket connection for realtime price updates
  const wsData = useWebSocket(`/ws/price/${pair}`)

  useEffect(() => {
    if (!chartContainerRef.current) return

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
      layout: {
        background: { color: '#0a0e27' },
        textColor: '#d1d5db',
      },
      grid: {
        vertLines: { color: '#1f2937' },
        horzLines: { color: '#1f2937' },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
      },
    })

    chartRef.current = chart

    // Add candlestick series
    const candlestickSeries = chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    })

    candlestickSeriesRef.current = candlestickSeries

    // Load historical data
    loadHistoricalData(pair, interval)

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current && chart) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        })
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [pair, interval])

  useEffect(() => {
    if (wsData && candlestickSeriesRef.current) {
      // Update chart with realtime data
      const price = (wsData as any).price
      if (price) {
        setCurrentPrice(parseFloat(price))
        // Update last candle or create new one based on time
        // This is simplified - in production, you'd update the last candle
      }
    }
  }, [wsData])

  const loadHistoricalData = async (symbol: string, inter: string) => {
    try {
      const klines = await getKlines(symbol, inter, 500)
      
      const formattedData: CandlestickData<Time>[] = klines.map((kline: any) => ({
        time: (kline[0] / 1000) as Time, // Convert milliseconds to seconds
        open: parseFloat(kline[1]),
        high: parseFloat(kline[2]),
        low: parseFloat(kline[3]),
        close: parseFloat(kline[4]),
      }))

      candlestickSeriesRef.current?.setData(formattedData)
    } catch (error) {
      console.error('Error loading historical data:', error)
    }
  }

  return (
    <div className="trading-chart-container">
      <div className="chart-header">
        <h2>{pair}</h2>
        {currentPrice && (
          <div className="current-price">
            ${currentPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
        )}
      </div>
      <div ref={chartContainerRef} className="chart-wrapper" />
    </div>
  )
}

