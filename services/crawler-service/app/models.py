from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, Float, JSON
from sqlalchemy.dialects.postgresql import ARRAY
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class Article(Base):
    """Enhanced news article model with entity extraction and event tracking"""
    __tablename__ = "articles"
    
    # Primary key
    id = Column(Integer, primary_key=True, index=True)
    
    # Source & URL
    source_id = Column(String(100), nullable=False, index=True)  # VNExpress, CoinDesk, etc.
    url = Column(String(2000), unique=True, nullable=False, index=True)
    url_normalized = Column(String(2000), index=True)  # Canonical URL for dedup
    
    # Deduplication
    content_hash = Column(String(64), index=True)  # SimHash for near-duplicate detection
    
    # Core content
    title = Column(String(1000), nullable=False)
    author = Column(String(200))
    published_at = Column(DateTime, index=True)
    crawled_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    
    # Storage
    html_raw_path = Column(String(500))  # S3/MinIO path to raw HTML
    content_text = Column(Text, nullable=False)
    language = Column(String(5), default='en')  # en, vi
    tags = Column(ARRAY(String), default=list)
    summary = Column(Text)
    
    # Entities & Analysis
    entities = Column(JSON)  # {tickers: [...], coins: [...], people: [...], orgs: [...]}
    sentiment_score = Column(Float)  # -1.0 to 1.0 (will be filled by AI service)
    embedding_vector = Column(ARRAY(Float))  # For semantic search (768-dim)
    confidence = Column(Float, default=1.0)  # Extraction confidence 0-1
    
    # Event tracking
    event_time = Column(DateTime)  # If article mentions specific event time
    event_type = Column(String(50), index=True)  # announcement, hack, regulation, listing, etc.
    
    # Metadata
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    def to_dict(self):
        return {
            "id": self.id,
            "source_id": self.source_id,
            "url": self.url,
            "title": self.title,
            "author": self.author,
            "published_at": self.published_at.isoformat() if self.published_at else None,
            "crawled_at": self.crawled_at.isoformat(),
            "content_text": self.content_text[:500] + "..." if len(self.content_text) > 500 else self.content_text,
            "language": self.language,
            "tags": self.tags,
            "summary": self.summary,
            "entities": self.entities,
            "sentiment_score": self.sentiment_score,
            "confidence": self.confidence,
            "event_time": self.event_time.isoformat() if self.event_time else None,
            "event_type": self.event_type,
            "created_at": self.created_at.isoformat()
        }
    
    def to_full_dict(self):
        """Full content for detail view"""
        data = self.to_dict()
        data["content_text"] = self.content_text
        data["html_raw_path"] = self.html_raw_path
        data["url_normalized"] = self.url_normalized
        data["content_hash"] = self.content_hash
        return data
