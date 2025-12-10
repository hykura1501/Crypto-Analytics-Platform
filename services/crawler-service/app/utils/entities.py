"""
Entity extraction utilities
"""
import re
from typing import Dict, List, Set

# Crypto tickers and coins mapping
CRYPTO_TICKERS = {
    'BTC': 'Bitcoin',
    'ETH': 'Ethereum',
    'USDT': 'Tether',
    'BNB': 'Binance Coin',
    'SOL': 'Solana',
    'XRP': 'Ripple',
    'ADA': 'Cardano',
    'DOT': 'Polkadot',
    'DOGE': 'Dogecoin',
    'MATIC': 'Polygon',
    'AVAX': 'Avalanche',
    'UNI': 'Uniswap',
    'LINK': 'Chainlink',
    'ATOM': 'Cosmos',
    'LTC': 'Litecoin'
}

# Event keywords mapping
EVENT_KEYWORDS = {
    'listing': ['list', 'listing', 'niêm yết', 'ra mắt'],
    'security': ['hack', 'exploit', 'breach', 'attack', 'bị tấn công', 'lộ', 'mất an toàn'],
    'regulation': ['regulation', 'SEC', 'quy định', 'cấm', 'luật'],
    'partnership': ['partnership', 'collaborate', 'hợp tác', 'đối tác'],
    'launch': ['launch', 'mainnet', 'testnet', 'ra mắt', 'khởi chạy'],
    'acquisition': ['acquire', 'merger', 'mua lại', 'sáp nhập'],
    'funding': ['funding', 'raise', 'investment', 'gọi vốn', 'đầu tư']
}

def extract_tickers(text: str) -> List[str]:
    """
    Extract crypto tickers from text
    Patterns: $BTC, BTC/USD, #bitcoin, Bitcoin
    """
    tickers = set()
    text_upper = text.upper()
    
    # Pattern 1: $BTC
    dollar_tickers = re.findall(r'\$([A-Z]{3,5})\b', text)
    tickers.update(dollar_tickers)
    
    # Pattern 2: BTC/USD, BTC/USDT
    pair_tickers = re.findall(r'\b([A-Z]{3,5})/(?:USD|USDT)\b', text)
    tickers.update(pair_tickers)
    
    # Pattern 3: Check known tickers
    for ticker in CRYPTO_TICKERS.keys():
        if re.search(r'\b' + ticker + r'\b', text_upper):
            tickers.add(ticker)
    
    return sorted(list(tickers))

def extract_coins(text: str) -> List[str]:
    """Extract cryptocurrency names"""
    coins = set()
    text_lower = text.lower()
    
    # Check known coins
    for ticker, name in CRYPTO_TICKERS.items():
        if name.lower() in text_lower:
            coins.add(name)
    
    return sorted(list(coins))

def extract_organizations(text: str) -> List[str]:
    """
    Extract crypto-related organizations
    Simple pattern matching for common orgs
    """
    orgs = set()
    
    # Common crypto organizations
    known_orgs = [
        'Binance', 'Coinbase', 'FTX', 'Kraken', 'Huobi',
        'SEC', 'CFTC', 'Federal Reserve', 'Fed',
        'MicroStrategy', 'Tesla', 'BlackRock',
        'Ethereum Foundation', 'Bitcoin Foundation'
    ]
    
    for org in known_orgs:
        if org.lower() in text.lower():
            orgs.add(org)
    
    return sorted(list(orgs))

def extract_people(text: str) -> List[str]:
    """
    Extract crypto-related people
    Simple pattern matching for well-known figures
    """
    people = set()
    
    # Well-known crypto figures
    known_people = [
        'Satoshi Nakamoto', 'Vitalik Buterin', 'CZ', 'Changpeng Zhao',
        'Michael Saylor', 'Elon Musk', 'Gary Gensler',
        'Brian Armstrong', 'Sam Bankman-Fried', 'SBF',
        'Cathie Wood', 'Jack Dorsey'
    ]
    
    for person in known_people:
        if person.lower() in text.lower():
            people.add(person)
    
    return sorted(list(people))

def classify_event(text: str, title: str = "") -> str:
    """
    Classify article event type based on keywords
    Returns: event_type or None
    """
    combined_text = (title + " " + text).lower()
    
    # Score each event type
    scores = {}
    for event_type, keywords in EVENT_KEYWORDS.items():
        score = sum(1 for keyword in keywords if keyword in combined_text)
        if score > 0:
            scores[event_type] = score
    
    if not scores:
        return None
    
    # Return event type with highest score
    return max(scores, key=scores.get)

def extract_entities(text: str, title: str = "") -> Dict:
    """
    Extract all entities from text
    
    Returns:
        {
            "tickers": [...],
            "coins": [...],
            "people": [...],
            "orgs": [...]
        }
    """
    combined_text = title + " " + text
    
    return {
        "tickers": extract_tickers(combined_text),
        "coins": extract_coins(combined_text),
        "people": extract_people(combined_text),
        "orgs": extract_organizations(combined_text)
    }
