from typing import List
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from sqlalchemy import desc, func
from app.database import get_db
from app.models import Article

router = APIRouter(prefix="/articles", tags=["articles"])

@router.get("/", response_model=List[dict])
def get_articles(
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    source_id: str = Query(None, description="Filter by source: VNExpress, CoinDesk, etc."),
    event_type: str = Query(None, description="Filter by event type: listing, security, regulation, etc."),
    language: str = Query(None, description="Filter by language: en, vi"),
    db: Session = Depends(get_db)
):
    """
    Get list of articles with enhanced filtering
    
    - **skip**: Number of articles to skip (pagination)
    - **limit**: Maximum number to return (max 100)
    - **source_id**: Filter by source  
    - **event_type**: Filter by event classification
    - **language**: Filter by language
    """
    query = db.query(Article)
    
    if source_id:
        query = query.filter(Article.source_id == source_id)
    
    if event_type:
        query = query.filter(Article.event_type == event_type)
    
    if language:
        query = query.filter(Article.language == language)
    
    articles = query.order_by(desc(Article.created_at)).offset(skip).limit(limit).all()
    
    return [article.to_dict() for article in articles]

@router.get("/{article_id}", response_model=dict)
def get_article_by_id(
    article_id: int,
    db: Session = Depends(get_db)
):
    """Get a specific article by ID with full content"""
    article = db.query(Article).filter(Article.id == article_id).first()
    if not article:
        return {"error": "Article not found"}
    return article.to_full_dict()

@router.get("/stats/summary")
def get_stats(db: Session = Depends(get_db)):
    """Get statistics about crawled articles"""
    total = db.query(Article).count()
    
    # By source
    sources = db.query(
        Article.source_id,
        func.count(Article.id).label("count")
    ).group_by(Article.source_id).all()
    
    # By event type
    events = db.query(
        Article.event_type,
        func.count(Article.id).label("count")
    ).filter(Article.event_type != None).group_by(Article.event_type).all()
    
    # By language
    languages = db.query(
        Article.language,
        func.count(Article.id).label("count")
    ).group_by(Article.language).all()
    
    return {
        "total_articles": total,
        "by_source": [{"source": s[0], "count": s[1]} for s in sources],
        "by_event_type": [{"event_type": e[0], "count": e[1]} for e in events],
        "by_language": [{"language": l[0], "count": l[1]} for l in languages]
    }

@router.get("/entities/tickers")
def get_ticker_mentions(
    ticker: str = Query(..., description="Ticker symbol, e.g. BTC, ETH"),
    limit: int = Query(10, ge=1, le=50),
    db: Session = Depends(get_db)
):
    """Get articles mentioning a specific ticker"""
    # Query JSONB column for ticker
    articles = db.query(Article).filter(
        Article.entities['tickers'].astext.contains(f'"{ticker}"')
    ).order_by(desc(Article.created_at)).limit(limit).all()
    
    return [article.to_dict() for article in articles]
