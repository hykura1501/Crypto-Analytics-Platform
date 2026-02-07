import { useState } from 'react';
import Chart from './Chart';

interface ChartContainerProps {
  symbol: string;
  defaultInterval: string;
}

export default function ChartContainer({ symbol, defaultInterval }: ChartContainerProps) {
  const [interval, setInterval] = useState(defaultInterval);
  const intervals = ['1m', '5m', '15m', '1h', '4h', '1d'];
  const [indicators, setIndicators] = useState({
    ma25: false,
    ma50: false,
    ema12: false,
    ema26: false,
  });

  return (
    <div className="flex flex-col h-full border border-[#2a2e39] rounded-lg overflow-hidden bg-[#131722] shadow-2xl transition-all duration-300 hover:shadow-[#26a69a]/20 hover:border-[#26a69a]/30">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#2a2e39] bg-gradient-to-r from-[#131722] to-[#1a1e2e]">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-sm text-[#d1d4dc]">{symbol}</span>
          <span className="text-[#758696] text-xs">•</span>
          <span className="text-[#758696] text-xs">{interval}</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="hidden md:flex items-center gap-2 text-[11px] text-[#d1d4dc]">
            <label className="flex items-center gap-1.5 select-none cursor-pointer">
              <input
                type="checkbox"
                checked={indicators.ma25}
                onChange={(e) =>
                  setIndicators((prev) => ({ ...prev, ma25: e.target.checked }))
                }
                className="accent-[#f59e0b]"
              />
              <span className="text-[#f59e0b]">MA(25)</span>
            </label>
            <label className="flex items-center gap-1.5 select-none cursor-pointer">
              <input
                type="checkbox"
                checked={indicators.ma50}
                onChange={(e) =>
                  setIndicators((prev) => ({ ...prev, ma50: e.target.checked }))
                }
                className="accent-[#ec4899]"
              />
              <span className="text-[#ec4899]">MA(50)</span>
            </label>
            <label className="flex items-center gap-1.5 select-none cursor-pointer">
              <input
                type="checkbox"
                checked={indicators.ema12}
                onChange={(e) =>
                  setIndicators((prev) => ({
                    ...prev,
                    ema12: e.target.checked,
                  }))
                }
                className="accent-[#22c55e]"
              />
              <span className="text-[#22c55e]">EMA(12)</span>
            </label>
            <label className="flex items-center gap-1.5 select-none cursor-pointer">
              <input
                type="checkbox"
                checked={indicators.ema26}
                onChange={(e) =>
                  setIndicators((prev) => ({
                    ...prev,
                    ema26: e.target.checked,
                  }))
                }
                className="accent-[#06b6d4]"
              />
              <span className="text-[#06b6d4]">EMA(26)</span>
            </label>
          </div>

          <select
            value={interval}
            onChange={(e) => setInterval(e.target.value)}
            className="px-3 py-1.5 text-xs bg-[#1e222d] border border-[#2a2e39] rounded-md text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] focus:border-[#26a69a] transition-all cursor-pointer hover:bg-[#252936]"
          >
            {intervals.map((int) => (
              <option key={int} value={int} className="bg-[#1e222d] text-[#d1d4dc]">
                {int}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="flex-1 relative min-h-[300px]">
        <Chart
          symbol={symbol}
          interval={interval}
          indicators={indicators}
        />
      </div>
    </div>
  );
}
