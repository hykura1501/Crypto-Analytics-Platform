"""
Redis cache wrapper for AI service predictions
"""
import json
import logging
import os
import redis
from typing import Optional, Dict, Any
from functools import wraps
import hashlib

logger = logging.getLogger(__name__)

class RedisCache:
    def __init__(self):
        redis_host = os.getenv("REDIS_HOST", "redis")
        redis_port = int(os.getenv("REDIS_PORT", "6379"))
        redis_db = int(os.getenv("REDIS_DB", "0"))
        
        try:
            self.client = redis.Redis(
                host=redis_host,
                port=redis_port,
                db=redis_db,
                decode_responses=True,
                socket_connect_timeout=5,
                socket_timeout=5
            )
            # Test connection
            self.client.ping()
            logger.info(f"✅ Redis cache connected: {redis_host}:{redis_port}")
            self.enabled = True
        except Exception as e:
            logger.warning(f"⚠️ Redis cache unavailable: {e}. Cache disabled.")
            self.client = None
            self.enabled = False
    
    def get(self, key: str) -> Optional[Dict[str, Any]]:
        """Get value from cache"""
        if not self.enabled:
            return None
        
        try:
            value = self.client.get(key)
            if value:
                return json.loads(value)
            return None
        except Exception as e:
            logger.error(f"Redis get error for key {key}: {e}")
            return None
    
    def set(self, key: str, value: Dict[str, Any], ttl: int = 300) -> bool:
        """Set value in cache with TTL (seconds)"""
        if not self.enabled:
            return False
        
        try:
            self.client.setex(key, ttl, json.dumps(value))
            return True
        except Exception as e:
            logger.error(f"Redis set error for key {key}: {e}")
            return False
    
    def delete(self, key: str) -> bool:
        """Delete key from cache"""
        if not self.enabled:
            return False
        
        try:
            self.client.delete(key)
            return True
        except Exception as e:
            logger.error(f"Redis delete error for key {key}: {e}")
            return False
    
    def delete_pattern(self, pattern: str) -> int:
        """Delete all keys matching pattern"""
        if not self.enabled:
            return 0
        
        try:
            keys = self.client.keys(pattern)
            if keys:
                return self.client.delete(*keys)
            return 0
        except Exception as e:
            logger.error(f"Redis delete_pattern error for pattern {pattern}: {e}")
            return 0

# Global cache instance
cache = RedisCache()

def cache_key(symbol: str, horizon_hours: int) -> str:
    """Generate cache key for prediction"""
    return f"prediction:{symbol}:{horizon_hours}h"

def get_cache_ttl(horizon_hours: int) -> int:
    """
    Get cache TTL based on prediction horizon
    - 4h horizon: 5 minutes (300s) - giá crypto thay đổi nhanh
    - 8h horizon: 10 minutes (600s)
    - 24h horizon: 30 minutes (1800s) - dự đoán dài hạn ít thay đổi
    """
    if horizon_hours <= 4:
        return 300  # 5 minutes
    elif horizon_hours <= 8:
        return 600  # 10 minutes
    elif horizon_hours <= 24:
        return 1800  # 30 minutes
    else:
        return 3600  # 1 hour for longer horizons
