import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.database import init_db
from app.routes import router
from app.scheduler import start_scheduler, stop_scheduler
from app.kafka_producer import news_producer
from app.config import settings

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan - startup and shutdown"""
    # Startup
    logger.info("🚀 Starting Crawler Service...")
    init_db()
    start_scheduler()
    yield
    # Shutdown
    logger.info("Shutting down Crawler Service...")
    stop_scheduler()
    news_producer.close()

app = FastAPI(
    title="Crypto News Crawler Service",
    description="Crawls crypto news from CoinDesk and CoinTelegraph",
    version="1.0.0",
    lifespan=lifespan
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(router)

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "crawler-service"}

@app.get("/")
def root():
    return {
        "service": "Crypto News Crawler",
        "version": "1.0.0",
        "endpoints": {
            "news": "/news",
            "health": "/health",
            "docs": "/docs"
        }
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=settings.server_port,
        reload=True
    )
