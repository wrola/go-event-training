package handler

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewEventBus(pub message.Publisher) *cqrs.EventBus {

	bus, err := cqrs.NewEventBusWithConfig(
		pub,
		cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
					GenerateName: cqrs.StructName,
			},
		},
	)

	if err != nil {
		panic(err)
	}


	return bus
}

func NewCommandBus(pub message.Publisher) *cqrs.CommandBus {
	bus, err := cqrs.NewCommandBusWithConfig(
		pub,
		cqrs.CommandBusConfig{
			GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
				return params.CommandName, nil
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
		},
	)

	if err != nil {
		panic(err)
	}

	return bus
}