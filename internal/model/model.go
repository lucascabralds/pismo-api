package model

import "time"

type Account struct {
	AccountID      int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

type OperationType struct {
	OperationTypeID int64  `json:"operation_type_id"`
	Description     string `json:"description"`
}

type Transaction struct {
	TransactionID   int64     `json:"transaction_id"`
	AccountID       int64     `json:"account_id"`
	OperationTypeID int64     `json:"operation_type_id"`
	Amount          float64   `json:"amount"`
	EventDate       time.Time `json:"event_date"`
}

// Request DTOs
type CreateAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

type CreateTransactionRequest struct {
	AccountID       int64   `json:"account_id"`
	OperationTypeID int64   `json:"operation_type_id"`
	Amount          float64 `json:"amount"`
}

// Error response
type ErrorResponse struct {
	Error string `json:"error"`
}
