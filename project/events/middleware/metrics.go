package middleware

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/prometheus/client_golang/prometheus"

	"tickets/metric"
)

func MetricsMiddleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		topic := message.SubscribeTopicFromCtx(msg.Context())
		handler := message.HandlerNameFromCtx(msg.Context())

		labels := prometheus.Labels{
			"topic":   topic,
			"handler": handler,
		}

		start := time.Now()

		msgs, err := h(msg)

		duration := time.Since(start).Seconds()
		metric.MessagesProcessingDuration.With(labels).Observe(duration)

		if err != nil {
			metric.MessagesProcessingFailedCounter.With(labels).Inc()
		} else {
			metric.MessagesProcessedCounter.With(labels).Inc()
		}

		return msgs, err
	}
}
