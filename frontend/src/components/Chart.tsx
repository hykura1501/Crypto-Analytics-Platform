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

    // Create chart with modern TradingView-style design
    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
      layout: {
        background: { 
          type: 'solid',
          color: "#0a0e27" // Dark background like TradingView
        },
        textColor: "#d1d4dc",
        fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
        fontSize: 12,
      },
      grid: {
        vertLines: { 
          color: "#1e222d",
          style: 0,
          visible: true,
        },
        horzLines: { 
          color: "#1e222d",
          style: 0,
          visible: true,
        },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          width: 1,
          color: "#758696",
          style: 0, // Solid
          labelBackgroundColor: "#131722",
        },
        horzLine: {
          width: 1,
          color: "#758696",
          style: 0, // Solid
          labelBackgroundColor: "#131722",
        },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        rightOffset: 12,
        barSpacing: 8,
        borderColor: "#2a2e39",
        borderVisible: true,
      },
      rightPriceScale: {
        borderColor: "#2a2e39",
        borderVisible: true,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
    });

    chartRef.current = chart;

    // Add candlestick series with modern TradingView colors
    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#26a69a", // Green for bullish
      downColor: "#ef5350", // Red for bearish
      borderVisible: true,
      borderUpColor: "#26a69a",
      borderDownColor: "#ef5350",
      wickUpColor: "#26a69a",
      wickDownColor: "#ef5350",
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
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
    <div className="w-full h-full flex flex-col relative bg-[#0a0e27] rounded-lg overflow-hidden">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-[#0a0e27] bg-opacity-90 backdrop-blur-sm z-10">
          <div className="flex flex-col items-center gap-3">
            <div className="w-8 h-8 border-4 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
            <div className="text-[#d1d4dc] text-sm font-medium">Loading chart data...</div>
          </div>
        </div>
      )}
      {error && (
        <div className="absolute top-4 right-4 z-20 p-3 bg-[#ef5350] bg-opacity-90 backdrop-blur-sm text-white rounded-lg shadow-xl border border-[#ef5350] text-sm">
          {error}
        </div>
      )}

      {/* Modern TradingView-style Legend - OHLC Only */}
      {currentLegendData && (
        <div className="absolute top-3 left-3 z-20 bg-[#131722] bg-opacity-95 backdrop-blur-sm px-3 py-2 rounded-lg border border-[#2a2e39] shadow-xl text-xs font-mono pointer-events-none transition-all duration-200">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-1.5">
              <span className="text-[#758696]">O</span>
              <span
                className={`font-semibold ${
                  currentLegendData.open > currentLegendData.close
                    ? "text-[#ef5350]"
                    : "text-[#26a69a]"
                }`}
              >
                {currentLegendData.open.toFixed(2)}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#758696]">H</span>
              <span
                className={`font-semibold ${
                  currentLegendData.high > currentLegendData.close
                    ? "text-[#ef5350]"
                    : "text-[#26a69a]"
                }`}
              >
                {currentLegendData.high.toFixed(2)}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#758696]">L</span>
              <span
                className={`font-semibold ${
                  currentLegendData.low > currentLegendData.close
                    ? "text-[#ef5350]"
                    : "text-[#26a69a]"
                }`}
              >
                {currentLegendData.low.toFixed(2)}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#758696]">C</span>
              <span
                className={`font-bold ${
                  currentLegendData.close >= currentLegendData.open
                    ? "text-[#26a69a]"
                    : "text-[#ef5350]"
                }`}
              >
                {currentLegendData.close.toFixed(2)}
              </span>
            </div>
          </div>
        </div>
      )}

      <div ref={chartContainerRef} className="flex-1 w-full" />
    </div>
  );
}
