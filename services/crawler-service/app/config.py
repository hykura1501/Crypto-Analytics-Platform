import os
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # Database
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "postgres"
    db_password: str = "postgres"
    db_name: str = "crypto_db"
    
    # Kafka
    kafka_broker: str = "localhost:29092"
    kafka_topic: str = "news_new_article"
    
    # RSS Feeds
    coindesk_rss_url: str = "https://www.coindesk.com/arc/outboundfeeds/rss/"
    cointelegraph_rss_url: str = "https://cointelegraph.com/rss"
    
    # Crawler
    crawl_interval_minutes: int = 10
    max_articles_per_run: int = 20
    
    # Server
    server_port: int = 8083
    
    @property
    def database_url(self) -> str:
        return f"postgresql://{self.db_user}:{self.db_password}@{self.db_host}:{self.db_port}/{self.db_name}"
    
    class Config:
        env_file = ".env"
        case_sensitive = False

settings = Settings()
