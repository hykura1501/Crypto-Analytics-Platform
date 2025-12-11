import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Chart from './Chart';
import NewsSidebar from './NewsSidebar';
import { apiClient } from '../api/client';

export default function Dashboard() {
  const navigate = useNavigate();
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [selectedInterval, setSelectedInterval] = useState('1h');
  const [selectedNewsTime, setSelectedNewsTime] = useState<number | null>(null);

  const handleNewsClick = (publishedAt: string) => {
    const timestamp = new Date(publishedAt).getTime();
    setSelectedNewsTime(timestamp);
  };

  const handleLogout = async () => {
    await apiClient.logout();
    navigate('/login');
  };

  const intervals = ['1m', '5m', '15m', '1h', '4h', '1d'];
  const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT'];

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Crypto Analysis System</h1>
          <div className="flex items-center gap-4">
            <select
              value={selectedSymbol}
              onChange={(e) => setSelectedSymbol(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              {symbols.map((sym) => (
                <option key={sym} value={sym}>
                  {sym}
                </option>
              ))}
            </select>
            <select
              value={selectedInterval}
              onChange={(e) => setSelectedInterval(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              {intervals.map((int) => (
                <option key={int} value={int}>
                  {int}
                </option>
              ))}
            </select>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 flex overflow-hidden">
        {/* News Sidebar */}
        <NewsSidebar onNewsClick={handleNewsClick} />

        {/* Chart Area */}
        <div className="flex-1 flex flex-col bg-white">
          <div className="flex-1 p-4">
            <Chart
              symbol={selectedSymbol}
              interval={selectedInterval}
              selectedNewsTime={selectedNewsTime}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

