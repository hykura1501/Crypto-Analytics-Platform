import axios from 'axios'

const API_BASE_URL = '/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

export interface TradingPair {
  id: number
  symbol: string
  base_asset: string
  quote_asset: string
  status: string
}

export interface News {
  id: number
  source_id: number
  title: string
  content: string
  url: string
  author: string
  published_at: string
  sentiment_score: number | null
  sentiment_label: string | null
}

export interface PriceHistory {
  id: number
  pair_id: number
  price: number
  volume: number
  high: number
  low: number
  timestamp: string
  interval: string
}

// Trading Pairs
export const getTradingPairs = async (): Promise<TradingPair[]> => {
  const response = await api.get('/pairs')
  return response.data.pairs
}

export const getPairInfo = async (pair: string): Promise<TradingPair> => {
  const response = await api.get(`/pairs/${pair}`)
  return response.data
}

// Price Data
export const getCurrentPrice = async (pair: string) => {
  const response = await api.get(`/price/${pair}`)
  return response.data
}

export const getPriceHistory = async (pair: string, interval: string = '1h', limit: number = 100): Promise<PriceHistory[]> => {
  const response = await api.get(`/price/${pair}/history`, {
    params: { interval, limit }
  })
  return response.data.history
}

export const getKlines = async (pair: string, interval: string = '1h', limit: number = 500): Promise<any[]> => {
  const response = await api.get(`/klines/${pair}`, {
    params: { interval, limit }
  })
  return response.data.klines || []
}

// News
export const getNews = async (limit: number = 20): Promise<News[]> => {
  const response = await api.get('/news', {
    params: { limit }
  })
  return response.data.news
}

export const getNewsDetail = async (id: number): Promise<News> => {
  const response = await api.get(`/news/${id}`)
  return response.data
}

// AI Analysis
export const getAnalysis = async (pair: string) => {
  const response = await api.get(`/analysis/${pair}`)
  return response.data
}

export const predictTrend = async (pair: string, timeHorizon: string = '24h') => {
  const response = await api.post('/analysis/predict', {
    pair,
    time_horizon: timeHorizon
  })
  return response.data
}

export default api

