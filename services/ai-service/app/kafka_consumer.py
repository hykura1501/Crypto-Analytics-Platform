"""
Kafka Consumer for News Messages
"""
import json
import logging
import time
from kafka import KafkaConsumer
from kafka.errors import KafkaError
from app.config import settings

logger = logging.getLogger(__name__)

def create_kafka_consumer():
    """Create and configure Kafka consumer"""
    max_retries = 5
    retry_delay = 5
    
    for attempt in range(max_retries):
        try:
            consumer = KafkaConsumer(
                settings.KAFKA_TOPIC,
                bootstrap_servers=settings.KAFKA_BROKER,
                group_id=settings.KAFKA_GROUP_ID,
                auto_offset_reset='earliest',
                enable_auto_commit=True,
                value_deserializer=lambda m: json.loads(m.decode('utf-8'))
            )
            logger.info(f"✅ Kafka consumer connected to {settings.KAFKA_BROKER}")
            logger.info(f"📡 Listening to topic: {settings.KAFKA_TOPIC}")
            return consumer
        except KafkaError as e:
            logger.error(f"Attempt {attempt + 1}/{max_retries}: Failed to connect to Kafka: {e}")
            if attempt < max_retries - 1:
                logger.info(f"Retrying in {retry_delay}s...")
                time.sleep(retry_delay)
            else:
                logger.error("Max retries reached. Could not connect to Kafka.")
                raise
    
    return None
