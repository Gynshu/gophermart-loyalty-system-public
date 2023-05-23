package models

import "time"

type User struct {
	ID       string `json:"id" db:"id" `
	Login    string `json:"login" db:"login" `
	Password string `json:"password" db:"password" `
}
type Balance struct {
	ID        string  `json:"id" db:"id"`
	UserID    string  `json:"user_id" db:"user_id"`
	Amount    float64 `json:"amount" db:"amount"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}
type Order struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
type ResponseOrder struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
type Withdrawal struct {
	ID          string    `json:"id,omitempty" db:"id"`
	UserID      string    `json:"user_id,omitempty" db:"user_id"`
	OrderID     string    `json:"order" db:"order_id"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

const (
	userID               = "userID"
	OrderStatusNew       = "NEW"
	OrderStatusInvalid   = "INVALID"
	OrderStatusProcessed = "PROCESSED"
)
