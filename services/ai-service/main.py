#!/usr/bin/env python3
"""
AI Analysis Service
Phân tích sentiment và dự đoán xu hướng giá dựa trên tin tức
"""

import os
import psycopg2
from datetime import datetime, timedelta
from flask import Flask, request, jsonify
from flask_cors import CORS
from dotenv import load_dotenv
import pandas as pd
from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
from textblob import TextBlob
import json

load_dotenv()

app = Flask(__name__)
CORS(app)

# Initialize sentiment analyzers
vader_analyzer = SentimentIntensityAnalyzer()

class AIAnalysisService:
    def __init__(self):
        db_url = os.getenv('DATABASE_URL', 'postgres://crypto_user:crypto_pass@localhost:5432/cryptodb')
        # Ensure database name is correct
        if 'cryptodb' not in db_url:
            # Fix database name if wrong
            if '/crypto_user' in db_url:
                db_url = db_url.replace('/crypto_user', '/cryptodb')
        if not db_url.endswith('?sslmode=disable') and '?sslmode=' not in db_url:
            db_url += '?sslmode=disable'
        self.db_url = db_url
    
    def get_db_connection(self, max_retries=5):
        """Kết nối database với retry logic"""
        import time
        for attempt in range(max_retries):
            try:
                conn = psycopg2.connect(self.db_url)
                # Test connection
                cursor = conn.cursor()
                cursor.execute("SELECT 1")
                cursor.close()
                return conn
            except (psycopg2.OperationalError, psycopg2.DatabaseError) as e:
                if attempt < max_retries - 1:
                    wait_time = 2 ** attempt  # Exponential backoff
                    print(f"Database connection failed, retrying in {wait_time} seconds... ({e})")
                    time.sleep(wait_time)
                else:
                    print(f"Failed to connect to database after {max_retries} attempts")
                    raise
    
    def analyze_sentiment(self, text):
        """
        Phân tích sentiment của text
        Returns: sentiment_label (positive/negative/neutral), sentiment_score (-1 to 1)
        """
        if not text or len(text.strip()) < 10:
            return 'neutral', 0.0
        
        # VADER sentiment analysis
        vader_scores = vader_analyzer.polarity_scores(text)
        
        # TextBlob sentiment analysis
        blob = TextBlob(text)
        blob_polarity = blob.sentiment.polarity
        
        # Combine scores
        combined_score = (vader_scores['compound'] + blob_polarity) / 2
        
        # Classify
        if combined_score >= 0.05:
            label = 'positive'
        elif combined_score <= -0.05:
            label = 'negative'
        else:
            label = 'neutral'
        
        return label, float(combined_score)
    
    def analyze_news_sentiment(self):
        """
        Phân tích sentiment cho tất cả tin tức chưa được phân tích
        """
        conn = self.get_db_connection()
        cur = conn.cursor()
        
        try:
            # Get news without sentiment analysis
            cur.execute("""
                SELECT id, title, content 
                FROM news 
                WHERE sentiment_score IS NULL 
                LIMIT 100
            """)
            
            news_items = cur.fetchall()
            
            for news_id, title, content in news_items:
                # Combine title and content for analysis
                text = f"{title} {content or ''}"
                
                label, score = self.analyze_sentiment(text)
                
                # Update database
                cur.execute("""
                    UPDATE news 
                    SET sentiment_score = %s, sentiment_label = %s
                    WHERE id = %s
                """, (score, label, news_id))
            
            conn.commit()
            print(f"Analyzed sentiment for {len(news_items)} news items")
        
        except Exception as e:
            print(f"Error analyzing sentiment: {e}")
            conn.rollback()
        
        finally:
            cur.close()
            conn.close()
    
    def align_news_with_price(self, pair_id, time_window_hours=24):
        """
        Align tin tức với giá lịch sử để phân tích tương quan
        """
        conn = self.get_db_connection()
        cur = conn.cursor()
        
        try:
            # Get news with sentiment but without price alignment
            cur.execute("""
                SELECT n.id, n.published_at, n.sentiment_score, n.sentiment_label
                FROM news n
                LEFT JOIN news_price_alignment npa ON n.id = npa.news_id AND npa.pair_id = %s
                WHERE n.sentiment_score IS NOT NULL 
                AND npa.id IS NULL
                ORDER BY n.published_at DESC
                LIMIT 50
            """, (pair_id,))
            
            news_items = cur.fetchall()
            
            for news_id, published_at, sentiment_score, sentiment_label in news_items:
                # Get price before news
                cur.execute("""
                    SELECT price FROM price_history
                    WHERE pair_id = %s 
                    AND timestamp <= %s
                    ORDER BY timestamp DESC
                    LIMIT 1
                """, (pair_id, published_at))
                
                price_before_row = cur.fetchone()
                
                # Get price after news (within time window)
                time_after = published_at + timedelta(hours=time_window_hours)
                cur.execute("""
                    SELECT price FROM price_history
                    WHERE pair_id = %s 
                    AND timestamp >= %s
                    AND timestamp <= %s
                    ORDER BY timestamp DESC
                    LIMIT 1
                """, (pair_id, published_at, time_after))
                
                price_after_row = cur.fetchone()
                
                if price_before_row and price_after_row:
                    price_before = float(price_before_row[0])
                    price_after = float(price_after_row[0])
                    
                    price_change = ((price_after - price_before) / price_before) * 100
                    
                    # Calculate correlation (simplified)
                    correlation = 0.0
                    if sentiment_label == 'positive' and price_change > 0:
                        correlation = min(abs(sentiment_score) * abs(price_change) / 10, 1.0)
                    elif sentiment_label == 'negative' and price_change < 0:
                        correlation = min(abs(sentiment_score) * abs(price_change) / 10, 1.0)
                    
                    # Insert alignment
                    cur.execute("""
                        INSERT INTO news_price_alignment 
                        (news_id, pair_id, price_before, price_after, price_change_percent, 
                         correlation_score, time_window_hours)
                        VALUES (%s, %s, %s, %s, %s, %s, %s)
                    """, (news_id, pair_id, price_before, price_after, price_change, 
                          correlation, time_window_hours))
            
            conn.commit()
            print(f"Aligned {len(news_items)} news items with prices")
        
        except Exception as e:
            print(f"Error aligning news with price: {e}")
            conn.rollback()
        
        finally:
            cur.close()
            conn.close()
    
    def predict_trend(self, pair_id, time_horizon='24h'):
        """
        Dự đoán xu hướng giá dựa trên phân tích tin tức và giá lịch sử
        """
        conn = self.get_db_connection()
        cur = conn.cursor()
        
        try:
            # Get recent news sentiment
            hours = int(time_horizon.replace('h', ''))
            time_threshold = datetime.now() - timedelta(hours=hours)
            
            cur.execute("""
                SELECT sentiment_score, sentiment_label, COUNT(*) as count
                FROM news n
                JOIN news_price_alignment npa ON n.id = npa.news_id
                WHERE npa.pair_id = %s
                AND n.published_at >= %s
                GROUP BY sentiment_score, sentiment_label
            """, (pair_id, time_threshold))
            
            sentiment_data = cur.fetchall()
            
            # Get recent price trend
            cur.execute("""
                SELECT price, timestamp
                FROM price_history
                WHERE pair_id = %s
                ORDER BY timestamp DESC
                LIMIT 100
            """, (pair_id,))
            
            price_data = cur.fetchall()
            
            if not price_data:
                return {
                    'prediction': 'unknown',
                    'confidence': 0.0,
                    'reasoning': 'No price data available'
                }
            
            # Calculate price trend
            prices = [float(row[0]) for row in price_data]
            price_change = ((prices[0] - prices[-1]) / prices[-1]) * 100
            
            # Analyze sentiment
            total_sentiment = 0.0
            sentiment_count = 0
            for score, label, count in sentiment_data:
                total_sentiment += float(score) * count
                sentiment_count += count
            
            avg_sentiment = total_sentiment / sentiment_count if sentiment_count > 0 else 0.0
            
            # Make prediction
            prediction = 'neutral'
            confidence = 0.5
            
            if avg_sentiment > 0.1 and price_change > 0:
                prediction = 'up'
                confidence = min(0.5 + abs(avg_sentiment) * 0.3, 0.9)
            elif avg_sentiment < -0.1 and price_change < 0:
                prediction = 'down'
                confidence = min(0.5 + abs(avg_sentiment) * 0.3, 0.9)
            elif avg_sentiment > 0.2:
                prediction = 'up'
                confidence = min(0.4 + abs(avg_sentiment) * 0.3, 0.8)
            elif avg_sentiment < -0.2:
                prediction = 'down'
                confidence = min(0.4 + abs(avg_sentiment) * 0.3, 0.8)
            
            reasoning = f"Average sentiment: {avg_sentiment:.3f}, Price change: {price_change:.2f}%, Recent news: {sentiment_count} items"
            
            return {
                'prediction': prediction,
                'confidence': round(confidence, 4),
                'reasoning': reasoning,
                'sentiment_score': round(avg_sentiment, 4),
                'price_change_percent': round(price_change, 2)
            }
        
        except Exception as e:
            print(f"Error predicting trend: {e}")
            return {
                'prediction': 'error',
                'confidence': 0.0,
                'reasoning': str(e)
            }
        
        finally:
            cur.close()
            conn.close()


