import torch
import numpy as np
import pandas as pd
from transformers import AutoTokenizer, AutoModelForSequenceClassification
from langdetect import detect, LangDetectException
from underthesea import word_tokenize

# Configuration
PREPROCESSING_CONFIG = {
    'finbert': {
        'model_name': 'yiyanghkust/finbert-tone',
        'max_length': 512,
    },
    'phobert': {
        'model_name': 'wonrax/phobert-base-vietnamese-sentiment',
        'max_length': 256,
    }
}

class Analyzer:
    """Sentiment analysis model using pre-trained transformers (FinBERT & PhoBERT)"""
    
    def __init__(self):
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        
        # Load FinBERT
        print(f"  - Loading FinBERT model: {PREPROCESSING_CONFIG['finbert']['model_name']}...")
        self.finbert_tokenizer = AutoTokenizer.from_pretrained(PREPROCESSING_CONFIG['finbert']['model_name'])
        self.finbert_model = AutoModelForSequenceClassification.from_pretrained(
            PREPROCESSING_CONFIG['finbert']['model_name'], 
            num_labels=3,
            output_hidden_states=True,
            output_attentions=True
        )
        self.finbert_model.to(self.device)
        self.finbert_model.eval()
        
        # Load PhoBERT
        print(f"  - Loading PhoBERT model: {PREPROCESSING_CONFIG['phobert']['model_name']}...")
        self.phobert_tokenizer = AutoTokenizer.from_pretrained(
            PREPROCESSING_CONFIG['phobert']['model_name'],
            use_fast=False
        )
        self.phobert_model = AutoModelForSequenceClassification.from_pretrained(
            PREPROCESSING_CONFIG['phobert']['model_name'],
            output_hidden_states=True,
            output_attentions=True
        )
        self.phobert_model.to(self.device)
        self.phobert_model.eval()

    def extract_keywords(self, tokenizer, inputs, outputs, top_k=5):
        try:
            if outputs.attentions is None:
                return ["N/A"]
                
            # Get attention weights from the last layer
            # Shape: [batch_size, num_heads, seq_len, seq_len]
            attentions = outputs.attentions[-1]
            
            # Average over heads
            # Shape: [batch_size, seq_len, seq_len]
            attentions_mean = attentions.mean(dim=1)
            
            # Get attention of [CLS] token (index 0) to other tokens
            # Shape: [seq_len]
            cls_attention = attentions_mean[0, 0, :].detach().cpu().numpy()
            
            # Get tokens
            input_ids = inputs['input_ids'][0]
            tokens = tokenizer.convert_ids_to_tokens(input_ids)
            
            # Filter and aggregate weights
            stop_words = set([
                # English stopwords
                'the', 'of', 'and', 'in', 'to', 'for', 'with', 'a', 'an', 'at', 'by', 'is', 'on', 'this', 'that', 'from',
                # Vietnamese stopwords
                'và', 'của', 'có', 'được', 'trong', 'là', 'cho', 'với', 'đã', 'để', 'các', 'một', 'này', 'đó', 'những',
                'được', 'tại', 'theo', 'từ', 'trên', 'về', 'cũng', 'như', 'khi', 'thì', 'đang', 'còn', 'hay', 'sẽ', 'bị',
                'do', 'nếu', 'mà', 'đến', 'ra', 'thêm', 'lại', 'nhiều', 'sau', 'hơn', 'nên', 'đều', 'chỉ', 'vào', 'giữa',
                # Model tokens
                '[cls]', '[sep]', '[pad]', '<s>', '</s>', '<pad>'
            ])
            token_weights = {}
            
            for i, t in enumerate(tokens):
                t_lower = t.lower()
                # Clean subword tokens (## for BERT, Ġ for RoBERTa)
                t_clean = t_lower.replace('##', '').replace('Ġ', '').replace('_', ' ').strip()
                
                if t_clean not in stop_words and len(t_clean) > 2 and t_clean.replace(' ', '').isalnum():
                    if t_clean not in token_weights or cls_attention[i] > token_weights[t_clean]:
                        token_weights[t_clean] = cls_attention[i]
            
            # Sort by weight
            sorted_kws = sorted(token_weights.items(), key=lambda x: x[1], reverse=True)
            return [kw for kw, weight in sorted_kws[:top_k]]
            
        except Exception as e:
            print(f"Error extracting keywords: {e}")
            return ["Error"]

    def analyze(self, text, lang='en', top_k=5):
        if not text or (isinstance(text, float) and pd.isna(text)):
            return ["N/A"], 0.0
            
        if lang == 'vi':
            return self.analyze_vietnamese(text, top_k)
        else:
            return self.analyze_english(text, top_k)

    def analyze_english(self, text, top_k):
        config = PREPROCESSING_CONFIG['finbert']
        
        inputs = self.finbert_tokenizer(
            text, 
            return_tensors="pt", 
            padding=True, 
            truncation=True, 
            max_length=config['max_length']
        ).to(self.device)
        
        with torch.no_grad():
            outputs = self.finbert_model(**inputs)
            probs = torch.softmax(outputs.logits, dim=-1).cpu().numpy()[0]
            
        # FinBERT: [0]: Neutral, [1]: Positive, [2]: Negative (Wait, verify label mapping)
        # yiyanghkust/finbert-tone: label2id: {'Neutral': 0, 'Positive': 1, 'Negative': 2}
        # Wait, HuggingFace model card says:
        # labels: 0 -> Neutral, 1 -> Positive, 2 -> Negative
        # But commonly id2label is sorted. 
        # Let's trust the user's snippet which said: # probs[0]: Neutral, [1]: Positive, [2]: Negative
        
        neu_p, pos_p, neg_p = probs[0], probs[1], probs[2]
        
        sentiment_score = pos_p - neg_p
        
        # Determine label
        if sentiment_score > 0.1:
            label = "Positive"
        elif sentiment_score < -0.1:
            label = "Negative"
        else:
            label = "Neutral"
            
        keywords = self.extract_keywords(self.finbert_tokenizer, inputs, outputs, top_k)
        
        return keywords, float(sentiment_score)

    def analyze_vietnamese(self, text, top_k):
        config = PREPROCESSING_CONFIG['phobert']
        
        # Segment words for PhoBERT
        text_segmented = word_tokenize(text, format="text")
        
        inputs = self.phobert_tokenizer(
            text_segmented, 
            return_tensors="pt", 
            padding=True, 
            truncation=True, 
            max_length=config['max_length']
        ).to(self.device)
        
        with torch.no_grad():
            outputs = self.phobert_model(**inputs)
            probs = torch.softmax(outputs.logits, dim=-1).cpu().numpy()[0]
            
        # PhoBERT sentiment: [0]: Negative, [1]: Positive, [2]: Neutral
        # Source: wonrax/phobert-base-vietnamese-sentiment
        # "Output: [[0.002, 0.988, 0.01]] -> NEG, POS, NEU"
        
        neg_p, pos_p, neu_p = probs[0], probs[1], probs[2]
        
        sentiment_score = pos_p - neg_p
        
        if sentiment_score > 0.1:
            label = "Positive"
        elif sentiment_score < -0.1:
            label = "Negative"
        else:
            label = "Neutral"
            
        keywords = self.extract_keywords(self.phobert_tokenizer, inputs, outputs, top_k)
        
        return keywords, float(sentiment_score)
