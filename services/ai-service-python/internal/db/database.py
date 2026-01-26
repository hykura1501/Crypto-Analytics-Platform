import logging
from urllib.parse import quote_plus

import psycopg2
from sqlalchemy import create_engine

from config import config


def _sqlalchemy_url() -> str:
    c = config.database
    user = quote_plus(c.user or "")
    pw = quote_plus(c.password or "")
    return f"postgresql+psycopg2://{user}:{pw}@{c.host}:{c.port}/{c.dbname}?sslmode={c.sslmode}"


class Database:
    def __init__(self):
        self.conn = None
        self.engine = None

    def init(self):
        try:
            self.conn = psycopg2.connect(config.database.dsn())
            self.engine = create_engine(_sqlalchemy_url(), pool_pre_ping=True)
            logging.info("Database connection established")
        except Exception as e:
            logging.error(f"Failed to connect to database: {e}")
            raise e

    def close(self):
        if self.engine:
            self.engine.dispose()
        if self.conn:
            self.conn.close()

db = Database()
