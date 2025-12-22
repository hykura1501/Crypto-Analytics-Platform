import feedparser
import logging
from typing import List, Dict, Optional
from datetime import datetime
from dateutil import parser as date_parser
from newspaper import Article as NewspaperArticle
from sqlalchemy.orm import Session

from app.models import Article
from app.sources import get_all_sources
from app.kafka_producer import news_producer
from app.utils.dedup import normalize_url, simhash_hex, is_duplicate
from app.utils.entities import extract_entities, classify_event
from app.utils.rate_limit import rate_limiter, robots_checker

logger = logging.getLogger(__name__)

class EnhancedNewsCrawler:
    """Enhanced news crawler with entity extraction and deduplication"""
    
    def __init__(self):
        self.sources = get_all_sources()
    
    def parse_rss_feed(self, source_name: str, source_config: Dict) -> List[Dict]:
        """Parse RSS feed and return list of article URLs"""
        try:
            rss_url = source_config.get("rss_url")
            logger.info(f"Fetching RSS feed from {source_name}: {rss_url}")
            
            # Set custom User-Agent for feedparser (RSS feeds should be accessible)
            import feedparser
            feedparser.USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
            
            feed = feedparser.parse(rss_url)
            
            articles = []
            for entry in feed.entries[:20]:  # Limit to prevent overload
                article_data = {
                    "source_id": source_name,
                    "url": entry.link,
                    "title": entry.get("title", ""),
                    "published_at": self._parse_date(entry.get("published")),
                    "language": source_config.get("language", "en")
                }
                articles.append(article_data)
            
            logger.info(f"Found {len(articles)} articles from {source_name}")
            return articles
        except Exception as e:
            logger.error(f"Error parsing RSS feed for {source_name}: {e}")
            return []
    
    def extract_article_content(self, url: str, requires_js: bool = False) -> Optional[Dict]:
        """
        Extract article content using newspaper3k or Playwright
        
        Args:
            url: Article URL
            requires_js: If True, use Playwright for JS rendering
        """
        if requires_js:
            return self._extract_with_playwright(url)
        else:
            return self._extract_with_newspaper(url)
    
    def _extract_with_newspaper(self, url: str) -> Optional[Dict]:
        """Extract using newspaper3k (fast, for static pages)"""
        try:
            # Set User-Agent for newspaper3k
            import newspaper
            article = NewspaperArticle(url)
            article.config.browser_user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
            article.download()
            article.parse()
            
            # Clean and validate text
            text = article.text or ""
            if len(text) < 100:  # Too short, probably failed
                logger.warning(f"Content too short from {url}")
                return None
            
            return {
                "title": article.title or "",
                "content_text": text,
                "author": ", ".join(article.authors) if article.authors else None,
                "published_at": article.publish_date
            }
        except Exception as e:
            logger.error(f"Error extracting content from {url}: {e}")
            return None
    
    def _extract_with_playwright(self, url: str) -> Optional[Dict]:
        """Extract using Playwright (for JS-heavy pages)"""
        try:
            from app.utils.playwright_scraper import PlaywrightScraper
            
            with PlaywrightScraper() as scraper:
                result = scraper.scrape_page(url)
                
                if not result or len(result.get("text", "")) < 100:
                    logger.warning(f"Playwright extraction failed or too short: {url}")
                    return None
                
                return {
                    "title": result.get("title", ""),
                    "content_text": result.get("text", ""),
                    "author": None,  # Playwright doesn't extract author
                    "published_at": None
                }
        except Exception as e:
            logger.error(f"Error with Playwright extraction from {url}: {e}")
            return None
    
    def _parse_date(self, date_string: str) -> Optional[datetime]:
        """Parse date string to datetime"""
        if not date_string:
            return None
        try:
            return date_parser.parse(date_string)
        except Exception as e:
            logger.debug(f"Error parsing date '{date_string}': {e}")
            return None
    
    def check_duplicate(self, db: Session, url: str, content_hash: str) -> bool:
        """
        Check if article is duplicate
        Returns: True if duplicate found
        """
        url_norm = normalize_url(url)
        
        # Check by normalized URL
        existing_url = db.query(Article).filter(
            Article.url_normalized == url_norm
        ).first()
        
        if existing_url:
            logger.debug(f"Duplicate URL found: {url}")
            return True
        
        # Check by content hash (SimHash)
        existing_hash = db.query(Article).filter(
            Article.content_hash != None
        ).all()
        
        for article in existing_hash:
            if is_duplicate(content_hash, article.content_hash):
                logger.debug(f"Duplicate content found (SimHash): {url}")
                return True
        
        return False
    
    def crawl_and_save(self, db: Session) -> int:
        """Main crawling logic - fetch, extract, enhance, and save"""
        total_saved = 0
        
        for source_name, source_config in self.sources.items():
            logger.info(f"📰 Crawling {source_name}...")
            
            # Get articles from RSS feed
            rss_articles = self.parse_rss_feed(source_name, source_config)
            
            logger.info(f"Processing {len(rss_articles)} articles from {source_name}")
            
            for idx, rss_article in enumerate(rss_articles, 1):
                url = rss_article["url"]
                
                try:
                    logger.info(f"[{idx}/{len(rss_articles)}] Processing {source_name}: {url[:80]}...")
                    
                    # Check robots.txt (skip for RSS feed domains - RSS feeds are public)
                    if not robots_checker.can_fetch(url, skip_for_rss=True):
                        logger.info(f"Skipping (robots.txt): {url}")
                        continue
                    
                    # Rate limiting
                    rate_limiter.wait(url)
                    
                    # Extract full content
                    requires_js = source_config.get("requires_js", False)
                    extracted = self.extract_article_content(url, requires_js=requires_js)
                    if not extracted or not extracted.get("content_text"):
                        logger.warning(f"Failed to extract content from {source_name}: {url}")
                        # Try fallback: if Playwright failed, try newspaper3k
                        if requires_js:
                            logger.info(f"Trying newspaper3k fallback for {url}")
                            extracted = self._extract_with_newspaper(url)
                            if not extracted or not extracted.get("content_text"):
                                logger.warning(f"Fallback also failed: {url}")
                                continue
                        else:
                            continue
                    
                    logger.debug(f"Successfully extracted content from {source_name}: {len(extracted.get('content_text', ''))} chars")
                    
                    # Combine data
                    title = extracted.get("title") or rss_article.get("title")
                    content_text = extracted["content_text"]
                    author = extracted.get("author")
                    published_at = extracted.get("published_at") or rss_article.get("published_at")
                    language = rss_article.get("language", "en")
                    
                    # Generate deduplication fingerprints
                    url_normalized = normalize_url(url)
                    content_hash = simhash_hex(content_text)
                    
                    # Check for duplicates
                    if self.check_duplicate(db, url, content_hash):
                        logger.debug(f"Skipping duplicate: {url}")
                        continue
                    
                    # Extract entities
                    entities = extract_entities(content_text, title)
                    
                    # Classify event type
                    event_type = classify_event(content_text, title)
                    
                    # Calculate confidence
                    confidence = self._calculate_confidence(
                        title, content_text, entities
                    )
                    
                    # Create article
                    article = Article(
                        source_id=source_name,
                        url=url,
                        url_normalized=url_normalized,
                        content_hash=content_hash,
                        title=title,
                        author=author,
                        published_at=published_at,
                        content_text=content_text,
                        language=language,
                        entities=entities,
                        event_type=event_type,
                        confidence=confidence
                    )
                    
                    db.add(article)
                    db.commit()
                    db.refresh(article)
                    
                    logger.info(f"✅ Saved: {title[:60]}... [entities: {len(entities.get('tickers', []))} tickers]")
                    
                    # Publish to Kafka
                    news_producer.publish_news(
                        news_id=article.id,
                        title=article.title,
                        content=article.content_text
                    )
                    
                    total_saved += 1
                    
                except Exception as e:
                    logger.error(f"Error processing article {url}: {e}")
                    db.rollback()
                    continue
        
        logger.info(f"🎉 Crawling completed. Total saved: {total_saved}")
        return total_saved
    
    def _calculate_confidence(self, title: str, content: str, entities: Dict) -> float:
        """
        Calculate extraction confidence score
        Factors: content length, entity count, title quality
        """
        score = 1.0
        
        # Penalize short content
        if len(content) < 500:
            score -= 0.2
        elif len(content) < 200:
            score -= 0.4
        
        # Reward entity extraction
        total_entities = sum(len(v) for v in entities.values())
        if total_entities == 0:
            score -= 0.1
        
        # Check title quality
        if not title or len(title) < 10:
            score -= 0.2
        
        return max(0.0, min(1.0, score))

# Singleton instance
news_crawler = EnhancedNewsCrawler()
