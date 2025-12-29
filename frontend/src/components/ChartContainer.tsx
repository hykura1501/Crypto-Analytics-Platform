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
    <div className="flex flex-col h-full border border-gray-200 rounded-lg overflow-hidden bg-white shadow-sm">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-100 bg-gray-50">
        <span className="font-medium text-sm text-gray-700">{symbol} - {interval}</span>
        <select
          value={interval}
          onChange={(e) => setInterval(e.target.value)}
          className="px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500"
        >
          {intervals.map((int) => (
            <option key={int} value={int}>
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
