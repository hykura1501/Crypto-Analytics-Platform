import { useState } from 'react';
import ChartContainer from '../components/ChartContainer';
import PredictionPanel from '../components/PredictionPanel';
import Layout from '../components/Layout';
import { useAuth } from '../contexts/AuthContext';

export default function Dashboard() {
  const { canViewPrediction } = useAuth();
  const [selectedSymbol, setSelectedSymbol] = useState('BTCUSDT');
  const [predictionHorizon, setPredictionHorizon] = useState(4);
  const [isPredictionExpanded, setIsPredictionExpanded] = useState(false);

  return (
    <Layout 
      showSymbolSelector={true}
      selectedSymbol={selectedSymbol}
      onSymbolChange={setSelectedSymbol}
    >
      {/* Main Content */}
      <div className="h-full p-4 overflow-hidden flex flex-col">
        <div className={`grid grid-cols-1 gap-4 h-full ${canViewPrediction ? 'lg:grid-cols-[1fr_auto]' : 'lg:grid-cols-1'}`}>
          {/* Charts Section - bị đẩy sang trái khi có panel/nút */}
          <div className={`grid grid-cols-1 md:grid-cols-2 grid-rows-2 gap-4 h-full min-w-0 ${canViewPrediction ? 'lg:col-span-1' : ''}`}>
            <ChartContainer symbol={selectedSymbol} defaultInterval="1m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="5m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="15m" />
            <ChartContainer symbol={selectedSymbol} defaultInterval="1h" />
          </div>

          {/* Cột phải - luôn có khi VIP/ADMIN: nút thu gọn full height hoặc panel mở */}
          {canViewPrediction && (
            <div className="h-full flex justify-end shrink-0">
              {isPredictionExpanded ? (
                <div className="h-full w-80 min-w-0 flex flex-col bg-[#131722] border border-[#2a2e39] rounded-lg shadow-2xl overflow-hidden">
                  <div className="px-4 py-3 border-b border-[#2a2e39] bg-gradient-to-r from-[#131722] to-[#1a1e2e] flex items-center justify-between shrink-0">
                    <div className="flex items-center gap-3">
                      <button
                        onClick={() => setIsPredictionExpanded(false)}
                        className="p-1 text-[#758696] hover:text-[#d1d4dc] hover:bg-[#2a2e39] rounded transition-colors"
                        title="Thu gọn"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9 18l6-6-6-6"/></svg>
                      </button>
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
                  <div className="flex-1 overflow-y-auto min-h-0">
                    <PredictionPanel symbol={selectedSymbol} horizonHours={predictionHorizon} />
                  </div>
                </div>
              ) : (
                <button
                  onClick={() => setIsPredictionExpanded(true)}
                  className="h-full w-12 flex flex-col items-center justify-center gap-1.5 p-2 bg-[#131722] border border-[#2a2e39] rounded-lg shadow-xl hover:bg-[#1a1e2e] transition-colors group shrink-0"
                  title="Xem AI Price Prediction"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-[#26a69a] shrink-0 group-hover:scale-110 transition-transform"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/></svg>
                  <span className="text-[9px] font-medium text-[#d1d4dc] text-center leading-tight">AI Predict</span>
                  <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-[#758696] shrink-0"><path d="M15 18l-6-6 6-6"/></svg>
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}

