import { useEffect, useState } from 'react';
import { type Article } from '../types';
import { apiClient } from '../api/client';

interface NewsSidebarProps {
  onNewsClick: (publishedAt: string) => void;
}

export default function NewsSidebar({ onNewsClick }: NewsSidebarProps) {
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
    <div className="w-80 bg-white border-r border-gray-200 flex flex-col h-full overflow-hidden">
      <div className="p-4 border-b border-gray-200">
        <h2 className="text-xl font-bold text-gray-900">Crypto News</h2>
      </div>
      
      <div className="flex-1 overflow-y-auto">
        {loading && (
          <div className="p-4 text-center text-gray-500">Loading news...</div>
        )}
        
        {error && (
          <div className="p-4 bg-red-50 text-red-800 rounded m-4">{error}</div>
        )}
        
        {!loading && !error && articles.length === 0 && (
          <div className="p-4 text-center text-gray-500">No news available</div>
        )}
        
        {articles.map((article) => (
          <a
            key={article.id}
            href={article.url}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => article.published_at && onNewsClick(article.published_at)}
            className="block p-4 border-b border-gray-100 hover:bg-gray-50 cursor-pointer transition-colors no-underline"
          >
            <div className="flex items-start justify-between mb-2">
              <span className="text-xs font-medium text-gray-500 uppercase">
                {article.source_id}
              </span>
              {article.sentiment_score !== undefined && article.sentiment_score !== null && (
                <span className={`text-xs font-semibold ${getSentimentColor(article.sentiment_score)}`}>
                  {article.sentiment_score > 0 ? 'Positive' : article.sentiment_score < 0 ? 'Negative' : 'Neutral'}
                </span>
              )}
            </div>
            
            <h3 className="text-sm font-semibold text-gray-900 mb-2 line-clamp-2">
              {article.title}
            </h3>
            
            <div className="flex items-center justify-between text-xs text-gray-500">
              <span>{formatTime(article.published_at)}</span>
              {article.sentiment_score !== undefined && article.sentiment_score !== null && (
                <span className={`font-medium ${getSentimentColor(article.sentiment_score)}`}>
                  {article.sentiment_score > 0 ? '↑' : article.sentiment_score < 0 ? '↓' : '→'}
                </span>
              )}
            </div>
          </a>
        ))}
      </div>
    </div>
  );
}

