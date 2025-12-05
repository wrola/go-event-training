package entities

import "time"

type IssueReceiptResponse struct {
	ReceiptNumber string
	IssuedAt      time.Time
}
