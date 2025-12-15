// This file contains tests that are executed to verify your solution.
// It's read-only, so all modifications will be ignored.
package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"remove_sagas/service"
)

func TestComponent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.Run(ctx)
	}()

	waitForHttpServer(t)

	product1ID := uuid.New().String()
	product1Quantity := 2
	addProductToStock(t, product1ID, product1Quantity)

	product2ID := uuid.New().String()
	product2Quantity := 3
	addProductToStock(t, product2ID, product2Quantity)

	t.Run("order_shipped", func(t *testing.T) {
		orderID := uuid.New().String()
		placeOrder(t, orderID, map[string]int{
			product1ID: 1,
			product2ID: 1,
		})
		requireOrderShipped(t, orderID)
	})

	t.Run("out_of_order_product", func(t *testing.T) {
		orderID := uuid.New().String()
		placeOrder(t, orderID, map[string]int{
			product1ID: 1,
			product2ID: 3, // 1 should be missing
		})
		requireOrderCancelled(t, orderID)
	})

	t.Run("order_shipped_with_all_left_products", func(t *testing.T) {
		orderID := uuid.New().String()
		placeOrder(t, orderID, map[string]int{
			product1ID: 1,
			product2ID: 2,
		})
		requireOrderShipped(t, orderID)
	})
}

func requireOrderShipped(t *testing.T, orderID string) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			order := getOrder(t, orderID)
			if !assert.NotEmpty(t, order, "Order not found: %s", orderID) {
				return
			}

			if !assert.True(t, order.Shipped, "Order not shipped: %s", orderID) {
				return
			}

			if !assert.False(t, order.Cancelled, "Order should not be cancelled: %s", orderID) {
				return
			}
		},
		time.Second*5,
		time.Millisecond*500,
		"Order not shipped: %s",
		orderID,
	)
}

func requireOrderCancelled(t *testing.T, orderID string) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			order := getOrder(t, orderID)
			if !assert.NotEmpty(t, order, "Order not found: %s", orderID) {
				return
			}

			if !assert.False(t, order.Shipped, "Order should not be shipped: %s", orderID) {
				return
			}

			if !assert.True(t, order.Cancelled, "Order should be cancelled: %s", orderID) {
				return
			}
		},
		time.Second*5,
		time.Millisecond*500,
		"Order not cancelled: %s",
		orderID,
	)
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
		"HTTP server not ready",
	)
}

func addProductToStock(t *testing.T, productID string, quantity int) {
	t.Helper()

	reqBody := map[string]any{
		"product_id": productID,
		"quantity":   quantity,
	}

	u := "http://localhost:8080/products-stock"
	resp, err := http.Post(
		u,
		"application/json",
		bytes.NewReader(lo.Must(json.Marshal(reqBody))),
	)
	require.NoError(t, err, "Failed to call POST %s", u)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to call POST %s", u)
}

func placeOrder(t *testing.T, orderID string, products map[string]int) {
	t.Helper()

	reqBody := map[string]any{
		"order_id": orderID,
		"products": products,
	}

	u := "http://localhost:8080/orders"

	resp, err := http.Post(
		u,
		"application/json",
		bytes.NewReader(lo.Must(json.Marshal(reqBody))),
	)
	require.NoError(t, err, "Failed to call POST %s", u)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to call POST %s", u)
}

type OrderResponse struct {
	OrderID   string `json:"order_id"`
	Shipped   bool   `json:"shipped"`
	Cancelled bool   `json:"cancelled"`
}

func getOrder(t assert.TestingT, orderID string) OrderResponse {
	u := "http://localhost:8080/orders/" + orderID

	resp, err := http.Get(
		u,
	)
	if !assert.NoError(t, err, "Failed to get order: %s", orderID) {
		return OrderResponse{}
	}
	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusOK,
		resp.StatusCode,
		"Unexpected status code for GET %s, expected 200, got %d",
		u,
		resp.StatusCode,
	)

	var orderResponse OrderResponse

	err = json.NewDecoder(resp.Body).Decode(&orderResponse)
	if !assert.NoError(t, err, "Failed to decode order response for GET %s", u) {
		return OrderResponse{}
	}

	return orderResponse
}
