import json
import logging
from kafka import KafkaProducer
from app.config import settings

logger = logging.getLogger(__name__)

class NewsKafkaProducer:
    def __init__(self):
        self.producer = None
        self._connect()
    
    def _connect(self):
        """Connect to Kafka"""
        try:
            self.producer = KafkaProducer(
                bootstrap_servers=settings.kafka_broker,
                value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                retries=3
            )
            logger.info(f"Connected to Kafka at {settings.kafka_broker}")
        except Exception as e:
            logger.error(f"Failed to connect to Kafka: {e}")
            self.producer = None
    
    def publish_news(self, news_id: int, title: str, content: str):
        """Publish news article to Kafka"""
        if not self.producer:
            logger.warning("Kafka producer not available, skipping publish")
            return False
        
        try:
            message = {
                "news_id": news_id,
                "title": title,
                "content": content[:1000]  # Limit content size for Kafka message
            }
            
            self.producer.send(
                settings.kafka_topic,
                key=str(news_id).encode('utf-8'),
                value=message
            )
            self.producer.flush()
            logger.info(f"✅ Published news #{news_id} to Kafka")
            return True
        except Exception as e:
            logger.error(f"Failed to publish to Kafka: {e}")
            return False
    
    def close(self):
        """Close Kafka producer"""
        if self.producer:
            self.producer.close()
            logger.info("Kafka producer closed")

# Singleton instance
news_producer = NewsKafkaProducer()
