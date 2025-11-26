package main

import (

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)



func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {

	bus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
		},
	)

	if err != nil {
		return nil, err
	}


	return bus, nil
}
