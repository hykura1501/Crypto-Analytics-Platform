"""
Playwright-based scraper for JavaScript-heavy pages
"""
import logging
from typing import Optional, Dict
from playwright.sync_api import sync_playwright, Page, Browser
import time

logger = logging.getLogger(__name__)

class PlaywrightScraper:
    """Scraper using Playwright for JS rendering"""
    
    def __init__(self, headless: bool = True):
        self.headless = headless
        self.browser: Optional[Browser] = None
        self.playwright = None
    
    def __enter__(self):
        """Context manager entry"""
        self.playwright = sync_playwright().start()
        self.browser = self.playwright.chromium.launch(headless=self.headless)
        logger.info("Playwright browser launched")
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit"""
        if self.browser:
            self.browser.close()
        if self.playwright:
            self.playwright.stop()
        logger.info("Playwright browser closed")
    
    def scrape_page(self, url: str, wait_for: str = "networkidle") -> Optional[Dict]:
        """
        Scrape a page with JavaScript rendering
        
        Args:
            url: URL to scrape
            wait_for: Wait condition (networkidle, load, domcontentloaded)
        
        Returns:
            Dict with title, html, text
        """
        if not self.browser:
            logger.error("Browser not initialized. Use context manager.")
            return None
        
        try:
            page = self.browser.new_page()
            
            # Set user agent
            page.set_extra_http_headers({
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
            })
            
            # Navigate and wait
            page.goto(url, wait_until=wait_for, timeout=30000)
            
            # Optional: Wait for specific selectors if needed
            # page.wait_for_selector("article", timeout=5000)
            
            # Extract content
            title = page.title()
            html = page.content()
            
            # Try to extract main content
            text = self._extract_main_content(page)
            
            page.close()
            
            return {
                "title": title,
                "html": html,
                "text": text,
                "url": url
            }
            
        except Exception as e:
            logger.error(f"Error scraping {url} with Playwright: {e}")
            return None
    
    def _extract_main_content(self, page: Page) -> str:
        """
        Extract main article content from page
        Try multiple selectors
        """
        selectors = [
            "article",
            ".article-content",
            ".post-content", 
            ".entry-content",
            "main",
            "#content"
        ]
        
        for selector in selectors:
            try:
                element = page.query_selector(selector)
                if element:
                    text = element.inner_text()
                    if len(text) > 200:  # Minimum content length
                        return text
            except:
                continue
        
        # Fallback: get body text
        try:
            return page.query_selector("body").inner_text()
        except:
            return ""

# Usage example:
# with PlaywrightScraper() as scraper:
#     result = scraper.scrape_page("https://example.com")
