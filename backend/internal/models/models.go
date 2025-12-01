package models

import "time"

type TradingPair struct {
	ID         int       `json:"id"`
	Symbol     string    `json:"symbol"`
	BaseAsset  string    `json:"base_asset"`
	QuoteAsset string    `json:"quote_asset"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PriceHistory struct {
	ID         int       `json:"id"`
	PairID     int       `json:"pair_id"`
	Price      float64   `json:"price"`
	Volume     float64   `json:"volume"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	OpenPrice  float64   `json:"open_price"`
	ClosePrice float64   `json:"close_price"`
	Timestamp  time.Time `json:"timestamp"`
	Interval   string    `json:"interval"`
	CreatedAt  time.Time `json:"created_at"`
}

type NewsSource struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	SelectorType    string     `json:"selector_type"`
	TitleSelector   string     `json:"title_selector"`
	ContentSelector string     `json:"content_selector"`
	DateSelector    string     `json:"date_selector"`
	LinkSelector    string     `json:"link_selector"`
	Status          string     `json:"status"`
	LastCrawledAt   *time.Time `json:"last_crawled_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type News struct {
	ID             int       `json:"id"`
	SourceID       int       `json:"source_id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	URL            string    `json:"url"`
	Author         string    `json:"author"`
	PublishedAt    time.Time `json:"published_at"`
	CrawledAt      time.Time `json:"crawled_at"`
	SentimentScore *float64  `json:"sentiment_score"`
	SentimentLabel *string   `json:"sentiment_label"`
	Keywords       []string  `json:"keywords"`
	CreatedAt      time.Time `json:"created_at"`
}

type AIAnalysis struct {
	ID              int        `json:"id"`
	PairID          int        `json:"pair_id"`
	AnalysisType    string     `json:"analysis_type"`
	Prediction      string     `json:"prediction"`
	ConfidenceScore float64    `json:"confidence_score"`
	Reasoning       string     `json:"reasoning"`
	TimeHorizon     string     `json:"time_horizon"`
	CreatedAt       time.Time  `json:"created_at"`
	ValidUntil      *time.Time `json:"valid_until"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Watchlist struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	PairID    int       `json:"pair_id"`
	CreatedAt time.Time `json:"created_at"`
}

type BinanceKline struct {
	OpenTime                 int64  `json:"openTime"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"closeTime"`
	QuoteAssetVolume         string `json:"quoteAssetVolume"`
	NumberOfTrades           int64  `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
	Ignore                   string `json:"ignore"`
}
