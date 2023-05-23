package pgadapter

import (
	"context"
	"github.com/gynshu-one/gophermart-loyalty-system/helpers"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	CreateWithdrawalSchema = `
    CREATE TABLE IF NOT EXISTS withdrawals (
        id VARCHAR(255) NOT NULL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL REFERENCES users(id),
        order_id VARCHAR(255) NOT NULL REFERENCES orders(id),
        sum FLOAT NOT NULL,
        processed_at DATE NOT NULL
    );`
	selectWithdrawal = `SELECT id, user_id, order_id, sum, processed_at FROM withdrawals WHERE user_id = $1;`
	createWithdrawal = `INSERT INTO withdrawals (id, user_id, order_id, sum, processed_at) VALUES ($1, $2, $3, $4, $5);`
)

type WithdrawalAdapter interface {
	CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error
	ReadWithdrawal(ctx context.Context, userID string) ([]*models.Withdrawal, error)
}
type withdrawalAdapter struct {
	conn *sqlx.DB
	WithdrawalAdapter
}

func NewWithdrawalAdapter(ctx context.Context, conn *sqlx.DB) *withdrawalAdapter {
	w := &withdrawalAdapter{conn: conn}
	err := w.createWithdrawalSchema(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create withdrawal schema")
	}
	return w
}

func (w *withdrawalAdapter) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	withdrawal.ID = helpers.GenerateUUID()
	_, err := w.conn.ExecContext(ctx, createWithdrawal, withdrawal.ID, withdrawal.UserID, withdrawal.OrderID, withdrawal.Sum, withdrawal.ProcessedAt)
	return err
}

func (w *withdrawalAdapter) ReadWithdrawal(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	var withdrawal []*models.Withdrawal
	err := w.conn.SelectContext(ctx, &withdrawal, selectWithdrawal, userID)
	return withdrawal, err
}

func (w *withdrawalAdapter) createWithdrawalSchema(ctx context.Context) error {
	_, err := w.conn.ExecContext(ctx, CreateWithdrawalSchema)
	return err
}
