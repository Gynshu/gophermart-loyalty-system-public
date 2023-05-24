package pgadapter

import (
	"context"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	CreateBalanceSchema = `
    CREATE TABLE IF NOT EXISTS balances (
        id VARCHAR(255) NOT NULL PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL REFERENCES users(id),
        amount FLOAT NOT NULL,
        withdrawn FLOAT NOT NULL                                
    );`
	selectAmount    = `SELECT amount FROM balances WHERE user_id = $1`
	updateBalance   = `UPDATE balances SET amount = amount + $1 WHERE user_id = $2`
	updateWithdrawn = `UPDATE balances SET withdrawn = withdrawn + $1, amount = amount - $1 WHERE user_id = $2 AND amount >= $1;`
	readBalance     = `SELECT id, user_id, amount, withdrawn FROM balances WHERE user_id = $1;`
)

type BalanceAdapter interface {
	ReadBalance(ctx context.Context, userID string) (*models.Balance, error)
	IncrementBalance(ctx context.Context, userID string, amount float64) error
	IncrementWithdrawn(ctx context.Context, userID string, amount float64) error
}
type balanceAdapter struct {
	conn *sqlx.DB
	BalanceAdapter
}

func NewBalanceAdapter(ctx context.Context, conn *sqlx.DB) *balanceAdapter {
	b := &balanceAdapter{conn: conn}
	err := b.createBalanceSchema(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create balance schema")
	}
	return b
}

func (b *balanceAdapter) ReadBalance(ctx context.Context, userID string) (*models.Balance, error) {
	var balance []*models.Balance
	err := b.conn.SelectContext(ctx, &balance, readBalance, userID)
	if len(balance) == 0 {
		return nil, nil
	}
	return balance[0], err
}
func (b *balanceAdapter) IncrementBalance(ctx context.Context, userID string, amount float64) error {
	_, err := b.conn.ExecContext(ctx, updateBalance, amount, userID)
	return err
}

// IncrementWithdrawn Increments Withdrawal and Decrements balance (if possible)
func (b *balanceAdapter) IncrementWithdrawn(ctx context.Context, userID string, amount float64) error {
	tx, err := b.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// Update withdrawn only if amount is available
	result, err := tx.ExecContext(ctx, updateWithdrawn, amount, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if rows == 0 {
		_ = tx.Rollback()
		return models.ErrorInsufficientFunds
	}
	return tx.Commit()
}
func (b *balanceAdapter) createBalanceSchema(ctx context.Context) error {
	_, err := b.conn.ExecContext(ctx, CreateBalanceSchema)
	return err
}
