// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InMemoryVipBundleRepository struct {
	vipBundles     map[VipBundleID]VipBundle
	vipBundlesLock sync.RWMutex
}

func (i *InMemoryVipBundleRepository) Add(ctx context.Context, vipBundle VipBundle) error {
	i.vipBundlesLock.Lock()
	defer i.vipBundlesLock.Unlock()

	i.vipBundles[vipBundle.VipBundleID] = vipBundle
	return nil
}

func (i *InMemoryVipBundleRepository) Get(ctx context.Context, vipBundleID VipBundleID) (VipBundle, error) {
	i.vipBundlesLock.RLock()
	defer i.vipBundlesLock.RUnlock()

	if vipBundle, ok := i.vipBundles[vipBundleID]; ok {
		return vipBundle, nil
	}

	return VipBundle{}, fmt.Errorf("vip bundle with id %s not found", vipBundleID)
}

func (i *InMemoryVipBundleRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (VipBundle, error) {
	i.vipBundlesLock.RLock()
	defer i.vipBundlesLock.RUnlock()

	for _, vipBundle := range i.vipBundles {
		if vipBundle.BookingID == bookingID {
			return vipBundle, nil
		}
	}

	return VipBundle{}, fmt.Errorf("vip bundle with booking id %s not found", bookingID)
}

func (i *InMemoryVipBundleRepository) UpdateByID(
	ctx context.Context,
	vipBundleID VipBundleID,
	updateFn func(vipBundle VipBundle) (VipBundle, error),
) (VipBundle, error) {
	i.vipBundlesLock.Lock()
	defer i.vipBundlesLock.Unlock()

	vipBundle, ok := i.vipBundles[vipBundleID]
	if !ok {
		return VipBundle{}, fmt.Errorf("vip bundle with id %s not found", vipBundleID)
	}

	vipBundle, err := updateFn(vipBundle)
	if err != nil {
		return VipBundle{}, err
	}

	i.vipBundles[vipBundle.VipBundleID] = vipBundle

	return vipBundle, nil
}

func (i *InMemoryVipBundleRepository) UpdateByBookingID(
	ctx context.Context,
	bookingID uuid.UUID,
	updateFn func(vipBundle VipBundle) (VipBundle, error),
) (VipBundle, error) {
	i.vipBundlesLock.Lock()
	defer i.vipBundlesLock.Unlock()

	for _, vipBundle := range i.vipBundles {
		if vipBundle.BookingID == bookingID {
			vipBundle, err := updateFn(vipBundle)
			if err != nil {
				return VipBundle{}, err
			}

			i.vipBundles[vipBundle.VipBundleID] = vipBundle

			return vipBundle, nil
		}
	}

	return VipBundle{}, fmt.Errorf("vip bundle with booking id %s not found", bookingID)
}

type InMemoryCommandBus struct {
	Commands []any
}

func (i *InMemoryCommandBus) Send(ctx context.Context, command any) error {
	i.Commands = append(i.Commands, command)
	return nil
}

func (i *InMemoryCommandBus) PopCommands() []any {
	commands := i.Commands
	i.Commands = nil
	return commands
}

type InMemoryEventBus struct {
	Events []any
}

func (i *InMemoryEventBus) Publish(ctx context.Context, event any) error {
	i.Events = append(i.Events, event)
	return nil
}

func (i *InMemoryEventBus) PopEvents() []any {
	events := i.Events
	i.Events = nil
	return events
}

func TestVipBundleProcessManager_successful_flow(t *testing.T) {
	commandBus := &InMemoryCommandBus{}
	eventBus := &InMemoryEventBus{}
	repo := &InMemoryVipBundleRepository{
		vipBundles: make(map[VipBundleID]VipBundle),
	}

	pm := NewVipBundleProcessManager(commandBus, eventBus, repo)
	vb := newTestVipBundle(t)

	ctx := context.Background()

	err := repo.Add(ctx, *vb)
	require.NoError(t, err)

	err = pm.OnVipBundleInitialized(ctx, &VipBundleInitialized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vb.VipBundleID,
	})
	require.NoError(t, err)

	commands := commandBus.PopCommands()
	require.Lenf(t, commands, 1, "commands: %#v", commands)
	assert.Equal(
		t,
		BookShowTickets{
			BookingID:       vb.BookingID,
			CustomerEmail:   vb.CustomerEmail,
			NumberOfTickets: vb.NumberOfTickets,
			ShowId:          vb.ShowId,
		},
		commands[0],
	)

	err = pm.OnBookingMade(ctx, &BookingMade_v1{
		Header:          NewMessageHeader(),
		NumberOfTickets: vb.NumberOfTickets,
		BookingID:       vb.BookingID,
		CustomerEmail:   vb.CustomerEmail,
		ShowID:          vb.ShowId,
	})
	require.NoError(t, err)

	vbFromRepo, err := repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	require.NotNil(t, vbFromRepo.BookingMadeAt)

	commands = commandBus.PopCommands()
	require.Lenf(t, commands, 1, "commands: %#v", commands)
	assert.Empty(
		t,
		cmp.Diff(
			BookFlight{
				CustomerEmail: vb.CustomerEmail,
				FlightID:      vb.InboundFlightID,
				Passengers:    vb.Passengers,
				ReferenceID:   vb.VipBundleID.String(),
			},
			commands[0],
			cmpopts.IgnoreFields(BookFlight{}, "IdempotencyKey"),
		),
	)

	inboundFlightBooked := FlightBooked_v1{
		Header:      NewMessageHeader(),
		FlightID:    vb.InboundFlightID,
		TicketIDs:   []uuid.UUID{uuid.New(), uuid.New(), uuid.New()},
		ReferenceID: vb.VipBundleID.String(),
	}
	err = pm.OnFlightBooked(ctx, &inboundFlightBooked)
	require.NoError(t, err)

	vbFromRepo, err = repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	assert.True(t, vbFromRepo.IsFinalized)

	commands = commandBus.PopCommands()
	require.Lenf(t, commands, 0, "commands: %#v", commands)

	events := eventBus.PopEvents()
	require.Lenf(t, events, 1, "commands: %#v", commands)
	finalizedEvent := events[0].(VipBundleFinalized_v1)

	assert.Equal(
		t,
		vb.VipBundleID,
		finalizedEvent.VipBundleID,
	)
	assert.Equal(
		t,
		true,
		finalizedEvent.Success,
	)
}

