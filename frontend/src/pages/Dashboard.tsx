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
    <div className="h-screen flex flex-col bg-gradient-to-br from-[#0a0e27] via-[#131722] to-[#0a0e27]">
      {/* Modern Header */}
      <header className="bg-[#131722] border-b border-[#2a2e39] px-6 py-4 shadow-lg backdrop-blur-sm bg-opacity-95">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-[#26a69a] to-[#1e7a6e] rounded-lg flex items-center justify-center shadow-lg">
              <span className="text-white font-bold text-lg">₿</span>
            </div>
            <h1 className="text-2xl font-bold bg-gradient-to-r from-[#d1d4dc] to-[#26a69a] bg-clip-text text-transparent">
              Crypto Analysis System
            </h1>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate('/news')}
              className="px-4 py-2 text-sm font-medium text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] rounded-lg hover:bg-[#252936] hover:border-[#26a69a]/50 focus:outline-none focus:ring-2 focus:ring-[#26a69a] transition-all duration-200"
            >
              News
            </button>
            {canManageSources && (
              <button
                onClick={() => navigate('/sources')}
                className="px-4 py-2 text-sm font-medium text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] rounded-lg hover:bg-[#252936] hover:border-[#26a69a]/50 focus:outline-none focus:ring-2 focus:ring-[#26a69a] transition-all duration-200"
              >
                Sources
              </button>
            )}
            {canManageUsers && (
              <button
                onClick={() => navigate('/users')}
                className="px-4 py-2 text-sm font-medium text-[#d1d4dc] bg-[#1e222d] border border-[#2a2e39] rounded-lg hover:bg-[#252936] hover:border-[#26a69a]/50 focus:outline-none focus:ring-2 focus:ring-[#26a69a] transition-all duration-200"
              >
                Users
              </button>
            )}
            <select
              value={selectedSymbol}
              onChange={(e) => setSelectedSymbol(e.target.value)}
              className="px-3 py-2 bg-[#1e222d] border border-[#2a2e39] rounded-lg text-sm text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
            >
              {symbols.map((sym) => (
                <option key={sym} value={sym} className="bg-[#1e222d] text-[#d1d4dc]">
                  {sym}
                </option>
              ))}
            </select>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-[#ef5350] to-[#c62828] rounded-lg hover:from-[#f56565] hover:to-[#d32f2f] focus:outline-none focus:ring-2 focus:ring-[#ef5350] transition-all duration-200 shadow-lg"
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
            <div className="h-full flex flex-col bg-[#131722] border border-[#2a2e39] rounded-lg shadow-2xl overflow-hidden">
              {canViewPrediction ? (
                <>
                  <div className="px-4 py-3 border-b border-[#2a2e39] bg-gradient-to-r from-[#131722] to-[#1a1e2e]">
                    <div className="flex items-center justify-between">
                      <label className="text-xs font-semibold text-[#d1d4dc]">Prediction Horizon:</label>
                      <select
                        value={predictionHorizon}
                        onChange={(e) => setPredictionHorizon(Number(e.target.value))}
                        className="px-2 py-1 text-xs bg-[#1e222d] border border-[#2a2e39] rounded-md text-[#d1d4dc] focus:outline-none focus:ring-1 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
                      >
                        <option value={1} className="bg-[#1e222d] text-[#d1d4dc]">1 hour</option>
                        <option value={4} className="bg-[#1e222d] text-[#d1d4dc]">4 hours</option>
                        <option value={24} className="bg-[#1e222d] text-[#d1d4dc]">24 hours</option>
                      </select>
                    </div>
                  </div>
                  <div className="flex-1 overflow-y-auto">
                    <PredictionPanel symbol={selectedSymbol} horizonHours={predictionHorizon} />
                  </div>
                </>
              ) : (
                <div className="flex-1 flex items-center justify-center p-4">
                  <p className="text-sm text-[#758696] text-center">
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

