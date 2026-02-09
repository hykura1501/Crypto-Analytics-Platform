import { useState } from 'react';
import Chart from './Chart';

interface ChartContainerProps {
  symbol: string;
  defaultInterval: string;
}

export default function ChartContainer({ symbol, defaultInterval }: ChartContainerProps) {
  const [interval, setInterval] = useState(defaultInterval);
  const intervals = ['1m', '5m', '15m', '1h', '4h', '1d'];

  return (
    <div className="flex flex-col h-full border border-[#2a2e39] rounded-lg overflow-hidden bg-[#131722] shadow-2xl transition-all duration-300 hover:shadow-[#26a69a]/20 hover:border-[#26a69a]/30">
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#2a2e39] bg-gradient-to-r from-[#131722] to-[#1a1e2e]">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-sm text-[#d1d4dc]">{symbol}</span>
          <span className="text-[#758696] text-xs">•</span>
          <span className="text-[#758696] text-xs">{interval}</span>
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
      <div className="flex-1 relative min-h-[300px]">
        <Chart symbol={symbol} interval={interval} />
      </div>
    </div>
  );
}
