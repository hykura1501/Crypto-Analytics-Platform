package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type BinanceService struct {
	BaseURL string
	Client  *http.Client
}

func NewBinanceService() *BinanceService {
	return &BinanceService{
		BaseURL: "https://api.binance.com/api/v3",
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type PriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type KlineResponse []interface{}

// GetCurrentPrice gets the current price for a trading pair
func (s *BinanceService) GetCurrentPrice(symbol string) (*PriceResponse, error) {
	url := fmt.Sprintf("%s/ticker/price?symbol=%s", s.BaseURL, symbol)

	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var price PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, err
	}

	return &price, nil
}

// GetKlines gets candlestick data for a trading pair
func (s *BinanceService) GetKlines(symbol, interval string, limit int) ([]interface{}, error) {
	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s&limit=%d", s.BaseURL, symbol, interval, limit)

	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance API error: %s", string(body))
	}

	var klines []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return nil, err
	}

	return klines, nil
}
