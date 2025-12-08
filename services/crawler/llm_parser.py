"""
AI-Enhanced Crawler with LLM Parser
Tự động học cấu trúc HTML và trích xuất thông tin bằng LLM
"""

import os
import json
from typing import Dict, Optional, List
from bs4 import BeautifulSoup
import openai
from anthropic import Anthropic

class LLMParser:
    """Parse HTML content using LLM (GPT-4o-mini or Claude)"""
    
    def __init__(self):
        self.openai_key = os.getenv('OPENAI_API_KEY')
        self.anthropic_key = os.getenv('ANTHROPIC_API_KEY')
        self.use_openai = bool(self.openai_key)
        self.use_anthropic = bool(self.anthropic_key)
        
        if self.use_openai:
            openai.api_key = self.openai_key
        
        if self.use_anthropic:
            self.anthropic_client = Anthropic(api_key=self.anthropic_key)
    
    def extract_with_openai(self, html_content: str, url: str) -> Optional[Dict]:
        """Extract news data using OpenAI GPT-4o-mini"""
        if not self.use_openai:
            return None
        
        # Clean HTML to reduce tokens
        soup = BeautifulSoup(html_content, 'html.parser')
        
        # Remove scripts, styles, and navigation
        for tag in soup(['script', 'style', 'nav', 'footer', 'header', 'aside']):
            tag.decompose()
        
        clean_html = str(soup)[:10000]  # Limit to 10k chars
        
        prompt = f"""Extract the following information from this news article HTML:

1. Title (main headline of the article)
2. Published date (in ISO format YYYY-MM-DD HH:MM:SS, estimate if not exact)
3. Author name (or "Unknown" if not found)
4. Main content (article body text only, without ads or navigation)

HTML snippet:
{clean_html}

URL: {url}

Return ONLY valid JSON in this exact format:
{{"title": "string", "date": "YYYY-MM-DD HH:MM:SS", "author": "string", "content": "string"}}

If any field cannot be found, use empty string for title/author/content, and current date for date.
"""
        
        try:
            response = openai.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "system", "content": "You are an expert at extracting structured data from HTML. Always return valid JSON."},
                    {"role": "user", "content": prompt}
                ],
                response_format={"type": "json_object"},
                temperature=0,
                max_tokens=1000
            )
            
            result = json.loads(response.choices[0].message.content)
            return result
        
        except Exception as e:
            print(f"OpenAI extraction error: {e}")
            return None
    
    def extract_with_anthropic(self, html_content: str, url: str) -> Optional[Dict]:
        """Extract news data using Anthropic Claude"""
        if not self.use_anthropic:
            return None
        
        soup = BeautifulSoup(html_content, 'html.parser')
        for tag in soup(['script', 'style', 'nav', 'footer', 'header', 'aside']):
            tag.decompose()
        
        clean_html = str(soup)[:10000]
        
        prompt = f"""Extract news article information from this HTML:

HTML:
{clean_html}

URL: {url}

Return JSON with: title, date (YYYY-MM-DD HH:MM:SS), author, content"""
        
        try:
            response = self.anthropic_client.messages.create(
                model="claude-3-5-sonnet-20241022",
                max_tokens=1000,
                temperature=0,
                messages=[{
                    "role": "user",
                    "content": prompt
                }]
            )
            
            # Extract JSON from response
            content = response.content[0].text
            
            # Try to find JSON in the response
            start_idx = content.find('{')
            end_idx = content.rfind('}') + 1
            if start_idx != -1 and end_idx > start_idx:
                json_str = content[start_idx:end_idx]
                result = json.loads(json_str)
                return result
            
            return None
        
        except Exception as e:
            print(f"Anthropic extraction error: {e}")
            return None
    
    def extract(self, html_content: str, url: str) -> Optional[Dict]:
        """Try extraction with available LLM providers"""
        # Try OpenAI first (cheaper and faster)
        if self.use_openai:
            result = self.extract_with_openai(html_content, url)
            if result and self.validate_result(result):
                return result
        
        # Fallback to Anthropic
        if self.use_anthropic:
            result = self.extract_with_anthropic(html_content, url)
            if result and self.validate_result(result):
                return result
        
        return None
    
    def validate_result(self, result: Dict) -> bool:
        """Validate extracted data"""
        if not result:
            return False
        
        # Check required fields
        if not result.get('title') or len(result['title']) < 10:
            return False
        
        if not result.get('content') or len(result['content']) < 100:
            return False
        
        return True


