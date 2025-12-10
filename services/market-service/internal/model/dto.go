package model

// HistoryRequest represents query parameters for historical data
type HistoryRequest struct {
	Symbol   string `form:"symbol" binding:"required"`
	Interval string `form:"interval" binding:"required"`
	Limit    int    `form:"limit"`
	From     int64  `form:"from"` // Unix timestamp in milliseconds
	To       int64  `form:"to"`   // Unix timestamp in milliseconds
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
