"""
Rate limiting and robots.txt compliance utilities
"""
import time
import logging
from typing import Dict, Optional
from urllib.parse import urlparse, urljoin
from urllib.robotparser import RobotFileParser
import requests

logger = logging.getLogger(__name__)

class RateLimiter:
    """Simple rate limiter with per-domain tracking"""
    
    def __init__(self, default_delay: float = 2.0):
        self.default_delay = default_delay
        self.last_request: Dict[str, float] = {}
    
    def wait(self, url: str):
        """Wait if necessary before making request to domain"""
        domain = urlparse(url).netloc
        
        if domain in self.last_request:
            elapsed = time.time() - self.last_request[domain]
            if elapsed < self.default_delay:
                sleep_time = self.default_delay - elapsed
                logger.debug(f"Rate limiting: sleeping {sleep_time:.2f}s for {domain}")
                time.sleep(sleep_time)
        
        self.last_request[domain] = time.time()

class RobotsTxtChecker:
    """Robots.txt compliance checker"""
    
    def __init__(self, user_agent: str = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"):
        self.user_agent = user_agent
        self.parsers: Dict[str, RobotFileParser] = {}
        # Domains that allow RSS feeds (skip robots.txt check)
        # Normalized domains (without www) for easier matching
        self.rss_allowed_domains = {
            "cointelegraph.com",
            "coindesk.com"  # This matches both www.coindesk.com and coindesk.com
        }
    
    def can_fetch(self, url: str, skip_for_rss: bool = False) -> bool:
        """
        Check if URL can be fetched according to robots.txt
        
        Args:
            url: URL to check
            skip_for_rss: If True, skip robots.txt check for RSS feed domains
        
        Returns:
            True if allowed, False if disallowed
        """
        try:
            parsed = urlparse(url)
            domain = parsed.netloc.lower()
            
            # Skip robots.txt check for RSS feed domains (RSS feeds are meant to be public)
            if skip_for_rss:
                # Normalize domain (remove www. prefix for comparison)
                domain_normalized = domain.replace('www.', '')
                # Check if normalized domain matches any RSS allowed domain
                if domain_normalized in self.rss_allowed_domains:
                    logger.info(f"✅ Skipping robots.txt check for RSS domain: {domain}")
                    return True
            
            domain_url = f"{parsed.scheme}://{parsed.netloc}"
            
            # Get or create parser for this domain
            if domain_url not in self.parsers:
                self._load_robots_txt(domain_url)
            
            parser = self.parsers.get(domain_url)
            if not parser:
                # If couldn't load robots.txt, allow by default
                return True
            
            # Check if allowed
            allowed = parser.can_fetch(self.user_agent, url)
            
            if not allowed:
                logger.warning(f"robots.txt disallows: {url}")
            
            return allowed
            
        except Exception as e:
            logger.error(f"Error checking robots.txt for {url}: {e}")
            # On error, allow by default (fail open)
            return True
    
    def _load_robots_txt(self, domain: str):
        """Load robots.txt for domain"""
        try:
            robots_url = urljoin(domain, "/robots.txt")
            logger.debug(f"Loading robots.txt from {robots_url}")
            
            parser = RobotFileParser()
            parser.set_url(robots_url)
            parser.read()
            
            self.parsers[domain] = parser
            logger.info(f"Loaded robots.txt for {domain}")
            
        except Exception as e:
            logger.warning(f"Couldn't load robots.txt for {domain}: {e}")
            # Store None to avoid repeated attempts
            self.parsers[domain] = None
    
    def get_crawl_delay(self, url: str) -> Optional[float]:
        """
        Get crawl delay from robots.txt
        
        Returns:
            Delay in seconds, or None if not specified
        """
        try:
            parsed = urlparse(url)
            domain = f"{parsed.scheme}://{parsed.netloc}"
            
            if domain not in self.parsers:
                self._load_robots_txt(domain)
            
            parser = self.parsers.get(domain)
            if parser:
                delay = parser.crawl_delay(self.user_agent)
                return delay
            
        except Exception as e:
            logger.error(f"Error getting crawl delay: {e}")
        
        return None

# Global instances
rate_limiter = RateLimiter(default_delay=2.0)
robots_checker = RobotsTxtChecker(user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
