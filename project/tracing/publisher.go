package tracing

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracedPublisher struct {
	wrapped message.Publisher
}

func (p *TracedPublisher) Publish(topic string, messages ...*message.Message) error {
	for i := range messages {
		otel.GetTextMapPropagator().Inject(
			messages[i].Context(),
			propagation.MapCarrier(messages[i].Metadata),
		)
	}
	return p.wrapped.Publish(topic, messages...)
}

func (p *TracedPublisher) Close() error {
	return p.wrapped.Close()
}

func NewTracePublisher(pub message.Publisher) message.Publisher {
	return &TracedPublisher{wrapped: pub}
}
