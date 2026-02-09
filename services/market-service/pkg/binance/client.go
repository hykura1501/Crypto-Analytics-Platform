package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/crypto-platform/market-service/internal/model"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetKlines fetches historical candlestick data from Binance
// interval: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
func (c *Client) GetKlines(symbol, interval string, limit int, startTime, endTime int64) ([]*model.MarketPrice, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s", c.baseURL, symbol, interval)

	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch klines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var rawKlines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, fmt.Errorf("failed to decode klines: %w", err)
	}

	prices := make([]*model.MarketPrice, 0, len(rawKlines))
	for _, kline := range rawKlines {
		if len(kline) < 12 {
			continue
		}

		price, err := parseKline(symbol, interval, kline)
		if err != nil {
			continue
		}
		prices = append(prices, price)
	}

	return prices, nil
}

func parseKline(symbol, interval string, kline []interface{}) (*model.MarketPrice, error) {
	openTime := int64(kline[0].(float64))

	open, err := strconv.ParseFloat(kline[1].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse open price: %w", err)
	}
	high, err := strconv.ParseFloat(kline[2].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse high price: %w", err)
	}
	low, err := strconv.ParseFloat(kline[3].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse low price: %w", err)
	}
	closePrice, err := strconv.ParseFloat(kline[4].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse close price: %w", err)
	}
	volume, err := strconv.ParseFloat(kline[5].(string), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume: %w", err)
	}

	return &model.MarketPrice{
		Symbol:   symbol,
		Time:     time.Unix(0, openTime*int64(time.Millisecond)),
		Interval: interval,
		Open:     open,
		High:     high,
		Low:      low,
		Close:    closePrice,
		Volume:   volume,
	}, nil
}
