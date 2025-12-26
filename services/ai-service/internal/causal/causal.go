package causal

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"
)

type MarketPrice struct {
	Symbol string
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type CausalResult struct {
	NewsID      int
	NewsTitle   string
	Symbol      string
	NewsTime    time.Time
	PriceBefore float64
	PriceAfter  float64
	ChangePct   float64
	Direction   string
}

func AlignNewsWithPrice(
	db *sql.DB,
	newsID int,
	newsTitle string,
	newsTime time.Time,
	symbol string,
) (*CausalResult, error) {
	// Get price 1 hour before news
	timeBefore := newsTime.Add(-1 * time.Hour)

	var priceBefore MarketPrice
	queryBefore := `
		SELECT symbol, time, open, high, low, close, volume
		FROM market_prices
		WHERE symbol = $1 AND time <= $2 AND time >= $3
		ORDER BY time DESC
		LIMIT 1
	`

	err := db.QueryRow(queryBefore, symbol, newsTime, timeBefore).Scan(
		&priceBefore.Symbol,
		&priceBefore.Time,
		&priceBefore.Open,
		&priceBefore.High,
		&priceBefore.Low,
		&priceBefore.Close,
		&priceBefore.Volume,
	)

	if err == sql.ErrNoRows {
		log.Printf("No price data before news #%d for %s", newsID, symbol)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying price before: %w", err)
	}

	// Get price 1 hour after news
	timeAfter := newsTime.Add(1 * time.Hour)

	var priceAfter MarketPrice
	queryAfter := `
		SELECT symbol, time, open, high, low, close, volume
		FROM market_prices
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC
		LIMIT 1
	`

	err = db.QueryRow(queryAfter, symbol, newsTime, timeAfter).Scan(
		&priceAfter.Symbol,
		&priceAfter.Time,
		&priceAfter.Open,
		&priceAfter.High,
		&priceAfter.Low,
		&priceAfter.Close,
		&priceAfter.Volume,
	)

	if err == sql.ErrNoRows {
		log.Printf("No price data after news #%d for %s", newsID, symbol)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying price after: %w", err)
	}

	// Calculate price movement
	changePct := ((priceAfter.Close - priceBefore.Close) / priceBefore.Close) * 100

	direction := "stable"
	if changePct > 1.0 {
		direction = "up"
	} else if changePct < -1.0 {
		direction = "down"
	}

	result := &CausalResult{
		NewsID:      newsID,
		NewsTitle:   truncate(newsTitle, 60),
		Symbol:      symbol,
		NewsTime:    newsTime,
		PriceBefore: priceBefore.Close,
		PriceAfter:  priceAfter.Close,
		ChangePct:   round(changePct, 2),
		Direction:   direction,
	}

	log.Printf(
		"📊 Causal Analysis: News #%d '%s' at %s → %s moved %.2f%% in next hour (%s)",
		newsID,
		truncate(newsTitle, 40),
		newsTime.Format("15:04"),
		symbol,
		changePct,
		direction,
	)

	return result, nil
}

func ExtractSymbolsFromEntities(entities map[string]interface{}) []string {
	if entities == nil {
		return []string{}
	}

	tickers, ok := entities["tickers"].([]interface{})
	if !ok {
		return []string{}
	}

	cryptoSymbols := map[string]bool{
		"BTC": true, "ETH": true, "USDT": true, "BNB": true,
		"SOL": true, "XRP": true, "ADA": true,
	}

	result := []string{}
	for _, ticker := range tickers {
		if tickerStr, ok := ticker.(string); ok {
			if cryptoSymbols[tickerStr] {
				result = append(result, tickerStr)
			}
		}
	}

	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func round(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}