func TestVipBundleProcessManager_show_booking_failed(t *testing.T) {
	commandBus := &InMemoryCommandBus{}
	eventBus := &InMemoryEventBus{}
	repo := &InMemoryVipBundleRepository{
		vipBundles: make(map[VipBundleID]VipBundle),
	}

	pm := NewVipBundleProcessManager(commandBus, eventBus, repo)
	vb := newTestVipBundle(t)

	ctx := context.Background()

	err := repo.Add(ctx, *vb)
	require.NoError(t, err)

	err = pm.OnVipBundleInitialized(ctx, &VipBundleInitialized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vb.VipBundleID,
	})
	require.NoError(t, err)

	commands := commandBus.PopCommands()
	require.Lenf(t, commands, 1, "commands: %#v", commands)
	assert.Equal(
		t,
		BookShowTickets{
			BookingID:       vb.BookingID,
			CustomerEmail:   vb.CustomerEmail,
			NumberOfTickets: vb.NumberOfTickets,
			ShowId:          vb.ShowId,
		},
		commands[0],
	)

	err = pm.OnBookingFailed(ctx, &BookingFailed_v1{
		Header:    NewMessageHeader(),
		BookingID: vb.BookingID,
	})
	require.NoError(t, err)

	vbFromRepo, err := repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	assert.True(t, vbFromRepo.IsFinalized)
	assert.True(t, vbFromRepo.Failed)

	commands = commandBus.PopCommands()
	require.Lenf(t, commands, 0, "commands: %#v", commands)

	events := eventBus.PopEvents()
	require.Lenf(t, events, 1, "commands: %#v", commands)
	finalizedEvent := events[0].(VipBundleFinalized_v1)

	assert.Equal(
		t,
		vb.VipBundleID,
		finalizedEvent.VipBundleID,
	)
	assert.Equal(
		t,
		false,
		finalizedEvent.Success,
	)
}

