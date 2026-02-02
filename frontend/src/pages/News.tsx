import { useEffect, useState, useRef, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { type Article, type Source } from '../types';
import { apiClient } from '../api/client';
import Layout from '../components/Layout';

export default function News() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [articles, setArticles] = useState<Article[]>([]);
  const [sources, setSources] = useState<Source[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nextCursor, setNextCursor] = useState<string | number | null>(null);
  const [hasMore, setHasMore] = useState(true);
  
  // Filters from URL or defaults
  const selectedSource = searchParams.get('source') || '';
  const sortBy = (searchParams.get('sort_by') as 'published_at' | 'created_at') || 'published_at';
  const sortOrder = (searchParams.get('sort_order') as 'ASC' | 'DESC') || 'DESC';
  
  const observerTarget = useRef<HTMLDivElement>(null);

  // Update URL params helper
  const updateParams = (key: string, value: string) => {
    setSearchParams(prev => {
      const newParams = new URLSearchParams(prev);
      if (value) {
        newParams.set(key, value);
      } else {
        newParams.delete(key);
      }
      return newParams;
    });
  };

  // Fetch sources for filter
  useEffect(() => {
    const fetchSources = async () => {
      try {
        const data = await apiClient.getSources();
        setSources(data || []);
      } catch (err) {
        console.error('Error fetching sources:', err);
      }
    };
    fetchSources();
  }, []);

  // Fetch news with filters
  const fetchNews = useCallback(async (cursor?: string | number, append = false) => {
    try {
      if (append) {
        setLoadingMore(true);
      } else {
        setLoading(true);
        setArticles([]);
        setNextCursor(null);
        setHasMore(true);
      }
      setError(null);

      const params: any = {
        limit: 20,
        sort_by: sortBy,
        sort_order: sortOrder,
      };

      if (cursor) {
        params.cursor = cursor;
      }

      if (selectedSource) {
        params.source_id = selectedSource;
      }

      const response = await apiClient.getNews(params);
      
      // Response is always in new format with pagination
      if (append) {
        setArticles(prev => [...prev, ...response.articles]);
      } else {
        setArticles(response.articles);
      }
      setNextCursor(response.next_cursor);
      setHasMore(response.has_more);
    } catch (err) {
      const error = err as { response?: { data?: { message?: string } }; message?: string };
      setError(error.response?.data?.message || error.message || 'Failed to load news');
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }, [selectedSource, sortBy, sortOrder]);

  // Initial load and when filters change
  useEffect(() => {
    fetchNews();
  }, [fetchNews]);

  // Infinite scroll observer
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore && !loading) {
          if (nextCursor) {
            fetchNews(nextCursor, true);
          }
        }
      },
      { threshold: 0.1 }
    );

    const currentTarget = observerTarget.current;
    if (currentTarget) {
      observer.observe(currentTarget);
    }

    return () => {
      if (currentTarget) {
        observer.unobserve(currentTarget);
      }
    };
  }, [hasMore, loadingMore, loading, nextCursor, fetchNews]);

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
    if (!timeString) return '';
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
    <Layout>
      <div className="flex flex-col min-h-full">
        {/* Filters and Sort */}
        <div className="bg-[#131722] border-b border-[#2a2e39] px-6 py-4 flex-shrink-0 sticky top-0 z-10">
          <div className="container mx-auto flex flex-wrap items-center gap-4">
            {/* Source Filter */}
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-[#d1d4dc]">Source:</label>
              <select
                value={selectedSource}
                onChange={(e) => updateParams('source', e.target.value)}
                className="px-3 py-2 bg-[#1e222d] border border-[#2a2e39] rounded-lg text-sm text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
              >
                <option value="" className="bg-[#1e222d] text-[#d1d4dc]">All Sources</option>
                {sources.map((source) => (
                  <option key={source.source_id} value={source.source_id} className="bg-[#1e222d] text-[#d1d4dc]">
                    {source.source_id}
                  </option>
                ))}
              </select>
            </div>

            {/* Sort By */}
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-[#d1d4dc]">Sort By:</label>
              <select
                value={sortBy}
                onChange={(e) => updateParams('sort_by', e.target.value)}
                className="px-3 py-2 bg-[#1e222d] border border-[#2a2e39] rounded-lg text-sm text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
              >
                <option value="published_at" className="bg-[#1e222d] text-[#d1d4dc]">Published Date</option>
                <option value="created_at" className="bg-[#1e222d] text-[#d1d4dc]">Created Date</option>
              </select>
            </div>

            {/* Sort Order */}
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-[#d1d4dc]">Order:</label>
              <select
                value={sortOrder}
                onChange={(e) => updateParams('sort_order', e.target.value)}
                className="px-3 py-2 bg-[#1e222d] border border-[#2a2e39] rounded-lg text-sm text-[#d1d4dc] focus:outline-none focus:ring-2 focus:ring-[#26a69a] hover:bg-[#252936] transition-all cursor-pointer"
              >
                <option value="DESC" className="bg-[#1e222d] text-[#d1d4dc]">Newest First</option>
                <option value="ASC" className="bg-[#1e222d] text-[#d1d4dc]">Oldest First</option>
              </select>
            </div>

            <div className="ml-auto text-sm text-[#758696]">
              {articles.length} {articles.length === 1 ? 'article' : 'articles'}
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 container mx-auto px-4 py-8">
        {loading && articles.length === 0 && (
          <div className="flex items-center justify-center py-20">
            <div className="flex flex-col items-center gap-3">
              <div className="w-8 h-8 border-4 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
              <div className="text-[#d1d4dc] text-sm font-medium">Loading news...</div>
            </div>
          </div>
        )}
        
        {error && (
          <div className="bg-[#ef5350] bg-opacity-90 backdrop-blur-sm text-white p-4 rounded-lg mb-4 border border-[#ef5350]">
            {error}
          </div>
        )}
        
        {!loading && !error && articles.length === 0 && (
          <div className="text-center text-[#758696] py-20">No news available</div>
        )}
        
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {articles.map((article) => (
            <a
              key={article.id}
              href={article.url}
              target="_blank"
              rel="noopener noreferrer"
              className="block bg-[#131722] p-6 rounded-lg shadow-xl hover:shadow-2xl transition-all border border-[#2a2e39] hover:border-[#26a69a]/50 group"
            >
              <div className="flex items-start justify-between mb-4">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-[#1e222d] text-[#d1d4dc] border border-[#2a2e39]">
                  {article.source_id}
                </span>
                {article.sentiment_score !== undefined && article.sentiment_score !== null && (
                  <span className={`text-sm font-semibold ${getSentimentColor(article.sentiment_score)}`}>
                    {article.sentiment_score > 0 ? 'Positive' : article.sentiment_score < 0 ? 'Negative' : 'Neutral'}
                  </span>
                )}
              </div>
              
              <h3 className="text-lg font-semibold text-[#d1d4dc] mb-2 line-clamp-2 group-hover:text-[#26a69a] transition-colors">
                {article.title}
              </h3>
              
              <p className="text-[#758696] text-sm mb-4 line-clamp-3">
                {article.summary || article.content_text}
              </p>
              
              <div className="flex items-center justify-between text-xs text-[#758696] mt-auto pt-4 border-t border-[#2a2e39]">
                <span>{article.published_at ? formatTime(article.published_at) : 'Unknown date'}</span>
                {article.language && (
                  <span className="px-2 py-0.5 bg-[#1e222d] rounded text-[#758696] uppercase">
                    {article.language}
                  </span>
                )}
              </div>
            </a>
          ))}
        </div>

        {/* Infinite scroll trigger */}
        {hasMore && (
          <div ref={observerTarget} className="py-8 flex justify-center">
            {loadingMore && (
              <div className="flex items-center gap-2 text-[#758696]">
                <div className="w-5 h-5 border-2 border-[#26a69a] border-t-transparent rounded-full animate-spin"></div>
                <span className="text-sm">Loading more...</span>
              </div>
            )}
          </div>
        )}

        {!hasMore && articles.length > 0 && (
          <div className="text-center text-[#758696] py-8 text-sm">
            No more articles to load
          </div>
        )}
        </div>
      </div>
    </Layout>
  );
}
