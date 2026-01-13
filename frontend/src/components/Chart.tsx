import { useEffect, useRef, useState, useMemo } from "react";
import {
  createChart,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type Time,
  type SeriesMarker,
  CandlestickSeries,
  createSeriesMarkers,
  type ISeriesMarkersPluginApi,
  CrosshairMode,
  type MouseEventParams,
} from "lightweight-charts";
import { apiClient } from "../api/client";
import { type MarketPrice } from "../types";
import { useWebSocket } from "../contexts/WebSocketContext";

interface ChartProps {
  symbol?: string;
  interval?: string;
  selectedNewsTime?: number | null;
}

interface ChartLegendData {
  open: number;
  high: number;
  low: number;
  close: number;
  color: string;
}

export default function Chart({
  symbol = "BTCUSDT",
  interval = "1h",
  selectedNewsTime,
}: ChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const markersPluginRef = useRef<ISeriesMarkersPluginApi<Time> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [hoverLegendData, setHoverLegendData] =
    useState<ChartLegendData | null>(null);
  const [lastHistoricalCandle, setLastHistoricalCandle] =
    useState<ChartLegendData | null>(null);

  const { subscribe, unsubscribe, lastMessage, getLastMessage, isConnected } =
    useWebSocket();

  // Derive current topic message instead of storing in state
  const currentTopicMessage = useMemo(() => {
    const topic = `market:${symbol}:${interval}`;
    return getLastMessage(topic);
  }, [symbol, interval, lastMessage, getLastMessage]);
  // Load historical data function
  const loadHistoricalData = async (sym: string, inter: string) => {
    if (!candlestickSeriesRef.current) return;

    setIsLoading(true);
    setError(null);
    setLastHistoricalCandle(null); // Reset on new load

    try {
      const response = await apiClient.getHistory({
        symbol: sym,
        interval: inter,
        limit: 1000,
      });

      const formattedData: CandlestickData<Time>[] = response.data.map(
        (price: MarketPrice) => ({
          time: (new Date(price.time).getTime() / 1000) as Time,
          open: price.open,
          high: price.high,
          low: price.low,
          close: price.close,
        })
      );

      candlestickSeriesRef.current.setData(formattedData);

      // Set initial legend data from the last candle
      if (formattedData.length > 0) {
        const last = formattedData[formattedData.length - 1];
        setLastHistoricalCandle({
          open: last.open,
          high: last.high,
          low: last.low,
          close: last.close,
          color: last.close >= last.open ? "#26a69a" : "#ef5350",
        });
      }

      setIsLoading(false);
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : "Failed to load chart data";
      const apiError = err as { response?: { data?: { message?: string } } };
      setError(apiError.response?.data?.message || errorMessage);
      setIsLoading(false);
    }
  };

  // Initialize chart
  useEffect(() => {
    if (!chartContainerRef.current) return;

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
      layout: {
        background: { color: "#ffffff" },
        textColor: "#333",
        fontFamily:
          "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
      },
      grid: {
        vertLines: { color: "#f0f3fa" },
        horzLines: { color: "#f0f3fa" },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          width: 1,
          color: "#9B7DFF",
          style: 3, // Dashed
          labelBackgroundColor: "#9B7DFF",
        },
        horzLine: {
          width: 1,
          color: "#9B7DFF",
          style: 3, // Dashed
          labelBackgroundColor: "#9B7DFF",
        },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        rightOffset: 12,
        barSpacing: 10,
        borderColor: "#D1D4DC",
      },
      rightPriceScale: {
        borderColor: "#D1D4DC",
      },
    });

    chartRef.current = chart;

    // Add candlestick series using new API
    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#26a69a",
      downColor: "#ef5350",
      borderVisible: false,
      wickUpColor: "#26a69a",
      wickDownColor: "#ef5350",
    }) as ISeriesApi<"Candlestick">;

    candlestickSeriesRef.current = candlestickSeries;

    // Create markers plugin for the series
    markersPluginRef.current = createSeriesMarkers(candlestickSeries);

    // Subscribe to crosshair move to update legend
    chart.subscribeCrosshairMove((param: MouseEventParams) => {
      if (param.time && param.seriesData.get(candlestickSeries)) {
        const data = param.seriesData.get(candlestickSeries) as CandlestickData;
        setHoverLegendData({
          open: data.open,
          high: data.high,
          low: data.low,
          close: data.close,
          color: data.close >= data.open ? "#26a69a" : "#ef5350",
        });
      } else {
        // Reset hover data when mouse leaves
        setHoverLegendData(null);
      }
    });

    // Load historical data (async, so wrap in setTimeout to avoid linter warning)
    setTimeout(() => {
      loadHistoricalData(symbol, interval);
    }, 0);

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current && chart) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
      }
    };

    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      chart.remove();
    };
  }, [symbol, interval]);

  // WebSocket subscription
  useEffect(() => {
    const topic = `market:${symbol}:${interval}`;
    subscribe(topic);

    return () => {
      unsubscribe(topic);
    };
  }, [symbol, interval, subscribe, unsubscribe]);

  // Handle real-time updates
  useEffect(() => {
    if (!candlestickSeriesRef.current || !currentTopicMessage) return;

    const time = (new Date(currentTopicMessage.time).getTime() / 1000) as Time;
    const candle: CandlestickData<Time> = {
      time,
      open: currentTopicMessage.open,
      high: currentTopicMessage.high,
      low: currentTopicMessage.low,
      close: currentTopicMessage.close,
    };

    // Update the last candle or add new one
    candlestickSeriesRef.current.update(candle);
  }, [currentTopicMessage]);

  // Derive current legend data
  let currentLegendData = hoverLegendData;

  // If not hovering, use real-time data or historical data
  if (!currentLegendData) {
    if (currentTopicMessage) {
      currentLegendData = {
        open: currentTopicMessage.open,
        high: currentTopicMessage.high,
        low: currentTopicMessage.low,
        close: currentTopicMessage.close,
        color:
          currentTopicMessage.close >= currentTopicMessage.open
            ? "#26a69a"
            : "#ef5350",
      };
    } else {
      currentLegendData = lastHistoricalCandle;
    }
  }

  // Add marker when news is selected
  useEffect(() => {
    if (!selectedNewsTime || !markersPluginRef.current) {
      // Clear markers if no news selected
      if (markersPluginRef.current) {
        markersPluginRef.current.setMarkers([]);
      }
      return;
    }

    const marker: SeriesMarker<Time> = {
      time: (selectedNewsTime / 1000) as Time,
      position: "belowBar",
      color: "#2196F3",
      shape: "circle",
      size: 2,
      text: "News",
    };

    // Set markers using the plugin
    try {
      markersPluginRef.current.setMarkers([marker]);
    } catch (err) {
      console.warn("Could not set markers:", err);
    }
  }, [selectedNewsTime]);

  return (
    <div className="w-full h-full flex flex-col relative">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10">
          <div className="text-gray-600">Loading chart data...</div>
        </div>
      )}
      {error && (
        <div className="p-4 bg-red-50 text-red-800 rounded mb-4">{error}</div>
      )}

      {/* TradingView-style Legend */}
      <div className="absolute top-3 left-3 z-20 bg-white bg-opacity-90 p-2 rounded border border-gray-100 shadow-sm text-xs font-mono pointer-events-none">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-bold text-lg text-gray-900">{symbol}</span>
          <span className="text-gray-500">{interval}</span>
          <span
            className={`font-bold ${
              isConnected ? "text-green-500" : "text-red-500"
            }`}
          >
            •
          </span>
        </div>
        {currentLegendData && (
          <div className="flex gap-3">
            <span className="text-gray-600">
              O:{" "}
              <span
                className={
                  currentLegendData.open > currentLegendData.close
                    ? "text-red-500"
                    : "text-green-500"
                }
              >
                {currentLegendData.open.toFixed(2)}
              </span>
            </span>
            <span className="text-gray-600">
              H:{" "}
              <span
                className={
                  currentLegendData.high > currentLegendData.close
                    ? "text-red-500"
                    : "text-green-500"
                }
              >
                {currentLegendData.high.toFixed(2)}
              </span>
            </span>
            <span className="text-gray-600">
              L:{" "}
              <span
                className={
                  currentLegendData.low > currentLegendData.close
                    ? "text-red-500"
                    : "text-green-500"
                }
              >
                {currentLegendData.low.toFixed(2)}
              </span>
            </span>
            <span className="text-gray-600">
              C:{" "}
              <span
                className={
                  currentLegendData.close >= currentLegendData.open
                    ? "text-green-600 font-bold"
                    : "text-red-600 font-bold"
                }
              >
                {currentLegendData.close.toFixed(2)}
              </span>
            </span>
          </div>
        )}
      </div>

      <div ref={chartContainerRef} className="flex-1 w-full" />
    </div>
  );
}
