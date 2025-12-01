import { useEffect, useState } from 'react'
import { getNews } from '../services/api'
import type { News } from '../services/api'
import './NewsPage.css'

export default function NewsPage() {
  const [news, setNews] = useState<News[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadNews()
  }, [])

  const loadNews = async () => {
    try {
      setLoading(true)
      const data = await getNews(50)
      setNews(data)
    } catch (error) {
      console.error('Error loading news:', error)
    } finally {
      setLoading(false)
    }
  }

  const getSentimentColor = (label: string | null) => {
    if (!label) return '#9ca3af'
    switch (label.toLowerCase()) {
      case 'positive':
        return '#34d399'
      case 'negative':
        return '#f87171'
      default:
        return '#9ca3af'
    }
  }

  return (
    <div className="news-page">
      <h1>News Feed</h1>
      {loading ? (
        <div className="loading">Loading news...</div>
      ) : (
        <div className="news-list">
          {news.map(article => (
            <div key={article.id} className="news-card">
              <div className="news-header">
                <h3>
                  <a href={article.url} target="_blank" rel="noopener noreferrer">
                    {article.title}
                  </a>
                </h3>
                {article.sentiment_label && (
                  <span 
                    className="sentiment-badge"
                    style={{ backgroundColor: getSentimentColor(article.sentiment_label) }}
                  >
                    {article.sentiment_label}
                  </span>
                )}
              </div>
              {article.content && (
                <p className="news-content">{article.content.substring(0, 200)}...</p>
              )}
              <div className="news-footer">
                <span className="news-date">
                  {new Date(article.published_at).toLocaleDateString()}
                </span>
                {article.sentiment_score && (
                  <span className="sentiment-score">
                    Score: {article.sentiment_score.toFixed(2)}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

