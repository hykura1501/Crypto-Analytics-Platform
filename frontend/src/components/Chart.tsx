import { useEffect, useRef, useState } from "react";
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
} from "lightweight-charts";
import { apiClient } from "../api/client";
import { type MarketPrice } from "../types";
import { WS_BASE_URL } from "../config";

interface ChartProps {
  symbol?: string;
  interval?: string;
  selectedNewsTime?: number | null;
}

export default function Chart({
  symbol = "BTCUSDT",
  interval = "1h",
  selectedNewsTime,
}: ChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const markersPluginRef = useRef<ISeriesMarkersPluginApi<Time> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load historical data function
  const loadHistoricalData = async (sym: string, inter: string) => {
    if (!candlestickSeriesRef.current) return;

    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.getHistory({
        symbol: sym,
        interval: inter,
        limit: 100,
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
      },
      grid: {
        vertLines: { color: "#f0f0f0" },
        horzLines: { color: "#f0f0f0" },
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
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

  // WebSocket connection for real-time updates
  useEffect(() => {
    if (!candlestickSeriesRef.current) return;

    const wsUrl = `${WS_BASE_URL}/ws/prices`;
    const ws = new WebSocket(wsUrl);
    let isMounted = true;

    ws.onopen = () => {
      if (isMounted) {
        console.log("WebSocket connected");
      }
    };

    ws.onmessage = (event) => {
      if (!isMounted || !candlestickSeriesRef.current) return;

      try {
        const price: MarketPrice = JSON.parse(event.data);

        // Only update if it matches the current symbol and interval
        if (
          price.symbol === symbol &&
          price.interval === interval &&
          candlestickSeriesRef.current
        ) {
          const time = (new Date(price.time).getTime() / 1000) as Time;
          const candle: CandlestickData<Time> = {
            time,
            open: price.open,
            high: price.high,
            low: price.low,
            close: price.close,
          };

          // Update the last candle or add new one
          candlestickSeriesRef.current.update(candle);
        }
      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
      }
    };

    ws.onerror = (error) => {
      if (isMounted) {
        console.error("WebSocket error:", error);
      }
    };

    ws.onclose = () => {
      if (isMounted) {
        console.log("WebSocket disconnected");
      }
    };

    wsRef.current = ws;

    return () => {
      isMounted = false;
      if (
        ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING
      ) {
        ws.close();
      }
    };
  }, [symbol, interval]);

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
    <div className="w-full h-full flex flex-col">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10">
          <div className="text-gray-600">Loading chart data...</div>
        </div>
      )}
      {error && (
        <div className="p-4 bg-red-50 text-red-800 rounded mb-4">{error}</div>
      )}
      <div ref={chartContainerRef} className="flex-1 w-full" />
    </div>
  );
}
