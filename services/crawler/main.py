#!/usr/bin/env python3
"""
News Crawler Service
Tự động thu thập tin tức từ nhiều nguồn khác nhau
Có khả năng tự động học cấu trúc HTML của mỗi trang
"""

import os
import sys
import time
import schedule
import psycopg2
from bs4 import BeautifulSoup
import requests
from datetime import datetime
from urllib.parse import urljoin, urlparse
import json
from dotenv import load_dotenv

load_dotenv()

class NewsCrawler:
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
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })
    
    def get_db_connection(self, max_retries=5):
        """Kết nối database với retry logic"""
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
    
    def learn_structure(self, html_content, sample_data):
        """
        Tự động học cấu trúc HTML của trang
        Tìm các selector phù hợp để trích xuất thông tin
        """
        soup = BeautifulSoup(html_content, 'html.parser')
        
        # Tìm title selector
        title_selectors = [
            'h1', 'h2', 'h3',
            '[class*="title"]', '[class*="headline"]',
            'article h1', 'article h2'
        ]
        
        # Tìm content selector
        content_selectors = [
            'article', '[class*="content"]', '[class*="body"]',
            'main', '.post-content', '.article-body'
        ]
        
        # Tìm link selector
        link_selectors = [
            'a[href]', 'h2 a', 'h3 a', '.title a'
        ]
        
        return {
            'title_selectors': title_selectors,
            'content_selectors': content_selectors,
            'link_selectors': link_selectors
        }
    
    def extract_news(self, url, selectors):
        """
        Trích xuất tin tức từ URL với các selector đã cho
        """
        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            soup = BeautifulSoup(response.content, 'html.parser')
            
            # Extract title
            title = None
            for sel in selectors.get('title_selectors', []):
                element = soup.select_one(sel)
                if element:
                    title = element.get_text(strip=True)
                    break
            
            # Extract content
            content = None
            for sel in selectors.get('content_selectors', []):
                element = soup.select_one(sel)
                if element:
                    content = element.get_text(strip=True)
                    break
            
            # Extract links
            links = []
            for sel in selectors.get('link_selectors', []):
                elements = soup.select(sel)
                for elem in elements[:10]:  # Limit to 10 links
                    href = elem.get('href', '')
                    if href:
                        full_url = urljoin(url, href)
                        links.append({
                            'url': full_url,
                            'title': elem.get_text(strip=True)
                        })
            
            return {
                'title': title,
                'content': content,
                'links': links,
                'source_url': url
            }
        
        except Exception as e:
            print(f"Error extracting from {url}: {e}")
            return None
    
    def crawl_source(self, source_id):
        """
        Crawl tin tức từ một nguồn cụ thể
        """
        conn = self.get_db_connection()
        cur = conn.cursor()
        
        try:
            # Lấy thông tin source
            cur.execute("""
                SELECT id, name, url, title_selector, content_selector, 
                       date_selector, link_selector, selector_type
                FROM news_sources 
                WHERE id = %s AND status = 'active'
            """, (source_id,))
            
            source = cur.fetchone()
            if not source:
                print(f"Source {source_id} not found or inactive")
                return
            
            source_id, name, url, title_sel, content_sel, date_sel, link_sel, sel_type = source
            
            print(f"Crawling {name} from {url}")
            
            # Fetch page
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            
            # Extract news items
            soup = BeautifulSoup(response.content, 'html.parser')
            
            # Find article links
            article_links = []
            if link_sel:
                links = soup.select(link_sel)
                for link in links[:20]:  # Limit to 20 articles
                    href = link.get('href', '')
                    if href:
                        full_url = urljoin(url, href)
                        if full_url not in [a['url'] for a in article_links]:
                            article_links.append({
                                'url': full_url,
                                'title': link.get_text(strip=True)
                            })
            
            # Process each article
            for article in article_links:
                try:
                    # Check if already exists
                    cur.execute("SELECT id FROM news WHERE url = %s", (article['url'],))
                    if cur.fetchone():
                        continue
                    
                    # Fetch article content
                    article_resp = self.session.get(article['url'], timeout=30)
                    article_soup = BeautifulSoup(article_resp.content, 'html.parser')
                    
                    # Extract title
                    title = article['title']
                    if title_sel:
                        title_elem = article_soup.select_one(title_sel)
                        if title_elem:
                            title = title_elem.get_text(strip=True)
                    
                    # Extract content
                    content = None
                    if content_sel:
                        content_elem = article_soup.select_one(content_sel)
                        if content_elem:
                            content = content_elem.get_text(strip=True)[:5000]  # Limit content
                    
                    # Extract date
                    published_at = datetime.now()
                    if date_sel:
                        date_elem = article_soup.select_one(date_sel)
                        if date_elem:
                            date_text = date_elem.get_text(strip=True)
                            # Try to parse date (simplified)
                            try:
                                published_at = datetime.strptime(date_text[:19], '%Y-%m-%d %H:%M:%S')
                            except:
                                pass
                    
                    # Insert into database
                    cur.execute("""
                        INSERT INTO news (source_id, title, content, url, published_at)
                        VALUES (%s, %s, %s, %s, %s)
                        ON CONFLICT (url) DO NOTHING
                        RETURNING id
                    """, (source_id, title, content, article['url'], published_at))
                    
                    news_id = cur.fetchone()
                    if news_id:
                        print(f"Saved: {title[:50]}...")
                        conn.commit()
                    
                    time.sleep(1)  # Be polite
                
                except Exception as e:
                    print(f"Error processing article {article['url']}: {e}")
                    continue
            
            # Update last_crawled_at
            cur.execute("""
                UPDATE news_sources 
                SET last_crawled_at = NOW() 
                WHERE id = %s
            """, (source_id,))
            conn.commit()
            
            print(f"Completed crawling {name}")
        
        except Exception as e:
            print(f"Error crawling source {source_id}: {e}")
            conn.rollback()
        
        finally:
            cur.close()
            conn.close()
    
    def crawl_all_sources(self):
        """Crawl tất cả các nguồn active"""
        conn = self.get_db_connection()
        cur = conn.cursor()
        
        try:
            cur.execute("SELECT id FROM news_sources WHERE status = 'active'")
            source_ids = [row[0] for row in cur.fetchall()]
            
            for source_id in source_ids:
                self.crawl_source(source_id)
                time.sleep(5)  # Wait between sources
        
        finally:
            cur.close()
            conn.close()
    
    def run_scheduler(self):
        """Chạy crawler theo lịch"""
        # Crawl mỗi 30 phút
        schedule.every(30).minutes.do(self.crawl_all_sources)
        
        # Crawl ngay lập tức
        self.crawl_all_sources()
        
        print("Crawler scheduler started. Running every 30 minutes.")
        
        while True:
            schedule.run_pending()
            time.sleep(60)


if __name__ == '__main__':
    crawler = NewsCrawler()
    
    if len(sys.argv) > 1:
        if sys.argv[1] == '--source':
            source_id = int(sys.argv[2])
            crawler.crawl_source(source_id)
        elif sys.argv[1] == '--all':
            crawler.crawl_all_sources()
    else:
        # Run scheduler
        crawler.run_scheduler()

