import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
} from "react";
import { WS_BASE_URL } from "../config";
import type { MarketPrice } from "../types";

interface WebSocketContextType {
  subscribe: (topic: string) => void;
  unsubscribe: (topic: string) => void;
  lastMessage: MarketPrice | null;
  getLastMessage: (topic: string) => MarketPrice | null;
  isConnected: boolean;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

// eslint-disable-next-line react-refresh/only-export-components
export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<MarketPrice | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const subscriptionsRef = useRef<Set<string>>(new Set());
  const messagesRef = useRef<Map<string, MarketPrice>>(new Map());

  useEffect(() => {
    const connect = () => {
      const wsUrl = `${WS_BASE_URL}/ws/prices`;
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log("WebSocket connected");
        setIsConnected(true);

        // Resubscribe to existing topics on reconnection
        if (subscriptionsRef.current.size > 0) {
          const topics = Array.from(subscriptionsRef.current);
          ws.send(
            JSON.stringify({
              action: "subscribe",
              topics: topics,
            })
          );
        }
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          // Store message by topic
          const topic = `market:${data.symbol}:${data.interval}`;
          messagesRef.current.set(topic, data);
          setLastMessage(data);
        } catch (err) {
          console.error("Error parsing WebSocket message:", err);
        }
      };

      ws.onclose = () => {
        console.log("WebSocket disconnected");
        setIsConnected(false);
        // Attempt to reconnect after 5 seconds
        setTimeout(connect, 5000);
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        ws.close();
      };

      wsRef.current = ws;
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const subscribe = useCallback((topic: string) => {
    subscriptionsRef.current.add(topic);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          action: "subscribe",
          topics: [topic],
        })
      );
    }
  }, []);

  const unsubscribe = useCallback((topic: string) => {
    subscriptionsRef.current.delete(topic);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          action: "unsubscribe",
          topics: [topic],
        })
      );
    }
  }, []);

  const getLastMessage = useCallback((topic: string) => {
    return messagesRef.current.get(topic) || null;
  }, []);

  return (
    <WebSocketContext.Provider
      value={{ subscribe, unsubscribe, lastMessage, getLastMessage, isConnected }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
