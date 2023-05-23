package pgadapter

import (
	"context"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	comp "github.com/gynshu-one/gophermart-loyalty-system/pgadapter/composer"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	createOrderSchema = `
    CREATE TABLE IF NOT EXISTS orders (
        id VARCHAR(255) NOT NULL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL REFERENCES users(id),
        status VARCHAR(255) NOT NULL,
        accrual FLOAT,
        uploaded_at DATE NOT NULL,
	    updated_at DATE NOT NULL
    );`
	createOrder = `INSERT INTO orders (id, user_id, status, accrual, uploaded_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6);`
	updateOrder = `UPDATE orders SET status = $1, accrual = $2, updated_at = $3 WHERE id = $4;`
)

type OrderAdapter interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	ReadOrder(ctx context.Context, condition comp.Condition) ([]*models.Order, error)
	UpdateOrders(ctx context.Context, orders ...*models.Order) error
}
type orderAdapter struct {
	conn *sqlx.DB
	OrderAdapter
}

func NewOrderAdapter(ctx context.Context, conn *sqlx.DB) *orderAdapter {
	o := &orderAdapter{conn: conn}
	err := o.createOrderSchema(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create withdrawal schema")
	}
	return o
}

func (o *orderAdapter) CreateOrder(ctx context.Context, order *models.Order) error {
	_, err := o.conn.ExecContext(ctx, createOrder, order.ID, order.UserID, order.Status, order.Accrual, order.UploadedAt, order.UpdatedAt)
	return err
}

func (o *orderAdapter) ReadOrder(ctx context.Context, condition comp.Condition) ([]*models.Order, error) {
	var orders []*models.Order
	cond, vars := condition.Build()
	stm := `SELECT id, user_id, status, accrual, uploaded_at, updated_at FROM orders WHERE ` + cond
	return orders, o.conn.SelectContext(ctx, &orders, stm, vars...)
}

func (o *orderAdapter) UpdateOrders(ctx context.Context, orders ...*models.Order) error {
	tx, err := o.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	smt, err := tx.PreparexContext(ctx, updateOrder)
	if err != nil {
		log.Error().Err(err).Msg("Unable to prepare transaction UpdateOrders")
		return err
	}
	defer smt.Close()
	for _, order := range orders {
		_, err = smt.ExecContext(ctx, order.Status, order.Accrual, order.UpdatedAt, order.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("Unable to commit transaction UpdateOrders")
		return err
	}
	return nil
}

func (o *orderAdapter) createOrderSchema(ctx context.Context) error {
	_, err := o.conn.ExecContext(ctx, createOrderSchema)
	return err
}
