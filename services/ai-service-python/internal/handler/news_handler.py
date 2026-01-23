import logging
import json
from internal.db.database import db
from internal.sentiment.sentiment import Analyzer

class NewsHandler:
    def __init__(self):
        self.analyzer = Analyzer()

    def handle(self, msg_value):
        try:
            msg = json.loads(msg_value)
        except Exception as e:
            logging.error(f"❌ Error decoding news message: {e}")
            return

        news_id = msg.get("news_id")
        title = msg.get("title")
        content = msg.get("content")

        if not news_id or not content:
            logging.error(f"Invalid message: news_id={news_id}")
            return

        logging.info(f"📰 Processing news #{news_id}: {title[:50]}...")

        # Analyze sentiment
        keywords, sentiment_score = self.analyzer.analyze(f"{title} {content}")
        logging.info(f"🎭 Sentiment: {keywords} (score: {sentiment_score:.3f})")

        # Update database with sentiment score
        try:
            with db.conn.cursor() as cur:
                query = "UPDATE articles SET sentiment_score = %s WHERE id = %s"
                cur.execute(query, (sentiment_score, news_id))
                db.conn.commit()
            logging.info(f"✅ Updated article #{news_id} with sentiment score")
        except Exception as e:
            logging.error(f"Error updating sentiment score for article #{news_id}: {e}")
            db.conn.rollback()
            return

        # Get article details for causal analysis (logic from Go code)
        try:
            with db.conn.cursor() as cur:
                query = "SELECT id, published_at FROM articles WHERE id = %s"
                cur.execute(query, (news_id,))
                article = cur.fetchone()
                
                if not article:
                    logging.info(f"Article #{news_id} not found in database")
                    return
                
                # Causal analysis skipped as per Go code
        except Exception as e:
            logging.error(f"Error fetching article #{news_id}: {e}")
            return
