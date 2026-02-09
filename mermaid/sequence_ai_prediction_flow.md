---
title: "AI Prediction with SHAP"
description: "Dự đoán giá với XAI"
---

sequenceDiagram
    participant Client as VIP User
    participant AI as AI Service
    participant DB as Database
    participant Model as XGBoost
    participant SHAP as SHAP

    Client->>AI: GET /ai/predict<br/>symbol=BTCUSDT, horizon=4h
    
    Note over AI,DB: Load Data
    AI->>DB: SELECT prices (2 years)
    AI->>DB: SELECT news (730 days)
    DB-->>AI: Historical data
    
    Note over AI,Model: Feature Engineering
    AI->>AI: Technical indicators<br/>(RSI, MACD, SMA, volatility)
    AI->>AI: News features<br/>(sentiment, lag, rolling)
    AI->>AI: Combine 30 features
    
    Note over AI,SHAP: Prediction + Explanation
    AI->>Model: predict(features)
    Model-->>AI: predicted_price
    
    AI->>SHAP: shap_values(features)
    SHAP-->>AI: Feature contributions
    AI->>AI: Top 10 features
    AI->>AI: Calculate confidence
    
    Note over AI,Client: Return
    AI-->>Client: {price, change%, confidence,<br/>top_features, top_news}
