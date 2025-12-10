import logging
from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.interval import IntervalTrigger
from app.database import SessionLocal
from app.crawler import news_crawler
from app.config import settings

logger = logging.getLogger(__name__)

scheduler = AsyncIOScheduler()

def run_crawler_job():
    """Job to run the crawler"""
    logger.info("🕐 Starting scheduled crawl job...")
    
    db = SessionLocal()
    try:
        total_saved = news_crawler.crawl_and_save(db)
        logger.info(f"Crawl job completed. Saved {total_saved} articles.")
    except Exception as e:
        logger.error(f"Error in crawler job: {e}")
    finally:
        db.close()

def start_scheduler():
    """Start the APScheduler"""
    logger.info(f"Starting scheduler (interval: {settings.crawl_interval_minutes} minutes)")
    
    # Add job to run every N minutes
    scheduler.add_job(
        run_crawler_job,
        trigger=IntervalTrigger(minutes=settings.crawl_interval_minutes),
        id="news_crawler",
        name="News Crawler Job",
        replace_existing=True
    )
    
    # Run immediately on startup
    scheduler.add_job(
        run_crawler_job,
        id="initial_crawl",
        name="Initial Crawl",
        replace_existing=True
    )
    
    scheduler.start()
    logger.info("✅ Scheduler started")

def stop_scheduler():
    """Stop the scheduler"""
    if scheduler.running:
        scheduler.shutdown()
        logger.info("Scheduler stopped")
