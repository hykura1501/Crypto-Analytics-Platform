package model

import (
	"time"
)

// MarketPrice represents candlestick/KLINE data
// Using composite primary key (symbol, time) for TimescaleDB efficiency
type MarketPrice struct {
	Symbol   string    `gorm:"primaryKey;size:20" json:"symbol"`
	Time     time.Time `gorm:"primaryKey;index" json:"time"`
	Interval string    `gorm:"size:10;index" json:"interval"` // 1m, 5m, 1h, 1d, etc.
	Open     float64   `gorm:"type:decimal(20,8)" json:"open"`
	High     float64   `gorm:"type:decimal(20,8)" json:"high"`
	Low      float64   `gorm:"type:decimal(20,8)" json:"low"`
	Close    float64   `gorm:"type:decimal(20,8)" json:"close"`
	Volume   float64   `gorm:"type:decimal(20,8)" json:"volume"`
}

// TableName specifies the table name
func (MarketPrice) TableName() string {
	return "market_prices"
}

// BinanceKline represents Binance API kline response
type BinanceKline struct {
	OpenTime                 int64  `json:"open_time"`
	Open                     string `json:"open"`
	High                     string `json:"high"`
	Low                      string `json:"low"`
	Close                    string `json:"close"`
	Volume                   string `json:"volume"`
	CloseTime                int64  `json:"close_time"`
	QuoteAssetVolume         string `json:"quote_asset_volume"`
	NumberOfTrades           int    `json:"number_of_trades"`
	TakerBuyBaseAssetVolume  string `json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume string `json:"taker_buy_quote_asset_volume"`
	Ignore                   string `json:"ignore"`
}

// BinanceWSMessage represents Binance WebSocket kline message
type BinanceWSMessage struct {
	EventType string `json:"e"` // Event type
	EventTime int64  `json:"E"` // Event time
	Symbol    string `json:"s"` // Symbol
	Kline     struct {
		StartTime           int64  `json:"t"` // Kline start time
		EndTime             int64  `json:"T"` // Kline close time
		Symbol              string `json:"s"` // Symbol
		Interval            string `json:"i"` // Interval
		FirstTradeID        int64  `json:"f"` // First trade ID
		LastTradeID         int64  `json:"L"` // Last trade ID
		Open                string `json:"o"` // Open price
		Close               string `json:"c"` // Close price
		High                string `json:"h"` // High price
		Low                 string `json:"l"` // Low price
		Volume              string `json:"v"` // Base asset volume
		NumberOfTrades      int    `json:"n"` // Number of trades
		IsClosed            bool   `json:"x"` // Is this kline closed?
		QuoteVolume         string `json:"q"` // Quote asset volume
		TakerBuyBaseVolume  string `json:"V"` // Taker buy base asset volume
		TakerBuyQuoteVolume string `json:"Q"` // Taker buy quote asset volume
		Ignore              string `json:"B"` // Ignore
	} `json:"k"`
}
