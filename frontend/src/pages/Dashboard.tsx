import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ChartContainer from '../components/ChartContainer';
import { apiClient } from '../api/client';

export default function Dashboard() {
  const navigate = useNavigate();
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');

  const handleLogout = async () => {
    await apiClient.logout();
    navigate('/login');
  };

  const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT'];

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Crypto Analysis System</h1>
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/news')}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              News
            </button>
            <button
              onClick={() => navigate('/sources')}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              Sources
            </button>
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
      <div className="flex-1 p-4 overflow-hidden">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 h-full">
          <ChartContainer symbol={selectedSymbol} defaultInterval="1m" />
          <ChartContainer symbol={selectedSymbol} defaultInterval="5m" />
          <ChartContainer symbol={selectedSymbol} defaultInterval="15m" />
          <ChartContainer symbol={selectedSymbol} defaultInterval="1h" />
        </div>
      </div>
    </div>
  );
}

