package main

import (
	"encoding/json"
	"time"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)


type EventName string

const (
	ProductOutOfStockEvent EventName = "ProductOutOfStock"
	ProductBackInStockEvent EventName = "ProductBackInStock"
)

type EventHeader struct {
	ID string `json:"id"`
	EventName EventName `json:"event_name"`
	OccurredAt string `json:"occurred_at"`
}

type ProductOutOfStock struct {
	Header   EventHeader `json:"header`
	ProductID string `json:"product_id"`
}

type ProductBackInStock struct {
	Header   EventHeader `json:"header"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Publisher struct {
	pub message.Publisher
}

func NewPublisher(pub message.Publisher) Publisher {
	return Publisher{
		pub: pub,
	}
}

func NewMesasgeHeader(eventName EventName) EventHeader {
	return EventHeader{
		ID: watermill.NewUUID(),
		EventName: eventName,
		OccurredAt: time.Now().Format(time.RFC3339),
	}
}

func (p Publisher) PublishProductOutOfStock(productID string) error {
	event := ProductOutOfStock{
		ProductID: productID,
		Header:    NewMesasgeHeader(ProductOutOfStockEvent),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	return p.pub.Publish("product-updates", msg)
}

func (p Publisher) PublishProductBackInStock(productID string, quantity int) error {
	event := ProductBackInStock{
		Header:    NewMesasgeHeader(ProductBackInStockEvent),
		ProductID: productID,
		Quantity:  quantity,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	return p.pub.Publish("product-updates", msg)
}
