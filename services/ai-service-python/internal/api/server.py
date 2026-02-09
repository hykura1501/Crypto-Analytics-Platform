import logging
from fastapi import Depends, FastAPI, HTTPException
from pydantic import BaseModel
from config import config
from internal.handler.rss_handler import RssHandler
from internal.handler.selector_handler import SelectorHandler
from internal.prediction import PredictionPipeline
from internal.sentiment.sentiment import Analyzer
from internal.db.database import db
from internal.auth_jwt import verify_prediction_access
from internal.cache import cache, cache_key, get_cache_ttl

logger = logging.getLogger(__name__)

class RssAnalysisRequest(BaseModel):
    xml_string: str

class SelectorAnalysisRequest(BaseModel):
    html_string: str

class TrainRequest(BaseModel):
    symbol: str = "BTCUSDT"
    horizon_hours: int = 4
    years: int = 2

class Server:
    def __init__(self, rss_handler: RssHandler, selector_handler: SelectorHandler, sentiment_handler: Analyzer):
        self.app = FastAPI()
        self.rss_handler = rss_handler
        self.selector_handler = selector_handler
        self.sentiment_handler = sentiment_handler
        self.prediction = PredictionPipeline(analyzer=self.sentiment_handler)
        self.db = db
        self.setup_routes()

    def setup_routes(self):
        @self.app.get("/api/v1/ai/health")
        async def health():
            return {"service": "ai-service", "status": "running"}

        @self.app.post("/api/v1/ai/rss/{source_id}")
        async def handle_rss(source_id: str, req: RssAnalysisRequest):
            msg = {
                "source_id": source_id,
                "xml_string": req.xml_string
            }
            response = self.rss_handler.handle(msg)
            if response is None:
                raise HTTPException(status_code=500, detail="RSS analysis failed")
            
            return {
                "message": "RSS analysis completed",
                "response": response
            }

        @self.app.post("/api/v1/ai/selector/{source_id}")
        async def handle_selector(source_id: str, req: SelectorAnalysisRequest):
            msg = {
                "source_id": source_id,
                "html_string": req.html_string
            }
            response = self.selector_handler.handle(msg)
            if response is None:
                raise HTTPException(status_code=500, detail="CSS selector analysis failed")
            
            return {
                "message": "CSS selector analysis completed",
                "response": response
            }
        
        @self.app.post("/api/v1/ai/prediction/train")
        async def train_model(req: TrainRequest, _=Depends(verify_prediction_access)):
            try:
                metrics = self.prediction.train(
                    symbol=req.symbol,
                    horizon_hours=req.horizon_hours,
                    years=req.years
                )
                
                # Invalidate cache for this symbol and horizon after training
                key = cache_key(req.symbol.upper(), req.horizon_hours)
                cache.delete(key)
                logger.info(f"🗑️ Invalidated cache for {req.symbol} {req.horizon_hours}h after training")
                
                return {"message": "Training completed", **metrics}
            except Exception as e:
                raise HTTPException(status_code=500, detail=str(e))
        
        @self.app.get("/api/v1/ai/prediction/predict/{symbol}/{horizon_hours}")
        async def predict(symbol: str, horizon_hours: int, _=Depends(verify_prediction_access)):
            try:
                # Check cache first
                key = cache_key(symbol.upper(), horizon_hours)
                cached_result = cache.get(key)
                
                if cached_result:
                    logger.info(f"✅ Cache HIT for {symbol} {horizon_hours}h prediction")
                    return {
                        "message": "Prediction completed (cached)",
                        "result": cached_result,
                        "cached": True
                    }
                
                # Cache miss - compute prediction
                logger.info(f"⏳ Cache MISS for {symbol} {horizon_hours}h prediction - computing...")
                result = self.prediction.predict(symbol=symbol, horizon_hours=horizon_hours)
                
                # Cache the result
                ttl = get_cache_ttl(horizon_hours)
                cache.set(key, result, ttl=ttl)
                logger.info(f"💾 Cached prediction for {symbol} {horizon_hours}h (TTL: {ttl}s)")
                
                return {
                    "message": "Prediction completed",
                    "result": result,
                    "cached": False
                }
            except ValueError as e:
                raise HTTPException(status_code=400, detail=str(e))
            except Exception as e:
                raise HTTPException(status_code=500, detail=str(e))

        @self.app.patch("/api/v1/ai/sentiments")
        async def update_sentiments():
            try:
                # Select all news from database
                with self.db.conn.cursor() as cur:
                    cur.execute("SELECT id, title, content_text, language FROM articles where id > 937 ORDER BY id")
                    news = cur.fetchall()
                    
                    updated_count = 0
                    for news_item in news:
                        news_id, title, content, language = news_item
                        
                        # Skip if both title and content are None/empty
                        if not title and not content:
                            print(f"Skipping news #{news_id}: no title or content")
                            continue
                        
                        # Combine title and content
                        text = (title or "") + " " + (content or "")
                        if not text.strip():
                            print(f"Skipping news #{news_id}: no text")
                            continue
                        
                        print(f"Analyzing sentiment for news #{news_id}: {text[:50]}...")

                        # Analyze sentiment (auto-detect language)
                        try:
                            top_keywords, sentiment_score = self.sentiment_handler.analyze(text, language, top_k=5)
                            
                            # Update sentiment score
                            cur.execute(
                                "UPDATE articles SET sentiment_score = %s WHERE id = %s",
                                (sentiment_score, news_id)
                            )
                            print(f"Updated sentiment score for news #{news_id}: {sentiment_score}")
                            self.db.conn.commit()
                            updated_count += 1
                        except Exception as e:
                            print(f"Error analyzing news #{news_id}: {e}")
                            # Skip this news and continue
                            continue
                    
                    # Commit once after all updates
                    self.db.conn.commit()
                    
                    return {
                        "message": "Sentiments updated",
                        "updated_count": updated_count,
                        "total_processed": len(news)
                    }
            except Exception as e:
                # Rollback on error
                if self.db.conn:
                    self.db.conn.rollback()
                raise HTTPException(status_code=500, detail=str(e))

    def run(self):
        import uvicorn
        uvicorn.run(self.app, host="0.0.0.0", port=config.api.port)
