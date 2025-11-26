package main

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type FollowRequestSent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EventsCounter interface {
	CountEvent() error
}

func NewFollowRequestSentHandler(counter EventsCounter) cqrs.EventHandler {

	e := EventCounter{
		counter
	}

	
}

