import { useEffect, useState, useRef } from 'react'

export function useWebSocket(url: string) {
  const [data, setData] = useState<any>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number>()

  useEffect(() => {
    const connect = () => {
      try {
        const wsUrl = url.startsWith('ws') ? url : `ws://localhost:8080${url}`
        const ws = new WebSocket(wsUrl)

        ws.onopen = () => {
          console.log('WebSocket connected:', url)
        }

        ws.onmessage = (event) => {
          try {
            const parsed = JSON.parse(event.data)
            setData(parsed)
          } catch (e) {
            console.error('Error parsing WebSocket message:', e)
          }
        }

        ws.onerror = (error) => {
          console.error('WebSocket error:', error)
        }

        ws.onclose = () => {
          console.log('WebSocket closed, reconnecting...')
          reconnectTimeoutRef.current = window.setTimeout(connect, 3000)
        }

        wsRef.current = ws
      } catch (error) {
        console.error('Error connecting WebSocket:', error)
        reconnectTimeoutRef.current = window.setTimeout(connect, 3000)
      }
    }

    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [url])

  return data
}

