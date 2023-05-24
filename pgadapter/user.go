package pgadapter

import (
	"context"
	"github.com/google/uuid"
	"github.com/gynshu-one/gophermart-loyalty-system/helpers"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"strings"
)

const (
	createBalance = `INSERT INTO balances (id, user_id, amount, withdrawn) VALUES ($1, $2, $3, $4);`
	createUser    = `INSERT INTO users (id, login, password) VALUES ($1, $2, $3);`
	readUser      = `SELECT id, login, password FROM users WHERE login = $1;`
)
const (
	CreateUserSchema = `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) NOT NULL PRIMARY KEY,
        login VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL
    );`
)

type UserAdapter interface {
	CreateUser(ctx context.Context, user *models.User) error
	ReadUser(ctx context.Context, id string) (*models.User, error)
}
type userAdapter struct {
	conn *sqlx.DB
	UserAdapter
}

func NewAdapter(ctx context.Context, conn *sqlx.DB) *userAdapter {
	a := &userAdapter{conn: conn}
	err := a.createUserSchema(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create user schema")
	}
	return a
}

func (u *userAdapter) CreateUser(ctx context.Context, user *models.User) error {
	hashedPassword, err := helpers.HashPassword(user.Password)
	if err != nil {
		log.Warn().Err(err).Msg("failed to hash password")
		hashedPassword = user.Password
	}
	tx, err := u.conn.BeginTxx(ctx, nil)

	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = u.conn.ExecContext(ctx, createUser, user.ID, user.Login, hashedPassword)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return models.ErrorUserAlreadyExists
		}
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, createBalance, uuid.New().String(), user.ID, 0, 0)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (u *userAdapter) ReadUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := u.conn.GetContext(ctx, user, readUser, id)
	return user, err
}

func (u *userAdapter) createUserSchema(ctx context.Context) error {
	_, err := u.conn.ExecContext(ctx, CreateUserSchema)
	return err
}
