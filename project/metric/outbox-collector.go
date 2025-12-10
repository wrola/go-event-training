package metric

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
)

// OutboxCollector is a custom Prometheus collector that queries the database
// to count pending messages in the outbox table on-demand when scraped.
type OutboxCollector struct {
	db          *sqlx.DB
	desc        *prometheus.Desc
	outboxTopic string
}

// NewOutboxCollector creates a new collector instance that will query
// the database for outbox message count when Prometheus scrapes metrics.
func NewOutboxCollector(db *sqlx.DB, outboxTopic string) *OutboxCollector {
	return &OutboxCollector{
		db:          db,
		outboxTopic: outboxTopic,
		desc: prometheus.NewDesc(
			"outbox_pending_messages",
			"The number of pending messages in the outbox waiting to be forwarded",
			[]string{"topic"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector interface.
// It sends the metric descriptor to the provided channel.
func (c *OutboxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

// Collect implements prometheus.Collector interface.
// It queries the database for the current outbox size and sends the metric value.
func (c *OutboxCollector) Collect(ch chan<- prometheus.Metric) {
	count, err := c.getOutboxSize()
	if err != nil {
		slog.Error("Failed to query outbox size",
			"error", err,
			"topic", c.outboxTopic,
		)
		return // Don't expose metric on error
	}

	ch <- prometheus.MustNewConstMetric(
		c.desc,
		prometheus.GaugeValue,
		count,
		c.outboxTopic,
	)
}

// getOutboxSize queries the database for the count of pending messages
// in the watermill_messages table for the configured outbox topic.
func (c *OutboxCollector) getOutboxSize() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM watermill_messages WHERE topic = $1`
	err := c.db.GetContext(ctx, &count, query, c.outboxTopic)
	if err != nil {
		return 0, err
	}

	return float64(count), nil
}
