from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, Float, JSON
from sqlalchemy.dialects.postgresql import ARRAY
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class Article(Base):
    """Article model - same as crawler-service"""
    __tablename__ = "articles"
    
    id = Column(Integer, primary_key=True, index=True)
    source_id = Column(String(100), nullable=False)
    url = Column(String(2000), unique=True, nullable=False)
    
    title = Column(String(1000), nullable=False)
    content_text = Column(Text, nullable=False)  
    published_at = Column(DateTime)
    
    # AI fields
    sentiment_score = Column(Float)  # Will be updated by AI
    entities = Column(JSON)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class MarketPrice(Base):
    """Market price model - for causal analysis"""
    __tablename__ = "market_prices"
    
    symbol = Column(String(20), primary_key=True)
    time = Column(DateTime, primary_key=True) 
    
    open = Column(Float, nullable=False)
    high = Column(Float, nullable=False)
    low = Column(Float, nullable=False)
    close = Column(Float, nullable=False)
    volume = Column(Float, nullable=False)
