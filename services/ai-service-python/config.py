import os
from dotenv import load_dotenv

load_dotenv()

class DatabaseConfig:
    def __init__(self):
        self.host = os.getenv("DB_HOST", "postgres")
        self.port = os.getenv("DB_PORT", "5432")
        self.user = os.getenv("DB_USER", "postgres")
        self.password = os.getenv("DB_PASSWORD", "postgres")
        self.dbname = os.getenv("DB_NAME", "crypto_db")
        self.sslmode = os.getenv("DB_SSLMODE", "disable")

    def dsn(self):
        return f"host={self.host} port={self.port} user={self.user} password={self.password} dbname={self.dbname} sslmode={self.sslmode}"

class KafkaConfig:
    def __init__(self):
        self.broker = os.getenv("KAFKA_BROKER", "kafka:9092")
        self.group_id = os.getenv("KAFKA_GROUP_ID", "ai-service-group-v2")
        self.topics = [
            "news_new_article",
            "news_analyze_rss_structure",
            "news_analyze_css_selector"
        ]

class APIConfig:
    def __init__(self):
        self.port = int(os.getenv("API_PORT", "9001"))
        self.jwt_secret = os.getenv("JWT_SECRET")
        if not self.jwt_secret:
            raise ValueError("❌ Required environment variable JWT_SECRET is not set")

class GeminiConfig:
    def __init__(self):
        self.api_key = os.getenv("GEMINI_API_KEY")
        if not self.api_key:
            raise ValueError("❌ Required environment variable GEMINI_API_KEY is not set")
        self.model = os.getenv("GEMINI_MODEL", "gemini-1.5-flash")

class Config:
    def __init__(self):
        self.database = DatabaseConfig()
        self.kafka = KafkaConfig()
        self.api = APIConfig()
        self.gemini = GeminiConfig()

config = Config()