# Initialize service
ai_service = AIAnalysisService()

@app.route('/health', methods=['GET'])
def health():
    return jsonify({'status': 'healthy'})

@app.route('/analyze/sentiment', methods=['POST'])
def analyze_sentiment_endpoint():
    """Analyze sentiment for news"""
    ai_service.analyze_news_sentiment()
    return jsonify({'message': 'Sentiment analysis completed'})

@app.route('/analyze/align', methods=['POST'])
def align_news_price():
    """Align news with price data"""
    data = request.get_json()
    pair_id = data.get('pair_id')
    time_window = data.get('time_window_hours', 24)
    
    ai_service.align_news_with_price(pair_id, time_window)
    return jsonify({'message': 'News-price alignment completed'})

@app.route('/predict/<pair_id>', methods=['GET'])
def predict_trend_endpoint(pair_id):
    """Predict trend for a trading pair"""
    time_horizon = request.args.get('time_horizon', '24h')
    
    result = ai_service.predict_trend(int(pair_id), time_horizon)
    return jsonify(result)

@app.route('/analyze/all', methods=['POST'])
def analyze_all():
    """Run all analysis steps"""
    ai_service.analyze_news_sentiment()
    # Align for all pairs
    conn = ai_service.get_db_connection()
    cur = conn.cursor()
    cur.execute("SELECT id FROM trading_pairs")
    pair_ids = [row[0] for row in cur.fetchall()]
    cur.close()
    conn.close()
    
    for pair_id in pair_ids:
        ai_service.align_news_with_price(pair_id)
    
    return jsonify({'message': 'All analysis completed'})

if __name__ == '__main__':
    port = int(os.getenv('PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=True)

