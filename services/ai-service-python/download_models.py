#!/usr/bin/env python3
"""
Download HuggingFace models to local cache directory.
This script downloads models to ./models_cache/ directory.
"""
import logging
import os
from pathlib import Path
from transformers import AutoTokenizer, AutoModelForSequenceClassification

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

MODELS = {
    'finbert': {
        'model_name': 'yiyanghkust/finbert-tone',
        'max_length': 512,
    },
    'phobert': {
        'model_name': 'wonrax/phobert-base-vietnamese-sentiment',
        'max_length': 256,
    }
}

def download_model(model_name, model_type, cache_dir):
    """Download and cache a model from HuggingFace to local directory."""
    logging.info(f"📥 Downloading {model_type}: {model_name}...")
    logging.info(f"   Cache directory: {cache_dir}")
    
    try:
        # Download tokenizer
        logging.info(f"   Downloading tokenizer...")
        if model_type == 'phobert':
            tokenizer = AutoTokenizer.from_pretrained(
                model_name,
                use_fast=False,
                cache_dir=cache_dir
            )
        else:  # finbert
            tokenizer = AutoTokenizer.from_pretrained(
                model_name,
                cache_dir=cache_dir
            )
        
        # Download model
        logging.info(f"   Downloading model (this may take a while)...")
        if model_type == 'finbert':
            model = AutoModelForSequenceClassification.from_pretrained(
                model_name,
                num_labels=3,
                output_hidden_states=True,
                output_attentions=True,
                cache_dir=cache_dir
            )
        else:  # phobert
            model = AutoModelForSequenceClassification.from_pretrained(
                model_name,
                output_hidden_states=True,
                output_attentions=True,
                cache_dir=cache_dir
            )
        
        logging.info(f"✅ {model_type} downloaded and cached successfully")
        return True
    except Exception as e:
        logging.error(f"❌ Failed to download {model_type}: {e}")
        return False

def main():
    # Get script directory
    script_dir = Path(__file__).parent.absolute()
    cache_dir = script_dir / "models_cache"
    
    # Create cache directory if it doesn't exist
    cache_dir.mkdir(exist_ok=True)
    
    logging.info("🚀 Starting model download...")
    logging.info(f"📁 Cache directory: {cache_dir}")
    logging.info("=" * 60)
    
    success = True
    for model_type, config in MODELS.items():
        if not download_model(config['model_name'], model_type, str(cache_dir)):
            success = False
        logging.info("")
    
    logging.info("=" * 60)
    if success:
        logging.info("✅ All models downloaded successfully!")
        logging.info(f"💡 Models cached at: {cache_dir}")
        logging.info("💡 To use these models, set HF_HOME environment variable or mount this directory")
    else:
        logging.error("❌ Some models failed to download")
        exit(1)

if __name__ == "__main__":
    main()
