package pgadapter

import (
	"context"
	"github.com/gynshu-one/gophermart-loyalty-system/config"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func NewConnection(ctx context.Context) *sqlx.DB {
	conn, err := sqlx.ConnectContext(ctx, "postgres", config.GetConfig().DBURI)
	if err != nil {
		log.Fatal().Msg("failed to connect to database")
	}
	return conn
}
