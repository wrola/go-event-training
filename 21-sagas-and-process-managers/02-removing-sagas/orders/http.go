package orders

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Order struct {
	OrderID   uuid.UUID         `json:"order_id"`
	Products  map[uuid.UUID]int `json:"products"`
	Shipped   bool              `json:"shipped"`
	Cancelled bool              `json:"cancelled"`
}

type GetOrderResponse struct {
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	Shipped   bool      `json:"shipped" db:"shipped"`
	Cancelled bool      `json:"cancelled" db:"cancelled"`
}

func mountHttpHandlers(e *echo.Echo, db *sqlx.DB) {
	e.POST("/products-stock", func(c echo.Context) error {
		productStock := ProductStock{}
		if err := c.Bind(&productStock); err != nil {
			return err
		}

		if productStock.Quantity <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "quantity must be greater than 0")
		}
		if productStock.ProductID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "product_id must be provided")
		}

		err := updateProductStock(db, productStock)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusCreated)
	})

	e.POST("/orders", func(c echo.Context) error {
		order := Order{}
		if err := c.Bind(&order); err != nil {
			return err
		}

		err := addOrder(c.Request().Context(), db, order)
		if err != nil {
			return fmt.Errorf("failed to add order: %w", err)
		}

		return c.NoContent(http.StatusCreated)
	})

	e.GET("/orders/:order_id", func(c echo.Context) error {
		orderID, err := uuid.Parse(c.Param("order_id"))
		if err != nil {
			return err
		}

		order := GetOrderResponse{}

		err = db.Get(
			&order,
			"SELECT order_id, shipped, cancelled FROM orders WHERE order_id = $1",
			orderID,
		)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		return c.JSON(http.StatusOK, Order{
			OrderID:   order.OrderID,
			Shipped:   order.Shipped,
			Cancelled: order.Cancelled,
		})
	})
}