class SmartCrawler:
    """
    Crawler thông minh với fallback mechanism
    1. Thử CSS selectors trước (nhanh, miễn phí)
    2. Nếu fail, dùng LLM parser
    """
    
    def __init__(self, session, db_connection):
        self.session = session
        self.db_conn = db_connection
        self.llm_parser = LLMParser()
        self.stats = {
            'css_success': 0,
            'llm_success': 0,
            'total_failed': 0
        }
    
    def extract_with_css(self, soup: BeautifulSoup, selectors: Dict) -> Optional[Dict]:
        """Try to extract using CSS selectors"""
        try:
            result = {}
            
            # Extract title
            for sel in selectors.get('title_selectors', []):
                element = soup.select_one(sel)
                if element:
                    result['title'] = element.get_text(strip=True)
                    break
            
            # Extract content
            for sel in selectors.get('content_selectors', []):
                element = soup.select_one(sel)
                if element:
                    result['content'] = element.get_text(strip=True)
                    break
            
            # Extract date
            date_sel = selectors.get('date_selector')
            if date_sel:
                element = soup.select_one(date_sel)
                if element:
                    result['date'] = element.get_text(strip=True)
            
            # Extract author
            result['author'] = 'Unknown'
            for sel in ['[rel="author"]', '.author', '[class*="author"]']:
                element = soup.select_one(sel)
                if element:
                    result['author'] = element.get_text(strip=True)
                    break
            
            return result if result.get('title') and result.get('content') else None
        
        except Exception as e:
            print(f"CSS extraction error: {e}")
            return None
    
    def validate_result(self, result: Dict) -> bool:
        """Validate extracted result"""
        if not result:
            return False
        
        title = result.get('title', '')
        content = result.get('content', '')
        
        if len(title) < 10:
            return False
        
        if len(content) < 100:
            return False
        
        # Check if content is not just navigation/menu text
        if 'cookie' in content.lower()[:200] or 'subscribe' in content.lower()[:200]:
            return False
        
        return True
    
    def crawl_article(self, url: str, selectors: Dict) -> Optional[Dict]:
        """
        Crawl article with smart fallback
        1. Try CSS selectors (fast)
        2. If fail, use LLM (slow but reliable)
        """
        try:
            # Fetch page
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            soup = BeautifulSoup(response.content, 'html.parser')
            
            # Try CSS selectors first
            result = self.extract_with_css(soup, selectors)
            
            if self.validate_result(result):
                print(f"✅ CSS extraction succeeded: {url}")
                self.stats['css_success'] += 1
                return result
            
            # Fallback to LLM
            print(f"⚠️  CSS failed, trying LLM for: {url}")
            result = self.llm_parser.extract(str(soup), url)
            
            if self.validate_result(result):
                print(f"✅ LLM extraction succeeded: {url}")
                self.stats['llm_success'] += 1
                
                # Learn from LLM: update selectors for future (optional)
                self.learn_selectors_from_llm(url, soup, result)
                
                return result
            
            print(f"❌ Both CSS and LLM failed: {url}")
            self.stats['total_failed'] += 1
            return None
        
        except Exception as e:
            print(f"Error crawling {url}: {e}")
            self.stats['total_failed'] += 1
            return None
    
    def learn_selectors_from_llm(self, url: str, soup: BeautifulSoup, extracted_data: Dict):
        """
        Learn CSS selectors from successful LLM extraction
        Find which elements contain the extracted data
        """
        try:
            # Find title element
            title = extracted_data.get('title', '')
            if title:
                for tag in ['h1', 'h2', 'h3']:
                    elements = soup.find_all(tag)
                    for elem in elements:
                        if title in elem.get_text():
                            selector = self.generate_selector(elem)
                            print(f"  💡 Learned title selector: {selector}")
                            break
        
        except Exception as e:
            print(f"Error learning selectors: {e}")
    
    def generate_selector(self, element) -> str:
        """Generate CSS selector for an element"""
        if element.get('id'):
            return f"#{element['id']}"
        
        if element.get('class'):
            classes = element['class']
            return f"{element.name}.{'.'.join(classes[:2])}"
        
        return element.name
    
    def print_stats(self):
        """Print crawling statistics"""
        total = self.stats['css_success'] + self.stats['llm_success'] + self.stats['total_failed']
        if total == 0:
            return
        
        print("\n📊 Crawling Statistics:")
        print(f"   CSS Success: {self.stats['css_success']} ({self.stats['css_success']/total*100:.1f}%)")
        print(f"   LLM Success: {self.stats['llm_success']} ({self.stats['llm_success']/total*100:.1f}%)")
        print(f"   Failed: {self.stats['total_failed']} ({self.stats['total_failed']/total*100:.1f}%)")
        print(f"   Total: {total}\n")


# Example usage
if __name__ == '__main__':
    import requests
    
    session = requests.Session()
    session.headers.update({
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
    })
    
    crawler = SmartCrawler(session, None)
    
    # Test URLs
    test_urls = [
        'https://cointelegraph.com/news/bitcoin-price-analysis',
        'https://www.coindesk.com/markets/latest-news',
    ]
    
    for url in test_urls:
        print(f"\nCrawling: {url}")
        result = crawler.crawl_article(url, {
            'title_selectors': ['h1', 'h2.title'],
            'content_selectors': ['article', '.post-content']
        })
        
        if result:
            print(f"Title: {result['title'][:50]}...")
            print(f"Content: {result['content'][:100]}...")
    
    crawler.print_stats()
