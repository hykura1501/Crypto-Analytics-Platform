flowchart TB
    subgraph Context["Context"]
        Pipeline["Sentiment Pipeline"]
    end
    
    subgraph Strategies["Strategies"]
        FinBERT["FinBERT\n(English)"]
        PhoBERT["PhoBERT\n(Vietnamese)"]
    end
    
    Article["Article\n(language: en/vi)"]
    
    Article -->|detect language| Pipeline
    Pipeline -->|en| FinBERT
    Pipeline -->|vi| PhoBERT
    FinBERT --> Score[sentiment_score]
    PhoBERT --> Score
