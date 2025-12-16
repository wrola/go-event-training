package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"

	"tickets/entities"
	"tickets/entities/events"
	"tickets/events/handler"
)

type VipBundleRepository struct {
	db     *sqlx.DB
	logger watermill.LoggerAdapter
}

func NewVipBundleRepository(db *sqlx.DB, logger watermill.LoggerAdapter) VipBundleRepository {
	if db == nil {
		panic("db is nil")
	}
	if logger == nil {
		panic("logger is nil")
	}

	return VipBundleRepository{
		db:     db,
		logger: logger,
	}
}

func (r VipBundleRepository) Add(ctx context.Context, vipBundle entities.VipBundle) error {
	return updateInTx(
		ctx,
		r.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			payload, err := json.Marshal(vipBundle)
			if err != nil {
				return fmt.Errorf("could not marshal vip bundle: %w", err)
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO vip_bundles (vip_bundle_id, booking_id, payload)
				VALUES ($1, $2, $3)
			`, vipBundle.VipBundleID.UUID, vipBundle.BookingID, payload)
			if err != nil {
				return fmt.Errorf("could not insert vip bundle: %w", err)
			}

			outboxPublisher, err := NewOutboxPublisher(tx, r.logger)
			if err != nil {
				return fmt.Errorf("could not create outbox publisher: %w", err)
			}

			eventBus := handler.NewEventBus(outboxPublisher)

			event := events.NewVipBundleInitialized(vipBundle.VipBundleID)
			err = eventBus.Publish(ctx, event)
			if err != nil {
				return fmt.Errorf("could not publish VipBundleInitialized event: %w", err)
			}

			return nil
		},
	)
}

func (r VipBundleRepository) Get(ctx context.Context, vipBundleID entities.VipBundleID) (entities.VipBundle, error) {
	var payload []byte
	err := r.db.GetContext(ctx, &payload, `
		SELECT payload FROM vip_bundles WHERE vip_bundle_id = $1
	`, vipBundleID.UUID)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not get vip bundle: %w", err)
	}

	var vipBundle entities.VipBundle
	err = json.Unmarshal(payload, &vipBundle)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not unmarshal vip bundle: %w", err)
	}

	return vipBundle, nil
}

func (r VipBundleRepository) GetByBookingID(ctx context.Context, bookingID string) (entities.VipBundle, error) {
	var payload []byte
	err := r.db.GetContext(ctx, &payload, `
		SELECT payload FROM vip_bundles WHERE booking_id = $1
	`, bookingID)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not get vip bundle by booking_id: %w", err)
	}

	var vipBundle entities.VipBundle
	err = json.Unmarshal(payload, &vipBundle)
	if err != nil {
		return entities.VipBundle{}, fmt.Errorf("could not unmarshal vip bundle: %w", err)
	}

	return vipBundle, nil
}

func (r VipBundleRepository) UpdateByID(
	ctx context.Context,
	vipBundleID entities.VipBundleID,
	updateFn func(vipBundle entities.VipBundle) (entities.VipBundle, error),
) (entities.VipBundle, error) {
	var updated entities.VipBundle

	err := updateInTx(
		ctx,
		r.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			var payload []byte
			err := tx.GetContext(ctx, &payload, `
				SELECT payload FROM vip_bundles WHERE vip_bundle_id = $1 FOR UPDATE
			`, vipBundleID.UUID)
			if err != nil {
				return fmt.Errorf("could not get vip bundle for update: %w", err)
			}

			var vipBundle entities.VipBundle
			err = json.Unmarshal(payload, &vipBundle)
			if err != nil {
				return fmt.Errorf("could not unmarshal vip bundle: %w", err)
			}

			updated, err = updateFn(vipBundle)
			if err != nil {
				return fmt.Errorf("update function failed: %w", err)
			}

			updatedPayload, err := json.Marshal(updated)
			if err != nil {
				return fmt.Errorf("could not marshal updated vip bundle: %w", err)
			}

			_, err = tx.ExecContext(ctx, `
				UPDATE vip_bundles SET payload = $1 WHERE vip_bundle_id = $2
			`, updatedPayload, vipBundleID.UUID)
			if err != nil {
				return fmt.Errorf("could not update vip bundle: %w", err)
			}

			return nil
		},
	)

	return updated, err
}

func (r VipBundleRepository) UpdateByBookingID(
	ctx context.Context,
	bookingID string,
	updateFn func(vipBundle entities.VipBundle) (entities.VipBundle, error),
) (entities.VipBundle, error) {
	var updated entities.VipBundle

	err := updateInTx(
		ctx,
		r.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			var payload []byte
			err := tx.GetContext(ctx, &payload, `
				SELECT payload FROM vip_bundles WHERE booking_id = $1 FOR UPDATE
			`, bookingID)
			if err != nil {
				return fmt.Errorf("could not get vip bundle for update: %w", err)
			}

			var vipBundle entities.VipBundle
			err = json.Unmarshal(payload, &vipBundle)
			if err != nil {
				return fmt.Errorf("could not unmarshal vip bundle: %w", err)
			}

			updated, err = updateFn(vipBundle)
			if err != nil {
				return fmt.Errorf("update function failed: %w", err)
			}

			updatedPayload, err := json.Marshal(updated)
			if err != nil {
				return fmt.Errorf("could not marshal updated vip bundle: %w", err)
			}

			_, err = tx.ExecContext(ctx, `
				UPDATE vip_bundles SET payload = $1 WHERE booking_id = $2
			`, updatedPayload, bookingID)
			if err != nil {
				return fmt.Errorf("could not update vip bundle: %w", err)
			}

			return nil
		},
	)

	return updated, err
}
