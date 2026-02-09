import { useState, useEffect } from 'react';
import { apiClient } from '../api/client';
import type { PredictionResult } from '../types';

interface PredictionPanelProps {
  symbol: string;
  horizonHours: number;
  onRefresh?: () => void;
}

export default function PredictionPanel({ symbol, horizonHours, onRefresh }: PredictionPanelProps) {
  const [prediction, setPrediction] = useState<PredictionResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isTraining, setIsTraining] = useState(false);

  const fetchPrediction = async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiClient.getPrediction(symbol, horizonHours);
      setPrediction(result);
      if (onRefresh) onRefresh();
    } catch (err: any) {
      setError(err.response?.data?.detail || err.message || 'Failed to fetch prediction');
    } finally {
      setLoading(false);
    }
  };

  const handleTrain = async () => {
    setIsTraining(true);
    setError(null);
    try {
      await apiClient.trainModel({ symbol, horizon_hours: horizonHours, years: 2 });
      await fetchPrediction();
    } catch (err: any) {
      setError(err.response?.data?.detail || err.message || 'Failed to train model');
    } finally {
      setIsTraining(false);
    }
  };

  useEffect(() => {
    fetchPrediction();
    // Auto-refresh every 5 minutes
    const interval = setInterval(fetchPrediction, 5 * 60 * 1000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [symbol, horizonHours]);

  const currentPrice = prediction ? parseFloat(prediction.current_price) : 0;
  const predictedPrice = prediction ? parseFloat(prediction.predicted_price) : 0;
  const changePct = prediction ? parseFloat(prediction.predicted_change_pct) : 0;
  const confidence = prediction ? parseFloat(prediction.confidence_score) : 0;
  const isPositive = changePct >= 0;

  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
      {/* Header */}
      <div className="bg-gradient-to-r from-indigo-600 to-purple-600 px-4 py-3">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-white font-semibold text-sm">AI Price Prediction</h3>
            <p className="text-indigo-100 text-xs mt-0.5">
              {symbol} • {horizonHours}h horizon
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleTrain}
              disabled={isTraining}
              className="px-3 py-1.5 text-xs font-medium text-white bg-white/20 hover:bg-white/30 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              title="Train model"
            >
              {isTraining ? 'Training...' : 'Train'}
            </button>
            <button
              onClick={fetchPrediction}
              disabled={loading}
              className="px-3 py-1.5 text-xs font-medium text-white bg-white/20 hover:bg-white/30 rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? '...' : '↻'}
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4">
        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800">
            {error}
          </div>
        )}

        {loading && !prediction ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
          </div>
        ) : prediction ? (
          <>
            {/* Main Prediction Card */}
            <div className="mb-4 p-4 bg-gradient-to-br from-gray-50 to-gray-100 rounded-lg border border-gray-200 shadow-sm">
              <div className="flex items-start justify-between mb-3">
                <div>
                  <p className="text-xs text-gray-500 mb-1">Current Price</p>
                  <p className="text-lg font-bold text-gray-900">
                    ${currentPrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-xs text-gray-500 mb-1">Predicted Price</p>
                  <p className={`text-lg font-bold ${isPositive ? 'text-green-600' : 'text-red-600'}`}>
                    ${predictedPrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </p>
                </div>
              </div>
              
              {/* Price Change Indicator */}
              <div className="mb-3 p-2 bg-white rounded border border-gray-200">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-600">Price Change</span>
                  <div className={`flex items-center gap-1 ${isPositive ? 'text-green-600' : 'text-red-600'}`}>
                    {isPositive ? (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                      </svg>
                    ) : (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                      </svg>
                    )}
                    <span className="text-sm font-bold">
                      {isPositive ? '+' : ''}{changePct.toFixed(2)}%
                    </span>
                  </div>
                </div>
                <div className="mt-1 text-xs text-gray-500">
                  ${Math.abs(predictedPrice - currentPrice).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                </div>
              </div>
              
              <div className="flex items-center justify-between pt-3 border-t border-gray-300">
                <div>
                  <p className="text-xs text-gray-500 mb-1">Confidence</p>
                  <div className="flex items-center gap-2">
                    <div className="w-20 h-2.5 bg-gray-200 rounded-full overflow-hidden shadow-inner">
                      <div
                        className={`h-full transition-all duration-300 ${
                          confidence > 0.7 ? 'bg-gradient-to-r from-green-400 to-green-600' 
                          : confidence > 0.4 ? 'bg-gradient-to-r from-yellow-400 to-yellow-600' 
                          : 'bg-gradient-to-r from-red-400 to-red-600'
                        }`}
                        style={{ width: `${confidence * 100}%` }}
                      />
                    </div>
                    <span className="text-sm font-semibold text-gray-700">
                      {(confidence * 100).toFixed(0)}%
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Feature Analysis */}
            <div className="mb-4">
              <h4 className="text-xs font-semibold text-gray-700 mb-2">Key Factors</h4>
              <div className="space-y-2">
                {prediction.explanation.primary_factors.most_important_feature && (
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-gray-600">Primary Factor:</span>
                    <span className="font-medium text-gray-900">
                      {prediction.explanation.primary_factors.most_important_feature}
                    </span>
                  </div>
                )}
                <div className="flex items-center justify-between text-xs">
                  <span className="text-gray-600">News Influence:</span>
                  <span className={`font-medium ${
                    prediction.explanation.primary_factors.is_news_important ? 'text-green-600' : 'text-gray-500'
                  }`}>
                    {prediction.explanation.primary_factors.news_influence_pct}
                  </span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-gray-600">Model Type:</span>
                  <span className="font-medium text-gray-900">
                    {prediction.explanation.primary_factors.news_vs_technical}
                  </span>
                </div>
              </div>
            </div>

            {/* Top News Articles */}
            {prediction.top_news_articles && prediction.top_news_articles.length > 0 && (
              <div className="mb-4">
                <h4 className="text-xs font-semibold text-gray-700 mb-2">
                  Recent News ({prediction.metadata.recent_news_analyzed} analyzed)
                </h4>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {prediction.top_news_articles.slice(0, 3).map((article) => {
                    const sentimentColor = article.sentiment_score > 0.5 
                      ? 'text-green-600' 
                      : article.sentiment_score < -0.5 
                      ? 'text-red-600' 
                      : 'text-gray-600';
                    
                    return (
                      <div
                        key={article.id}
                        className="p-2 bg-gray-50 rounded border border-gray-200 hover:bg-gray-100 transition-colors"
                      >
                        <a
                          href={article.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="block"
                        >
                          <p className="text-xs font-medium text-gray-900 line-clamp-2 mb-1">
                            {article.title}
                          </p>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <span className={`font-medium ${sentimentColor}`}>
                                {article.sentiment_score > 0 ? '+' : ''}
                                {(article.sentiment_score * 100).toFixed(0)}%
                              </span>
                              {article.keywords && article.keywords.length > 0 && (
                                <span className="text-gray-500">
                                  {article.keywords.slice(0, 2).join(', ')}
                                </span>
                              )}
                            </div>
                            <span className="text-gray-400 text-xs">
                              {new Date(article.published_at).toLocaleDateString()}
                            </span>
                          </div>
                        </a>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Technical Indicators */}
            {prediction.feature_analysis.top_technical_indicators.length > 0 && (
              <div>
                <h4 className="text-xs font-semibold text-gray-700 mb-2">Top Technical Indicators</h4>
                <div className="flex flex-wrap gap-2">
                  {prediction.feature_analysis.top_technical_indicators.slice(0, 5).map((indicator) => (
                    <span
                      key={indicator.feature}
                      className="px-2 py-1 text-xs bg-blue-50 text-blue-700 rounded border border-blue-200"
                    >
                      {indicator.feature}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="text-center py-8 text-gray-500 text-sm">
            No prediction available
          </div>
        )}
      </div>
    </div>
  );
}
