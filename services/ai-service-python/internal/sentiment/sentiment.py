import math
import re

class Result:
    def __init__(self, compound, pos, neg, neu, label):
        self.compound = compound
        self.pos = pos
        self.neg = neg
        self.neu = neu
        self.label = label

class Analyzer:
    def __init__(self):
        self.positive_words = self.load_positive_words()
        self.negative_words = self.load_negative_words()

    def load_positive_words(self):
        words = [
            "good", "great", "excellent", "amazing", "wonderful", "fantastic",
            "positive", "bullish", "rise", "surge", "gain", "profit", "success",
            "up", "high", "strong", "growth", "increase", "boost", "rally",
            "breakthrough", "innovation", "adoption", "partnership", "launch",
            "approval", "support", "investment", "funding", "milestone",
            "tăng", "tốt", "tích cực", "tăng trưởng", "phát triển", "thành công",
        ]
        return set(words)

    def load_negative_words(self):
        words = [
            "bad", "terrible", "awful", "horrible", "negative", "bearish",
            "fall", "drop", "crash", "loss", "decline", "decrease", "fail",
            "down", "low", "weak", "risk", "concern", "worry", "fear",
            "hack", "attack", "scam", "fraud", "ban", "regulation", "ban",
            "lawsuit", "fine", "penalty", "rejection", "delay", "problem",
            "giảm", "xấu", "tiêu cực", "sụt giảm", "thất bại", "rủi ro",
        ]
        return set(words)

    def tokenize(self, text):
        # Simple tokenization: lowercase and keep only letters/numbers
        text = text.lower()
        tokens = re.findall(r'[a-z0-9]+', text)
        # Handle unicode characters for Vietnamese
        # In Python re, [a-z0-9] matches ASCII. For unicode support we might need more complex regex
        # But the Go code used unicode.IsLetter which supports unicode.
        # Let's try a better regex or just split by non-alphanumeric
        tokens = re.findall(r'\w+', text, re.UNICODE)
        return tokens

    def analyze(self, text):
        words = self.tokenize(text)
        
        pos_score = 0.0
        neg_score = 0.0
        neutral_count = 0
        
        for word in words:
            if word in self.positive_words:
                pos_score += 1
            elif word in self.negative_words:
                neg_score += 1
            else:
                neutral_count += 1
        
        total = float(len(words))
        if total == 0:
            return Result(0.0, 0.0, 0.0, 1.0, "Neutral")
        
        pos = pos_score / total
        neg = neg_score / total
        neu = float(neutral_count) / total
        
        # Calculate compound score (normalized difference)
        compound = (pos_score - neg_score) / max(total, 1.0)
        compound = max(-1.0, min(1.0, compound))
        
        label = "Neutral"
        if compound >= 0.05:
            label = "Positive"
        elif compound <= -0.05:
            label = "Negative"
        
        return Result(compound, pos, neg, neu, label)
