package main

import (
	"time"
)

// Invoice represents invoices sent to customers
type invoice struct {
	ID          int       `json:"id"`
	CustomerID  int       `json:"customerID"`
	Description string    `json:"description,omitempty"`
	DueDate     time.Time `json:"dueDate,omitempty"`
	Amount      float64   `json:"amount"`
}

// Invoices represents a list of invoices
type Invoices []invoice

// NotFoundError represents an item not found error
type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}