func TestVipBundleProcessManager_inbound_flight_failed(t *testing.T) {
	commandBus := &InMemoryCommandBus{}
	eventBus := &InMemoryEventBus{}
	repo := &InMemoryVipBundleRepository{
		vipBundles: make(map[VipBundleID]VipBundle),
	}

	pm := NewVipBundleProcessManager(commandBus, eventBus, repo)
	vb := newTestVipBundle(t)

	ctx := context.Background()

	err := repo.Add(ctx, *vb)
	require.NoError(t, err)

	err = pm.OnVipBundleInitialized(ctx, &VipBundleInitialized_v1{
		Header:      NewMessageHeader(),
		VipBundleID: vb.VipBundleID,
	})
	require.NoError(t, err)

	commands := commandBus.PopCommands()
	require.Lenf(t, commands, 1, "commands: %#v", commands)
	assert.Equal(
		t,
		BookShowTickets{
			BookingID:       vb.BookingID,
			CustomerEmail:   vb.CustomerEmail,
			NumberOfTickets: vb.NumberOfTickets,
			ShowId:          vb.ShowId,
		},
		commands[0],
	)

	err = pm.OnBookingMade(ctx, &BookingMade_v1{
		Header:          NewMessageHeader(),
		NumberOfTickets: vb.NumberOfTickets,
		BookingID:       vb.BookingID,
		CustomerEmail:   vb.CustomerEmail,
		ShowID:          vb.ShowId,
	})
	require.NoError(t, err)

	vbFromRepo, err := repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	require.NotNil(t, vbFromRepo.BookingMadeAt)

	commands = commandBus.PopCommands()
	require.Lenf(t, commands, 1, "commands: %#v", commands)
	assert.Empty(
		t,
		cmp.Diff(
			BookFlight{
				CustomerEmail: vb.CustomerEmail,
				FlightID:      vb.InboundFlightID,
				Passengers:    vb.Passengers,
				ReferenceID:   vb.VipBundleID.String(),
			},
			commands[0],
			cmpopts.IgnoreFields(BookFlight{}, "IdempotencyKey"),
		),
	)

	inboundFlightBookingFailed := FlightBookingFailed_v1{
		Header:      NewMessageHeader(),
		FlightID:    vb.InboundFlightID,
		ReferenceID: vb.VipBundleID.String(),
	}
	err = pm.OnFlightBookingFailed(ctx, &inboundFlightBookingFailed)
	require.Error(
		t,
		err,
		"it should fail, because TicketBookingConfirmed_v1 was not handled yet, "+
			"you should check if len(vipBundle.TicketIDs) == vipBundle.NumberOfTickets",
	)

	ticketIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, ticketID := range ticketIDs {
		ticketID := ticketID

		err = pm.OnTicketBookingConfirmed(ctx, &TicketBookingConfirmed_v1{
			Header:        NewMessageHeader(),
			TicketID:      ticketID.String(),
			CustomerEmail: vb.CustomerEmail,
			Price: Money{
				Amount:   "100",
				Currency: "EUR",
			},
			BookingID: vb.BookingID.String(),
		})
		require.NoError(t, err)
	}

	vbFromRepo, err = repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	require.Len(t, vbFromRepo.TicketIDs, 3, "vbFromRepo.TicketIDs: %v", vbFromRepo.TicketIDs)

	err = pm.OnFlightBookingFailed(ctx, &inboundFlightBookingFailed)
	require.NoError(t, err)

	vbFromRepo, err = repo.Get(ctx, vb.VipBundleID)
	require.NoError(t, err)
	assert.True(t, vbFromRepo.IsFinalized)
	assert.True(t, vbFromRepo.Failed)

	commands = commandBus.PopCommands()
	require.Lenf(t, commands, 3, "commands: %#v", commands)

	refundedTicketIDs := []string{}
	for _, cmd := range commands {
		refundedTicketIDs = append(refundedTicketIDs, cmd.(RefundTicket).TicketID)
	}
	expectedRefundedTicketIDs := []string{}
	for _, ticketID := range vbFromRepo.TicketIDs {
		expectedRefundedTicketIDs = append(expectedRefundedTicketIDs, ticketID.String())
	}
	assert.ElementsMatch(t, refundedTicketIDs, expectedRefundedTicketIDs)

	events := eventBus.PopEvents()
	require.Lenf(t, events, 1, "commands: %#v", commands)
	finalizedEvent := events[0].(VipBundleFinalized_v1)

	assert.Equal(
		t,
		vb.VipBundleID,
		finalizedEvent.VipBundleID,
	)
	assert.Equal(
		t,
		false,
		finalizedEvent.Success,
	)
}

func newTestVipBundle(t *testing.T) *VipBundle {
	vb, err := NewVipBundle(
		VipBundleID{uuid.New()},
		uuid.New(),
		"example@example.com",
		3,
		uuid.New(),
		[]string{
			"Mariusz Pudzianowski",
			"Janusz Tracz",
			"Robert Kubica",
		},
		uuid.New(),
	)
	require.NoError(t, err)

	return vb
}
