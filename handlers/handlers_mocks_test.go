package handlers

import (
	"context"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	comp "github.com/gynshu-one/gophermart-loyalty-system/pgadapter/composer"
)

type mockUserAdapter struct {
	err error
}

func (m mockUserAdapter) CreateUser(ctx context.Context, user *models.User) error {
	return m.err
}
func (m mockUserAdapter) ReadUser(ctx context.Context, login string) (*models.User, error) {
	return nil, m.err
}

type mockAccrualAdapter struct {
	err error
}

func (m mockAccrualAdapter) FallowOrder(order *models.Order) error {
	return m.err
}

type mockOrderAdapter struct {
	order *models.Order
	err   error
}

func (m mockOrderAdapter) ReadOrder(ctx context.Context, condition comp.Condition) ([]*models.Order, error) {
	var out []*models.Order
	if m.order != nil {
		out = append(out, m.order)
	}
	return out, m.err
}
func (m mockOrderAdapter) CreateOrder(ctx context.Context, order *models.Order) error {
	return nil
}
func (m mockOrderAdapter) UpdateOrders(ctx context.Context, orders ...*models.Order) error {
	return nil
}

type mockBalanceAdapter struct {
	balance *models.Balance
	err     error
}

func (m mockBalanceAdapter) ReadBalance(ctx context.Context, userID string) (*models.Balance, error) {
	return m.balance, m.err
}
func (m mockBalanceAdapter) IncrementBalance(ctx context.Context, userID string, amount float64) error {
	return m.err
}
func (m mockBalanceAdapter) IncrementWithdrawn(ctx context.Context, userID string, amount float64) error {
	return m.err
}

type mockWithdrawalAdapter struct {
	withdrawal []*models.Withdrawal
	err        error
}

func (m mockWithdrawalAdapter) CreateWithdrawal(ctx context.Context, withdrawal *models.Withdrawal) error {
	return nil
}
func (m mockWithdrawalAdapter) ReadWithdrawal(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	return m.withdrawal, m.err
}
