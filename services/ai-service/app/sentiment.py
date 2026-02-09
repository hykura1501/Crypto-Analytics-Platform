"""
Sentiment Analysis using NLTK VADER
"""
import logging
from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer

logger = logging.getLogger(__name__)

class SentimentAnalyzer:
    """VADER-based sentiment analyzer for crypto news"""
    
    def __init__(self):
        self.analyzer = SentimentIntensityAnalyzer()
        logger.info("VADER Sentiment Analyzer initialized")
    
    def analyze(self, text: str) -> dict:
        """
        Analyze sentiment of text
        
        Returns:
            {
                "compound": -1.0 to 1.0,
                "pos": 0.0 to 1.0,
                "neg": 0.0 to 1.0,
                "neu": 0.0 to 1.0,
                "label": "Positive" | "Negative" | "Neutral"
            }
        """
        scores = self.analyzer.polarity_scores(text)
        
        # Determine label
        compound = scores['compound']
        if compound >= 0.05:
            label = "Positive"
        elif compound <= -0.05:
            label = "Negative"
        else:
            label = "Neutral"
        
        return {
            "compound": compound,
            "pos": scores['pos'],
            "neg": scores['neg'],
            "neu": scores['neu'],
            "label": label
        }
    
    def get_sentiment_score(self, text: str) -> float:
        """Get just the compound score (-1 to 1)"""
        return self.analyzer.polarity_scores(text)['compound']

# Singleton instance
sentiment_analyzer = SentimentAnalyzer()
