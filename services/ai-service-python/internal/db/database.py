import psycopg2
import logging
from config import config

class Database:
    def __init__(self):
        self.conn = None

    def init(self):
        try:
            self.conn = psycopg2.connect(config.database.dsn())
            logging.info("Database connection established")
        except Exception as e:
            logging.error(f"Failed to connect to database: {e}")
            raise e

    def close(self):
        if self.conn:
            self.conn.close()

db = Database()
