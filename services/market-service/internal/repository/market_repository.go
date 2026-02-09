package repository

import (
	"time"

	"github.com/crypto-platform/market-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MarketRepository interface {
	Create(price *model.MarketPrice) error
	BulkCreate(prices []*model.MarketPrice) error
	FindBySymbolAndTimeRange(symbol, interval string, from, to time.Time) ([]*model.MarketPrice, error)
	FindLatest(symbol, interval string, limit int) ([]*model.MarketPrice, error)
}

type marketRepository struct {
	db *gorm.DB
}

func NewMarketRepository(db *gorm.DB) MarketRepository {
	return &marketRepository{db: db}
}

func (r *marketRepository) Create(price *model.MarketPrice) error {
	// Use upsert to handle duplicate klines (Binance sends updates for same minute)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "time"}, {Name: "interval"}},
		DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume"}),
	}).Create(price).Error
}

func (r *marketRepository) BulkCreate(prices []*model.MarketPrice) error {
	if len(prices) == 0 {
		return nil
	}
	// Use upsert for bulk create as well
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "time"}, {Name: "interval"}},
		DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume"}),
	}).Create(prices).Error
}

func (r *marketRepository) FindBySymbolAndTimeRange(
	symbol, interval string,
	from, to time.Time,
) ([]*model.MarketPrice, error) {
	var prices []*model.MarketPrice
	err := r.db.Where("symbol = ? AND interval = ? AND time >= ? AND time <= ?",
		symbol, interval, from, to).
		Order("time ASC").
		Find(&prices).Error
	return prices, err
}

func (r *marketRepository) FindLatest(symbol, interval string, limit int) ([]*model.MarketPrice, error) {
	if limit <= 0 {
		limit = 100
	}
	var prices []*model.MarketPrice
	err := r.db.Where("symbol = ? AND interval = ?", symbol, interval).
		Order("time DESC").
		Limit(limit).
		Find(&prices).Error

	// Reverse to get chronological order
	for i, j := 0, len(prices)-1; i < j; i, j = i+1, j-1 {
		prices[i], prices[j] = prices[j], prices[i]
	}

	return prices, err
}
