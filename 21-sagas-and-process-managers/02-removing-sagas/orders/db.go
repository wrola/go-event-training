package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"remove_sagas/common"
)

func initializeDatabaseSchema(db *sqlx.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			order_id UUID PRIMARY KEY,
			shipped BOOLEAN NOT NULL,
			cancelled BOOLEAN NOT NULL
		);

		CREATE TABLE IF NOT EXISTS order_products (
			order_id UUID NOT NULL,
			product_id UUID NOT NULL,
			quantity INT NOT NULL,

		    PRIMARY KEY (order_id, product_id)
		);

		CREATE TABLE IF NOT EXISTS stock (
			product_id UUID PRIMARY KEY,
			quantity INT NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}
}

func addOrder(ctx context.Context, db *sqlx.DB, order Order) error {
	return common.UpdateInTx(
		ctx,
		db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			outOfStock := false
			err := removeProductsFromStock(order.Products, tx)

			var productsOutOfStockError ProductsOutOfStockError
			if errors.As(err, &productsOutOfStockError) {
				outOfStock = true // Not enough stock
			} else if err != nil {
				return fmt.Errorf("failed to remove products from stock: %w", err)
			}

			shipped := !outOfStock
			cancelled := outOfStock

			_, err = tx.Exec(
				"INSERT INTO orders (order_id, shipped, cancelled) VALUES ($1, $2, $3)",
				order.OrderID,
				shipped,
				cancelled,
			)
			if err != nil {
				return fmt.Errorf("failed to insert order: %w", err)
			}

			for product, quantity := range order.Products {
				_, err := tx.Exec(
					"INSERT INTO order_products (order_id, product_id, quantity) VALUES ($1, $2, $3)",
					order.OrderID,
					product,
					quantity,
				)
				if err != nil {
					return fmt.Errorf("failed to insert order_products: %w", err)
				}
			}

			return nil
		},
	)
}

type ProductStock struct {
	ProductID string `db:"product_id" json:"product_id"`
	Quantity  int    `db:"quantity" json:"quantity"`
}

func updateProductStock(db *sqlx.DB, productStock ProductStock) error {
	_, err := db.Exec(`
		INSERT INTO stock (product_id, quantity)
		VALUES ($1, $2)
		ON CONFLICT (product_id) DO UPDATE SET quantity = stock.quantity + $2
	`, productStock.ProductID, productStock.Quantity)
	if err != nil {
		return fmt.Errorf("update product stock: %w", err)
	}
	return nil
}

func removeProductsFromStock(
	products map[uuid.UUID]int,
	tx *sqlx.Tx,
) error {
	missingProducts := make(map[uuid.UUID]int)

	for productID, quantity := range products {
		quantityInStock := 0

		err := tx.Get(
			&quantityInStock,
			"SELECT quantity FROM stock WHERE product_id = $1",
			productID,
		)
		if err != nil {
			return fmt.Errorf("failed to get quantity in stock: %w", err)
		}

		if quantityInStock < quantity {
			missingProducts[productID] = quantity - quantityInStock
		}

		if len(missingProducts) > 0 {
			continue
		}
	}

	if len(missingProducts) > 0 {
		return ProductsOutOfStockError{MissingProducts: missingProducts}
	}

	for productID, quantity := range products {
		_, err := tx.Exec(
			"UPDATE stock SET quantity = quantity - $1 WHERE product_id = $2",
			quantity,
			productID,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}
	}

	return nil
}

type ProductsOutOfStockError struct {
	MissingProducts map[uuid.UUID]int
}

func (p ProductsOutOfStockError) Error() string {
	return "products out of stock"
}
