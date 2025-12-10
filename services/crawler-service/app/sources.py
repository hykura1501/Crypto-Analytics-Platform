"""
Source configurations for different news sites
"""

SOURCES = {
    "CoinDesk": {
        "type": "rss",
        "rss_url": "https://www.coindesk.com/arc/outboundfeeds/rss/",
        "language": "en",
        "requires_js": False
    },
    "CoinTelegraph": {
        "type": "rss",
        "rss_url": "https://cointelegraph.com/rss",
        "language": "en",
        "requires_js": True  # May need Playwright for full content
    },
    "VNExpress": {
        "type": "rss",
        "rss_url": "https://vnexpress.net/rss/kinh-doanh.rss",
        "language": "vi",
        "requires_js": False,
        "selectors": {
            "title": "h1.title-detail",
            "content": "article.fck_detail",
            "author": "p.author_mail"
        }
    },
    "VnEconomy": {
        "type": "rss",
        "rss_url": "https://vneconomy.vn/thi-truong-chung-khoan.rss",
        "language": "vi",
        "requires_js": False
    }
}

def get_all_sources():
    """Get all enabled sources"""
    return SOURCES

def get_source_config(source_name: str):
    """Get configuration for a specific source"""
    return SOURCES.get(source_name)
