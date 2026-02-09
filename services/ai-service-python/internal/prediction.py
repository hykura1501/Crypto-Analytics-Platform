"""
Crypto Price Prediction Pipeline - Simple Version
"""

import pandas as pd
import numpy as np
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional
from sklearn.metrics import mean_squared_error, mean_absolute_error
import xgboost as xgb
import shap

from internal.db.database import db
from internal.sentiment.sentiment import Analyzer

logger = logging.getLogger(__name__)

class PredictionPipeline:
    """Simple prediction pipeline"""
    
    def __init__(self, analyzer: Optional[Analyzer] = None):
        # Dictionary to store models for different symbol+horizon combinations
        # Key format: f"{symbol}_{horizon_hours}"
        self.models = {} 
        self.feature_configs = {} # Store feature names for each model
        self.explainers = {} # Store SHAP explainers for each model
        self.analyzer = analyzer

    def _get_model_key(self, symbol: str, horizon_hours: int) -> str:
        return f"{symbol}_{horizon_hours}"
    
    def load_price_data(self, symbol: str = "BTCUSDT", years: int = 2) -> pd.DataFrame:
        """Load price data from database"""
        query = """
            SELECT symbol, time, interval, open, high, low, close, volume
            FROM market_prices
            WHERE symbol = %s AND interval = '1h'
            AND time >= NOW() - INTERVAL '%s years'
            ORDER BY time ASC
        """
        df = pd.read_sql(query, db.conn, params=(symbol, years))
        return df
    
    def load_news_data(self, days: int = 730) -> pd.DataFrame:
        """Load news data from database"""
        query = """
            SELECT id, source_id, title, content_text, published_at, crawled_at,
                   sentiment_score, url, language
            FROM articles
            WHERE crawled_at >= NOW() - INTERVAL '%s days'
            ORDER BY crawled_at ASC
        """
        df = pd.read_sql(query, db.conn, params=(days,))
        
        # Parse time
        if 'published_at' in df.columns:
            df['time'] = pd.to_datetime(df['published_at'], errors='coerce')
            df['time'] = df['time'].fillna(pd.to_datetime(df['crawled_at']))
        else:
            df['time'] = pd.to_datetime(df['crawled_at'])
        
        return df
    
    def create_price_features(self, df: pd.DataFrame) -> pd.DataFrame:
        """Create price features"""
        df = df.copy()
        df = df.sort_values('time').reset_index(drop=True)
        
        # Returns
        df['return_1h'] = df['close'].pct_change(1)
        df['return_24h'] = df['close'].pct_change(24)
        
        # Volatility
        df['volatility_24h'] = df['return_1h'].rolling(24).std()
        
        # SMA
        df['sma_7'] = df['close'].rolling(7).mean()
        df['sma_24'] = df['close'].rolling(24).mean()
        
        # RSI
        delta = df['close'].diff()
        gain = delta.where(delta > 0, 0).rolling(14).mean()
        loss = -delta.where(delta < 0, 0).rolling(14).mean()
        rs = gain / loss
        df['rsi'] = 100 - (100 / (1 + rs))
        
        # MACD
        ema12 = df['close'].ewm(span=12).mean()
        ema26 = df['close'].ewm(span=26).mean()
        df['macd'] = ema12 - ema26
        
        # Lag features
        df['close_lag_1'] = df['close'].shift(1)
        df['close_lag_24'] = df['close'].shift(24)
        
        return df
    
    def process_news(self, news_df: pd.DataFrame, price_df: pd.DataFrame) -> pd.DataFrame:
        """Process news and align with price time buckets with advanced features"""
        price_df = price_df.copy()
        
        if news_df.empty:
            price_df['sentiment_mean'] = 0
            price_df['news_count'] = 0
            price_df['sentiment_positive_ratio'] = 0
            price_df['sentiment_lag_1h'] = 0
            price_df['sentiment_lag_6h'] = 0
            price_df['sentiment_ma_24h'] = 0
            price_df['sentiment_std_24h'] = 0
            return price_df
        
        # Simple sentiment analysis
        news_df['sentiment'] = news_df.get('sentiment_score', 0).fillna(0)
        
        # Normalize timezone before creating buckets
        if news_df['time'].dt.tz is not None:
            news_df['time'] = news_df['time'].dt.tz_localize(None)
        
        price_time = pd.to_datetime(price_df['time'])
        if price_time.dt.tz is not None:
            price_time = price_time.dt.tz_localize(None)
        
        # Round to hourly buckets (naive datetime, no timezone)
        news_df['time_bucket'] = news_df['time'].dt.floor('1h')
        price_df['time_bucket'] = price_time.dt.floor('1h')
        
        # Aggregate news per hour with more features
        news_agg = news_df.groupby('time_bucket').agg({
            'sentiment': ['mean', 'count', 'std', lambda x: (x > 0.1).sum()]
        }).reset_index()
        news_agg.columns = ['time_bucket', 'sentiment_mean', 'news_count', 'sentiment_std', 'positive_count']
        
        # Merge with price
        result = price_df.merge(news_agg, on='time_bucket', how='left')
        result['sentiment_mean'] = result['sentiment_mean'].fillna(0)
        result['news_count'] = result['news_count'].fillna(0)
        result['sentiment_std'] = result['sentiment_std'].fillna(0)
        result['positive_count'] = result['positive_count'].fillna(0)
        
        # Additional sentiment features
        result['sentiment_positive_ratio'] = result['positive_count'] / (result['news_count'] + 1)
        
        # Lag features for sentiment
        result['sentiment_lag_1h'] = result['sentiment_mean'].shift(1).fillna(0)
        result['sentiment_lag_6h'] = result['sentiment_mean'].shift(6).fillna(0)
        
        # Rolling features for sentiment
        result['sentiment_ma_24h'] = result['sentiment_mean'].rolling(24, min_periods=1).mean().fillna(0)
        result['sentiment_std_24h'] = result['sentiment_mean'].rolling(24, min_periods=1).std().fillna(0)
        
        return result
    
    def select_features(self, df: pd.DataFrame, target: pd.Series, n: int = 30) -> List[str]:
        """Select top N features by correlation, ensuring news features are included"""
        exclude_cols = ['time', 'symbol', 'interval', 'open', 'high', 'low', 'close', 'volume', 
                       'target', 'time_bucket', 'id', 'url', 'title', 'content_text', 
                       'published_at', 'crawled_at', 'source_id', 'language', 'sentiment_score',
                       'positive_count', 'sentiment_std']
        
        correlations = {}
        news_features = {}
        tech_features = {}
        
        for col in df.select_dtypes(include=[np.number]).columns:
            if col not in exclude_cols and col != target.name:
                try:
                    corr = abs(df[col].corr(target))
                    if not np.isnan(corr):
                        # Categorize features
                        if 'sentiment' in col.lower() or 'news' in col.lower():
                            news_features[col] = corr
                        else:
                            tech_features[col] = corr
                        correlations[col] = corr
                except:
                    pass
        
        # Ensure balanced feature selection: at least 20% news features
        min_news_features = max(3, int(n * 0.2))
        
        # Select top news features
        top_news = sorted(news_features.items(), key=lambda x: x[1], reverse=True)[:min_news_features]
        selected_news = [f[0] for f in top_news]
        
        # Select remaining from all features
        remaining_n = n - len(selected_news)
        all_sorted = sorted(correlations.items(), key=lambda x: x[1], reverse=True)
        
        selected_features = selected_news.copy()
        for feat, corr in all_sorted:
            if feat not in selected_features and len(selected_features) < n:
                selected_features.append(feat)
        
        return selected_features if selected_features else list(df.select_dtypes(include=[np.number]).columns[:n])
    
    def train(self, symbol: str = "BTCUSDT", horizon_hours: int = 4, years: int = 2):
        """Train model"""
        logger.info(f"Loading data for {symbol}...")
        
        # Load data
        price_df = self.load_price_data(symbol, years)
        if price_df.empty:
            raise ValueError(f"No price data found for {symbol}")
        
        logger.info(f"Loaded {len(price_df)} price records")
        
        news_df = self.load_news_data(days=years*365)
        logger.info(f"Loaded {len(news_df)} news records")
        
        # Create features
        price_df = self.create_price_features(price_df)
        
        # Process news
        df = self.process_news(news_df, price_df)
        
        if df.empty:
            raise ValueError("No data after feature engineering")
        
        logger.info(f"Data after merging: {len(df)} samples")
        
        # Create target (future price)
        df['target'] = df['close'].shift(-horizon_hours)
        df = df.dropna(subset=['target'])
        
        if df.empty:
            raise ValueError("No data after creating target")
        
        logger.info(f"Data after target creation: {len(df)} samples")
        
        # Select features
        feature_cols = self.select_features(df, df['target'], n=30)
        if not feature_cols:
            raise ValueError("No features selected")
        
        model_key = self._get_model_key(symbol, horizon_hours)
        self.feature_configs[model_key] = feature_cols
        
        # Log feature distribution
        news_feat_count = sum(1 for f in feature_cols if 'sentiment' in f.lower() or 'news' in f.lower())
        tech_feat_count = len(feature_cols) - news_feat_count
        logger.info(f"Selected {len(feature_cols)} features: {news_feat_count} news, {tech_feat_count} technical")
        
        # Prepare data
        X = df[feature_cols].fillna(0)
        y = df['target']
        
        if len(X) < 100:
            raise ValueError(f"Not enough data: {len(X)} samples (minimum 100 required)")
        
        # Train/test split (80/20)
        split_idx = int(len(X) * 0.8)
        X_train, X_test = X.iloc[:split_idx], X.iloc[split_idx:]
        y_train, y_test = y.iloc[:split_idx], y.iloc[split_idx:]
        
        if len(X_test) == 0:
            raise ValueError("No test data available")
        
        # Train model with better hyperparameters
        logger.info(f"Training model for {model_key} with {len(X_train)} samples, {len(feature_cols)} features...")
        model = xgb.XGBRegressor(
            n_estimators=200, 
            max_depth=6, 
            learning_rate=0.05, 
            subsample=0.8,
            colsample_bytree=0.8,
            random_state=42
        )
        model.fit(X_train, y_train)
        self.models[model_key] = model
        
        # Evaluate
        y_pred = model.predict(X_test)
        rmse = np.sqrt(mean_squared_error(y_test, y_pred))
        mae = mean_absolute_error(y_test, y_pred)
        
        # Directional accuracy
        if len(y_test) > 1:
            y_test_series = pd.Series(y_test.values, index=y_test.index)
            y_pred_series = pd.Series(y_pred, index=y_test.index)
            direction_true = np.sign(y_test_series.diff().iloc[1:])
            direction_pred = np.sign(y_pred_series.diff().iloc[1:])
            # Ensure same index for comparison
            common_idx = direction_true.index.intersection(direction_pred.index)
            if len(common_idx) > 0:
                dir_acc = (direction_true.loc[common_idx] == direction_pred.loc[common_idx]).mean() * 100
            else:
                dir_acc = 0.0
        else:
            dir_acc = 0.0
        
        logger.info(f"Test RMSE: {rmse:.2f}, MAE: {mae:.2f}, Directional Accuracy: {dir_acc:.1f}%")
        
        # Initialize SHAP explainer
        try:
            sample = X_train.sample(min(100, len(X_train)), random_state=42)
            self.explainers[model_key] = shap.TreeExplainer(model)
            logger.info("SHAP explainer initialized successfully")
        except Exception as e:
            logger.warning(f"SHAP not available: {e}")
        
        return {
            'rmse': float(rmse),
            'mae': float(mae),
            'directional_accuracy': float(dir_acc),
            'features_count': len(feature_cols),
            'news_features_count': news_feat_count,
            'tech_features_count': tech_feat_count
        }
    
    def predict(self, symbol: str = "BTCUSDT", horizon_hours: int = 4) -> Dict:
        """Make prediction"""
        model_key = self._get_model_key(symbol, horizon_hours)
        model = self.models.get(model_key)
        feature_names = self.feature_configs.get(model_key)
        explainer = self.explainers.get(model_key)
        
        if model is None or feature_names is None:
            logger.info(f"Model not found for {model_key}. Training new model...")
            self.train(symbol=symbol, horizon_hours=horizon_hours)
            # Reload model after training
            model = self.models.get(model_key)
            feature_names = self.feature_configs.get(model_key)
            explainer = self.explainers.get(model_key)
            
            if model is None:
                 raise ValueError(f"Failed to train model for {symbol} with {horizon_hours}h horizon.")
        
        # Load recent data (consistent with training range for better feature alignment)
        price_df = self.load_price_data(symbol, years=1)
        # Load news with longer horizon for better context
        lookback_days = max(90, horizon_hours // 24 + 30)
        news_df = self.load_news_data(days=lookback_days)
        
        # Create features
        price_df = self.create_price_features(price_df)
        df = self.process_news(news_df, price_df)
        
        # Get latest features
        latest = df.iloc[[-1]]
        X = latest[feature_names].fillna(0)
        
        # Predict
        current_price = float(latest['close'].iloc[0])
        predicted_price = float(model.predict(X)[0])
        
        # 1. Feature Importance (từ model)
        importance = {f: float(imp) for f, imp in zip(feature_names, model.feature_importances_)}
        top_features = sorted(importance.items(), key=lambda x: x[1], reverse=True)[:10]
        
        # 2. SHAP Explanation (XAI)
        explanation = {}
        feature_contrib = {}
        try:
            if explainer:
                shap_values = explainer.shap_values(X)
                feature_contrib = dict(zip(feature_names, shap_values[0]))
                # Sort by absolute contribution
                top_contrib = sorted(feature_contrib.items(), key=lambda x: abs(x[1]), reverse=True)[:10]
                explanation = {k: float(v) for k, v in top_contrib}
        except Exception as e:
            logger.warning(f"SHAP error: {e}")
            # Fallback: use feature importance
            top_contrib = sorted(importance.items(), key=lambda x: x[1], reverse=True)[:5]
            explanation = {k: float(v) for k, v in top_contrib}
        
        # 3. Phân loại features: News sentiment vs Technical indicators
        news_features = [f for f in feature_names if 'sentiment' in f.lower() or 'news' in f.lower()]
        tech_features = [f for f in feature_names if f not in news_features]
        
        # Top news features
        news_importance = {f: float(importance.get(f, 0)) for f in news_features if f in importance}
        top_news = sorted(news_importance.items(), key=lambda x: x[1], reverse=True)[:5]
        
        # Top technical indicators
        tech_importance = {f: float(importance.get(f, 0)) for f in tech_features if f in importance}
        top_tech = sorted(tech_importance.items(), key=lambda x: x[1], reverse=True)[:5]
        
        # 4. Top 3 News Articles with Keywords
        top_news_articles = []
        recent_news_count = 0
        
        if not news_df.empty:
            # Filter news relevant to the prediction horizon
            lookback_hours = max(horizon_hours * 2, 48)
            cutoff_time = news_df['time'].max() - timedelta(hours=lookback_hours)
            recent_news = news_df[news_df['time'] >= cutoff_time].copy()
            
            if recent_news.empty:
                recent_news = news_df.tail(10)
            
            recent_news_count = len(recent_news)
            
            # Sort by time descending (most recent first)
            recent_news = recent_news.sort_values('time', ascending=False)
            
            # Take top 3 most recent articles
            top_3_news = recent_news.head(3)
            
            for idx, row in top_3_news.iterrows():
                article_id = int(row.get('id', 0))
                title = str(row.get('title', ''))
                url = str(row.get('url', ''))
                published_at = str(row.get('published_at', ''))
                sentiment_score = float(row.get('sentiment_score', 0))
                language = str(row.get('language', 'en'))
                
                # Extract keywords for this specific article
                text = (title + " " + str(row.get('content_text', ''))).strip()
                keywords = []
                
                if len(text) >= 10 and self.analyzer:
                    try:
                        # Analyzer.analyze trả về (keywords, sentiment_score)
                        kws, _ = self.analyzer.analyze(text, lang=language, top_k=5)
                        # Filter out noise
                        keywords = [k for k in kws if k not in ["N/A", "Error", "n/a"] and len(k) > 2]
                    except Exception as e:
                        logger.warning(f"Error extracting keywords for article {article_id}: {e}")
                        keywords = []
                else:
                    # Fallback: simple extraction from title
                    import re
                    words = re.findall(r'\b[a-z]{4,}\b', title.lower())
                    stopwords = {'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 
                                'for', 'of', 'with', 'by', 'is', 'are', 'was', 'were', 'this', 'that'}
                    keywords = [w for w in words if w not in stopwords][:5]
                
                top_news_articles.append({
                    "id": article_id,
                    "title": title,
                    "url": url,
                    "published_at": published_at,
                    "sentiment_score": sentiment_score,
                    "language": language,
                    "keywords": keywords[:5]  # Ensure max 5 keywords
                })
        
        # Tính confidence dựa trên nhiều yếu tố
        # 1. Feature coverage
        feature_coverage = len(feature_names) / 30.0
        # 2. News availability (có tin tức gần đây không?)
        news_availability = min(1.0, recent_news_count / max(10, horizon_hours // 4))
        # 3. Balance between news and tech (model có đa dạng features không?)
        news_feat_ratio = len(news_features) / len(feature_names) if feature_names else 0
        balance_score = 1.0 - abs(news_feat_ratio - 0.3)  # Ideal: 30% news features
        
        # Combine scores
        confidence = (feature_coverage * 0.4 + news_availability * 0.3 + balance_score * 0.3)
        
        return {
            "prediction_horizon": f"{horizon_hours}h",
            "predicted_price": str(predicted_price),
            "current_price": str(current_price),
            "predicted_change_pct": str((predicted_price - current_price) / current_price * 100),
            "confidence_score": str(confidence),
            
            # Metadata
            "metadata": {
                "model_key": model_key,
                "features_count": len(feature_names),
                "news_features_count": len(news_features),
                "tech_features_count": len(tech_features),
                "recent_news_analyzed": recent_news_count
            },
            
            # Top influential features (tất cả)
            "top_influential_features": [f[0] for f in top_features[:5]],
            
            # XAI: Phân loại features
            "feature_analysis": {
                # News sentiment features (ảnh hưởng mạnh nhất)
                "top_news_features": [{"feature": f[0], "importance": float(f[1])} for f in top_news[:5]],
                # Technical indicators (ảnh hưởng mạnh nhất)
                "top_technical_indicators": [{"feature": f[0], "importance": float(f[1])} for f in top_tech[:5]],
                # SHAP contributions (đóng góp vào prediction này)
                "shap_contributions": explanation
            },
            
            # Top 3 news articles with their keywords
            "top_news_articles": top_news_articles,
            
            # Summary explanation
            "explanation": {
                "primary_factors": {
                    "most_important_feature": top_features[0][0] if top_features else None,
                    "is_news_important": bool(len(top_news) > 0 and top_news[0][1] > (top_tech[0][1] if top_tech else 0)),
                    "news_vs_technical": "News sentiment" if (top_news and top_news[0][1] > (top_tech[0][1] if top_tech else 0)) else "Technical indicators",
                    "news_influence_pct": f"{(sum(f[1] for f in top_news) / sum(f[1] for f in top_features) * 100) if top_features else 0:.1f}%"
                }
            }
        }
