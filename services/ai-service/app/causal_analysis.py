"""
Causal Analysis: News-Price Correlation
"""
import logging
from datetime import datetime, timedelta
from typing import Optional, Dict
from sqlalchemy.orm import Session
from app.models import MarketPrice

logger = logging.getLogger(__name__)

def align_news_with_price(
    news_id: int,
    news_title: str,
    news_time: datetime,
    symbol: str,
    db: Session
) -> Optional[Dict]:
    """
    Analyze price movement around news time
    
    Args:
        news_id: Article ID
        news_title: Article title
        news_time: When news was published
        symbol: Crypto symbol (BTC, ETH)
        db: Database session
    
    Returns:
        {
            "news_id": int,
            "symbol": str,
            "price_before": float,
            "price_after": float, 
            "change_pct": float,
            "direction": "up" | "down" | "stable"
        }
    """
    try:
        # Get price 1 hour before news
        time_before = news_time - timedelta(hours=1)
        price_before = db.query(MarketPrice).filter(
            MarketPrice.symbol == symbol,
            MarketPrice.time <= news_time,
            MarketPrice.time >= time_before
        ).order_by(MarketPrice.time.desc()).first()
        
        # Get price 1 hour after news
        time_after = news_time + timedelta(hours=1)
        price_after = db.query(MarketPrice).filter(
            MarketPrice.symbol == symbol,
            MarketPrice.time >= news_time,
            MarketPrice.time <= time_after
        ).order_by(MarketPrice.time.asc()).first()
        
        if not price_before or not price_after:
            logger.warning(f"Not enough price data for news #{news_id}")
            return None
        
        # Calculate price movement
        price_change_pct = ((price_after.close - price_before.close) / price_before.close) * 100
        
        direction = "stable"
        if price_change_pct > 1.0:
            direction = "up"
        elif price_change_pct < -1.0:
            direction = "down"
        
        result = {
            "news_id": news_id,
            "news_title": news_title[:60],
            "symbol": symbol,
            "news_time": news_time.isoformat(),
            "price_before": price_before.close,
            "price_after": price_after.close,
            "change_pct": round(price_change_pct, 2),
            "direction": direction
        }
        
        # Log causal insight
        logger.info(
            f"📊 Causal Analysis: News #{news_id} '{news_title[:40]}...' at {news_time.strftime('%H:%M')} "
            f"→ {symbol} moved {price_change_pct:+.2f}% in next hour ({direction})"
        )
        
        return result
        
    except Exception as e:
        logger.error(f"Error in causal analysis for news #{news_id}: {e}")
        return None

def extract_symbols_from_entities(entities: dict) -> list:
    """Extract crypto symbols from news entities"""
    if not entities or not isinstance(entities, dict):
        return []
    
    tickers = entities.get("tickers", [])
    
    # Common crypto symbols
    crypto_symbols = {"BTC", "ETH", "USDT", "BNB", "SOL", "XRP", "ADA"}
    
    return [ticker for ticker in tickers if ticker in crypto_symbols]
