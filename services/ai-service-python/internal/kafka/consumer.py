from confluent_kafka import Consumer, KafkaError
import logging
import json
from config import config

class KafkaConsumer:
    def __init__(self):
        logging.info(f"Creating Kafka consumer for broker: {config.kafka.broker}")
        logging.info(f"Consumer group: {config.kafka.group_id}")
        logging.info(f"Subscribing to topics: {', '.join(config.kafka.topics)}")
        
        self.consumer = Consumer({
            'bootstrap.servers': config.kafka.broker,
            'group.id': config.kafka.group_id,
            'auto.offset.reset': 'earliest',
            'session.timeout.ms': 10000,  # 10 seconds
            'metadata.max.age.ms': 5000,  # 5 seconds - faster metadata refresh
            'socket.timeout.ms': 10000,   # 10 seconds
            'api.version.request.timeout.ms': 10000,  # 10 seconds
            'enable.auto.commit': True
        })
        
        logging.info("Subscribing to Kafka topics...")
        try:
            self.consumer.subscribe(config.kafka.topics)
            logging.info("✅ Kafka consumer subscribed successfully")
        except Exception as e:
            logging.error(f"❌ Failed to subscribe to Kafka topics: {e}")
            raise

    def read_message(self):
        msg = self.consumer.poll(1.0)

        if msg is None:
            return None, None
        if msg.error():
            if msg.error().code() == KafkaError._PARTITION_EOF:
                return None, None
            else:
                logging.error(f"Kafka error: {msg.error()}")
                return None, msg.error()

        return msg, None

    def close(self):
        self.consumer.close()
