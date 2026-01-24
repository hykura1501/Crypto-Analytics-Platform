import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ChartContainer from '../components/ChartContainer';
import PredictionPanel from '../components/PredictionPanel';
import { apiClient } from '../api/client';
import { useAuth } from '../contexts/AuthContext';

export default function Dashboard() {
  const navigate = useNavigate();
  const { canViewPrediction, canManageSources, canManageUsers } = useAuth();
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [predictionHorizon, setPredictionHorizon] = useState(4);

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
            {canManageSources && (
              <button
                onClick={() => navigate('/sources')}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                Sources
              </button>
            )}
            {canManageUsers && (
              <button
                onClick={() => navigate('/users')}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                Users
              </button>
            )}
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
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 h-full">
          {/* Charts Section - 3 columns */}
          <div className="lg:col-span-3 grid grid-cols-1 md:grid-cols-2 gap-4 h-full">
            <ChartContainer symbol={selectedSymbol} defaultInterval="1m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="5m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="15m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="1h" />
          </div>
          
          {/* AI Prediction Panel - 1 column (VIP/ADMIN only) */}
          <div className="lg:col-span-1 h-full">
            <div className="h-full flex flex-col">
              {canViewPrediction ? (
                <>
                  <div className="mb-2 flex items-center justify-between">
                    <label className="text-xs font-medium text-gray-700">Prediction Horizon:</label>
                    <select
                      value={predictionHorizon}
                      onChange={(e) => setPredictionHorizon(Number(e.target.value))}
                      className="px-2 py-1 text-xs border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    >
                      <option value={1}>1 hour</option>
                      <option value={4}>4 hours</option>
                      <option value={24}>24 hours</option>
                    </select>
                  </div>
                  <div className="flex-1 overflow-y-auto">
                    <PredictionPanel symbol={selectedSymbol} horizonHours={predictionHorizon} />
                  </div>
                </>
              ) : (
                <div className="flex-1 flex items-center justify-center p-4 bg-white border border-gray-200 rounded-lg">
                  <p className="text-sm text-gray-500 text-center">
                    AI prediction is available for VIP and Admin accounts.
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

