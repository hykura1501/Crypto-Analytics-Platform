import { useState } from 'react';
import ChartContainer from '../components/ChartContainer';
import PredictionPanel from '../components/PredictionPanel';
import Layout from '../components/Layout';
import { useAuth } from '../contexts/AuthContext';

export default function Dashboard() {
  const { canViewPrediction } = useAuth();
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [predictionHorizon, setPredictionHorizon] = useState(4);

  return (
    <Layout 
      showSymbolSelector={true}
      selectedSymbol={selectedSymbol}
      onSymbolChange={setSelectedSymbol}
    >
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
    </Layout>
  );
}

