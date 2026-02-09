"""
Utility functions for content deduplication
"""
import hashlib
import re
from urllib.parse import urlparse, parse_qs, urlencode, urlunparse

def normalize_url(url: str) -> str:
    """
    Normalize URL for deduplication
    - Remove tracking parameters
    - Canonicalize domain
    - Sort query parameters
    - Remove fragments
    """
    parsed = urlparse(url)
    
    # Remove www prefix
    domain = parsed.netloc.replace('www.', '')
    
    # Remove tracking parameters
    tracking_params = {'utm_source', 'utm_medium', 'utm_campaign', 'utm_content', 'utm_term',
                      'ref', 'source', 'fbclid', 'gclid', '_ga'}
    
    query_params = parse_qs(parsed.query)
    filtered_params = {k: v for k, v in query_params.items() if k not in tracking_params}
    
    # Sort parameters
    sorted_query = urlencode(sorted(filtered_params.items()), doseq=True)
    
    # Reconstruct URL without fragment
    normalized = urlunparse((
        parsed.scheme,
        domain,
        parsed.path.rstrip('/'),  # Remove trailing slash
        parsed.params,
        sorted_query,
        ''  # No fragment
    ))
    
    return normalized.lower()

def simhash(text: str, hash_bits: int = 64) -> int:
    """
    Generate SimHash for near-duplicate detection
    
    Args:
        text: Input text
        hash_bits: Number of bits for hash (default 64)
    
    Returns:
        Integer hash value
    """
    # Tokenize and create shingles
    tokens = re.findall(r'\w+', text.lower())
    
    # Use 3-grams (shingles)
    shingles = [' '.join(tokens[i:i+3]) for i in range(len(tokens) - 2)]
    
    # Initialize feature vector
    v = [0] * hash_bits
    
    for shingle in shingles:
        # Hash shingle
        h = int(hashlib.md5(shingle.encode()).hexdigest(), 16)
        
        # Update feature vector
        for i in range(hash_bits):
            if h & (1 << i):
                v[i] += 1
            else:
                v[i] -= 1
    
    # Generate final hash
    fingerprint = 0
    for i in range(hash_bits):
        if v[i] > 0:
            fingerprint |= (1 << i)
    
    return fingerprint

def simhash_hex(text: str) -> str:
    """Generate SimHash and return as hex string"""
    hash_int = simhash(text)
    return format(hash_int, '016x')

def hamming_distance(hash1: int, hash2: int) -> int:
    """Calculate Hamming distance between two hashes"""
    x = hash1 ^ hash2
    distance = 0
    while x:
        distance += 1
        x &= x - 1
    return distance

def is_duplicate(hash1: str, hash2: str, threshold: int = 3) -> bool:
    """
    Check if two content hashes represent duplicate content
    
    Args:
        hash1: First hash (hex string)
        hash2: Second hash (hex string)  
        threshold: Maximum Hamming distance for duplicates (default 3)
    
    Returns:
        True if content is duplicate
    """
    h1 = int(hash1, 16)
    h2 = int(hash2, 16)
    return hamming_distance(h1, h2) <= threshold
