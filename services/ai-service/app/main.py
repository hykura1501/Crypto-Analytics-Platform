"""
AI Service Main Application
Consumes news from Kafka, analyzes sentiment, and updates database
"""
import logging
import sys
from app.kafka_consumer import create_kafka_consumer
from app.database import get_db, init_db
from app.models import Article
from app.sentiment import sentiment_analyzer
from app.causal_analysis import align_news_with_price, extract_symbols_from_entities

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger(__name__)

def process_news_message(message: dict):
    """
    Process a single news message from Kafka
    
    Message format:
    {
        "news_id": 123,
        "title": "Bitcoin hits $100k",
        "content": "Full article text..."
    }
    """
    try:
        news_id = message.get("news_id")
        title = message.get("title", "")
        content = message.get("content", "")
        
        if not news_id or not content:
            logger.warning(f"Invalid message: {message}")
            return
        
        logger.info(f"📰 Processing news #{news_id}: {title[:50]}...")
        
        # Get database session
        db = get_db()
        
        try:
            # Find article in database
            article = db.query(Article).filter(Article.id == news_id).first()
            
            if not article:
                logger.warning(f"Article #{news_id} not found in database")
                return
            
            # Analyze sentiment
            sentiment_result = sentiment_analyzer.analyze(title + " " + content)
            sentiment_score = sentiment_result["compound"]
            sentiment_label = sentiment_result["label"]
            
            logger.info(
                f"🎭 Sentiment: {sentiment_label} (score: {sentiment_score:.3f})"
            )
            
            # Update database with sentiment score
            article.sentiment_score = sentiment_score
            db.commit()
            
            logger.info(f"✅ Updated article #{news_id} with sentiment score")
            
            # Causal analysis (if article has entities and published time)
            if article.entities and article.published_at:
                symbols = extract_symbols_from_entities(article.entities)
                
                for symbol in symbols:
                    causal_result = align_news_with_price(
                        news_id=news_id,
                        news_title=title,
                        news_time=article.published_at,
                        symbol=symbol,
                        db=db
                    )
                    
                    if causal_result:
                        logger.info(
                            f"📈 {symbol}: {causal_result['price_before']:.2f} "
                            f"→ {causal_result['price_after']:.2f} "
                            f"({causal_result['change_pct']:+.2f}%, {causal_result['direction']})"
                        )
        
        finally:
            db.close()
    
    except Exception as e:
        logger.error(f"Error processing news #{news_id}: {e}", exc_info=True)

def main():
    """Main application loop"""
    logger.info("🚀 Starting AI Service...")
    logger.info("🧠 NLTK VADER Sentiment Analysis Engine")
    
    # Initialize database
    logger.info("Initializing database...")
    init_db()
    
    # Create Kafka consumer
    logger.info("Connecting to Kafka...")
    consumer = create_kafka_consumer()
    
    if not consumer:
        logger.error("Failed to create Kafka consumer. Exiting.")
        sys.exit(1)
    
    logger.info("=" * 60)
    logger.info("AI Service is ready. Waiting for news messages...")
    logger.info("=" * 60)
    
    # Main consumption loop
    try:
        for message in consumer:
            try:
                news_data = message.value
                process_news_message(news_data)
            except Exception as e:
                logger.error(f"Error in main loop: {e}", exc_info=True)
                continue
    
    except KeyboardInterrupt:
        logger.info("Shutting down AI Service...")
    finally:
        if consumer:
            consumer.close()
        logger.info("AI Service stopped.")

if __name__ == "__main__":
    main()
