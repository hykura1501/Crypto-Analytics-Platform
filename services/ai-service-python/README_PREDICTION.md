# Crypto Price Prediction - Đơn giản

## Cài đặt

```bash
pip install -r requirements.txt
```

## Sử dụng

### 1. Train Model

```python
from internal.prediction import PredictionPipeline

pipeline = PredictionPipeline()
metrics = pipeline.train(symbol="BTCUSDT", horizon_hours=4, years=2)
print(metrics)
```

### 2. Predict

```python
result = pipeline.predict(symbol="BTCUSDT", horizon_hours=4)
print(result)
```

## API Endpoints

### Train

**Endpoint:** `POST /api/v1/ai/prediction/train`

**Request Body:**
```json
{
  "symbol": "BTCUSDT",
  "horizon_hours": 4,
  "years": 2
}
```

**Curl Command:**
```bash
curl -X POST http://localhost:9001/api/v1/ai/prediction/train \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "horizon_hours": 4,
    "years": 2
  }'
```

**Response:**
```json
{
  "message": "Training completed",
  "rmse": 125.50,
  "mae": 85.20,
  "directional_accuracy": 72.5
}
```

### Predict

**Endpoint:** `GET /api/v1/ai/prediction/predict/{symbol}`

**Query Parameters:**
- `horizon_hours` (optional, default: 4): Prediction horizon in hours (4, 8, or 24)

**Curl Command:**
```bash
# Predict với horizon mặc định (4h)
curl http://localhost:9001/api/v1/ai/prediction/predict/BTCUSDT

# Predict với horizon cụ thể
curl "http://localhost:9001/api/v1/ai/prediction/predict/BTCUSDT?horizon_hours=4"
curl "http://localhost:9001/api/v1/ai/prediction/predict/BTCUSDT?horizon_hours=8"
curl "http://localhost:9001/api/v1/ai/prediction/predict/BTCUSDT?horizon_hours=24"
```

**Response Example:**
```json
{
  "prediction_horizon": "4h",
  "predicted_price": "42150.50",
  "current_price": "42100.00",
  "predicted_change_pct": "0.12",
  "confidence_score": "0.8",
  "top_influential_features": ["rsi", "sma_24", "volatility_24h", "sentiment_mean", "news_count"],
  "feature_analysis": {
    "top_news_features": [
      {"feature": "sentiment_mean", "importance": 0.15}
    ],
    "top_technical_indicators": [
      {"feature": "rsi", "importance": 0.12},
      {"feature": "sma_24", "importance": 0.10}
    ],
    "shap_contributions": {
      "rsi": 0.25,
      "sentiment_mean": 0.18,
      "sma_24": 0.12
    }
  },
  "top_keywords": ["bitcoin", "etf", "regulation", "crypto", "market"],
  "keyword_sentiment": {
    "bitcoin": 0.15,
    "etf": 0.25
  },
  "explanation": {
    "primary_factors": {
      "most_important_feature": "rsi",
      "is_news_important": true,
      "news_vs_technical": "News sentiment"
    }
  }
}
```

### Health Check

```bash
curl http://localhost:9001/api/v1/ai/health
```

## Features

- Price features: Returns, Volatility, SMA, RSI, MACD, Lag features
- News features: Sentiment mean, News count (aligned hourly)
- Model: XGBoost
- Evaluation: RMSE, MAE, Directional Accuracy
- XAI: Feature importance + SHAP (if available)
