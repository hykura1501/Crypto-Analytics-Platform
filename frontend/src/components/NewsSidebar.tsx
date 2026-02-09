import { useEffect, useState } from 'react';
import { type Article } from '../types';
import { apiClient } from '../api/client';
import { getSentimentColor, getSentimentLabel, getSentimentArrow, formatTime } from '../utils/formatters';
import { extractErrorMessage } from '../utils/errors';

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
        setArticles(data.articles);
        setError(null);
      } catch (err) {
        setError(extractErrorMessage(err, 'Failed to load news'));
      } finally {
        setLoading(false);
      }
    };

    fetchNews();
    // Refresh news every 5 minutes
    const interval = setInterval(fetchNews, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

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
                  {getSentimentLabel(article.sentiment_score)}
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
                  {getSentimentArrow(article.sentiment_score)}
                </span>
              )}
            </div>
          </a>
        ))}
      </div>
    </div>
  );
}

