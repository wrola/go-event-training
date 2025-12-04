package models

type Money struct {
    Amount   string `json:"amount" db:"amount"`
    Currency string `json:"currency" db:"currency"`
}