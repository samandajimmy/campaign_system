package models

import (
	"time"
)

type Validators struct {
	Channel         string `json:"channel" validate:"required"`
	Product         string `json:"product" validate:"required"`
	TransactionType string `json:"transactionType" validate:"required"`
	Unit            string `json:"unit" validate:"required"`
}

type Campaign struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description" validate:"required"`
	StartDate   string     `json:"startDate" validate:"required"`
	EndDate     string     `json:"endDate" validate:"required"`
	Status      string     `json:"status"`
	Validators  Validators `json:"validators"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
