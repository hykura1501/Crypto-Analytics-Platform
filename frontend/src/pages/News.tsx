import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { type Article } from '../types';
import { apiClient } from '../api/client';

export default function News() {
  const navigate = useNavigate();
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchNews = async () => {
      try {
        setLoading(true);
        const data = await apiClient.getNews({ limit: 50 });
        setArticles(data);
        setError(null);
      } catch (err) {
        const error = err as { response?: { data?: { message?: string } }; message?: string };
        setError(error.response?.data?.message || error.message || 'Failed to load news');
      } finally {
        setLoading(false);
      }
    };

    fetchNews();
    // Refresh news every 5 minutes
    const interval = setInterval(fetchNews, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  const getSentimentColor = (sentimentScore?: number): string => {
    if (sentimentScore === undefined || sentimentScore === null) {
      return 'text-gray-500';
    }
    if (sentimentScore > 0) {
      return 'text-green-600';
    } else if (sentimentScore < 0) {
      return 'text-red-600';
    }
    return 'text-gray-500';
  };

  const formatTime = (timeString: string): string => {
    const date = new Date(timeString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Crypto News</h1>
          <button
            onClick={() => navigate('/')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            Back to Dashboard
          </button>
        </div>
      </header>

      <div className="flex-1 container mx-auto px-4 py-8">
        {loading && (
          <div className="text-center text-gray-500">Loading news...</div>
        )}
        
        {error && (
          <div className="bg-red-50 text-red-800 p-4 rounded mb-4">{error}</div>
        )}
        
        {!loading && !error && articles.length === 0 && (
          <div className="text-center text-gray-500">No news available</div>
        )}
        
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {articles.map((article) => (
            <a
              key={article.id}
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="block bg-white p-6 rounded-lg shadow hover:shadow-md transition-shadow border border-gray-200"
            >
              <div className="flex items-start justify-between mb-4">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                  {article.source_id}
                </span>
                {article.sentiment_score !== undefined && article.sentiment_score !== null && (
                  <span className={`text-sm font-semibold ${getSentimentColor(article.sentiment_score)}`}>
                    {article.sentiment_score > 0 ? 'Positive' : article.sentiment_score < 0 ? 'Negative' : 'Neutral'}
                  </span>
                )}
              </div>
              
              <h3 className="text-lg font-medium text-gray-900 mb-2 line-clamp-2">
                {article.title}
              </h3>
              
              <p className="text-gray-500 text-sm mb-4 line-clamp-3">
                {article.summary || article.content_text}
              </p>
              
              <div className="flex items-center justify-between text-xs text-gray-400 mt-auto">
                <span>{article.published_at ? formatTime(article.published_at) : ''}</span>
              </div>
            </a>
          ))}
        </div>
      </div>
    </div>
  );
}
