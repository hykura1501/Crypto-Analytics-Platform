import logging
import threading
import time
import json
import signal
import sys

from config import config
from internal.db.database import db
from internal.kafka.consumer import KafkaConsumer
from internal.gemini.gemini_client import GeminiClient
from internal.handler.news_handler import NewsHandler
from internal.handler.rss_handler import RssHandler
from internal.handler.selector_handler import SelectorHandler
from internal.api.server import Server
from internal.sentiment.sentiment import Analyzer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

def main():
    logging.info("🚀 Starting AI Service (Python)...")
    logging.info("🧠 Sentiment Analysis Engine")

    # Initialize database
    logging.info("Initializing database...")
    db.init()

    # Create Kafka consumer
    logging.info("Connecting to Kafka...")
    consumer = KafkaConsumer()

    # Initialize dependencies and handlers
    logging.info("Initializing sentiment analyzer and handlers...")
    logging.info("(This may take a while - loading ML models from HuggingFace)")
    sentiment_handler = Analyzer()
    news_handler = NewsHandler(sentiment_handler)
    
    gemini_client = GeminiClient(config.gemini.api_key, config.gemini.model)
    
    rss_handler = RssHandler(gemini_client)
    selector_handler = SelectorHandler(gemini_client)

    logging.info("=" * 60)
    logging.info("AI Service is ready. Waiting for news messages...")
    logging.info("=" * 60)

    # Create API server
    api_server = Server(rss_handler, selector_handler, sentiment_handler)
    
    # Run API server in a separate thread
    api_thread = threading.Thread(target=api_server.run)
    api_thread.daemon = True
    api_thread.start()

    # Handle signals
    def signal_handler(sig, frame):
        logging.info("Shutting down AI Service...")
        consumer.close()
        db.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # Main consumption loop
    while True:
        msg, err = consumer.read_message()

        if err:
            logging.error(f"Error reading message: {err} (will retry)")
            time.sleep(2)
            continue
        
        if msg is None:
            continue

        topic = msg.topic()
        value = msg.value().decode('utf-8')
        
        logging.info(f"Received message from topic: {topic}")

        if topic == "news_new_article":
            news_handler.handle(value)
        
        elif topic == "news_analyze_rss_structure":
            try:
                rss_structure = json.loads(value)
                rss_handler.handle(rss_structure)
            except Exception as e:
                logging.error(f"Error unmarshalling message: {e}")
        
        elif topic == "news_analyze_css_selector":
            try:
                selector_msg = json.loads(value)
                selector_handler.handle(selector_msg)
            except Exception as e:
                logging.error(f"Error unmarshalling selector message: {e}")
        
        else:
            logging.warning(f"⚠️ Received message from unknown topic: {topic}")

if __name__ == "__main__":
    main()
