from confluent_kafka import Consumer, KafkaError
import logging
import json
from config import config

class KafkaConsumer:
    def __init__(self):
        self.consumer = Consumer({
            'bootstrap.servers': config.kafka.broker,
            'group.id': config.kafka.group_id,
            'auto.offset.reset': 'earliest'
        })
        self.consumer.subscribe(config.kafka.topics)

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
