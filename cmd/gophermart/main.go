package main

import (
	"context"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/gynshu-one/gophermart-loyalty-system/config"
	"github.com/gynshu-one/gophermart-loyalty-system/external"
	"github.com/gynshu-one/gophermart-loyalty-system/handlers"
	"github.com/gynshu-one/gophermart-loyalty-system/middlwares"
	"github.com/gynshu-one/gophermart-loyalty-system/pgadapter"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var (
	sessionManager *scs.SessionManager
	handler        handlers.Handler
	balance        pgadapter.BalanceAdapter
	order          pgadapter.OrderAdapter
	user           pgadapter.UserAdapter
	withdrawal     pgadapter.WithdrawalAdapter
	db             *sqlx.DB
)

func init() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * 30 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = false // set to true when using HTTPS

}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db = pgadapter.NewConnection(ctx)
	// Order matters, first should be user, then balance, then order
	user = pgadapter.NewAdapter(ctx, db)
	balance = pgadapter.NewBalanceAdapter(ctx, db)
	order = pgadapter.NewOrderAdapter(ctx, db)
	withdrawal = pgadapter.NewWithdrawalAdapter(ctx, db)
	accrualAdapter := external.Start(config.GetConfig().AccrualSystemAddress, order, balance)
	handler = handlers.NewHandler(balance,
		order,
		user,
		withdrawal,
		accrualAdapter)

	r := chi.NewRouter()
	r.Route("/api/user", func(r chi.Router) {
		r.Use(Logger)
		r.Post("/login", handler.LoginHandler)
		r.Post("/register", handler.RegisterHandler)

		r.With(middlwares.AuthMiddleware).Post("/orders", handler.AddOrderHandler)
		r.With(middlwares.AuthMiddleware).Get("/orders", handler.GetOrderHandler)
		r.With(middlwares.AuthMiddleware).Get("/balance", handler.GetBalanceHandler)
		r.With(middlwares.AuthMiddleware).Post("/balance/withdraw", handler.WithdrawBalanceHandler)
		r.With(middlwares.AuthMiddleware).Get("/withdrawals", handler.GetWithdrawalsHandler)

	})
	http.ListenAndServe(config.GetConfig().RunAddress, r)
}
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Debug().Msgf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Debug().Msgf("Completed in %v", time.Since(start))
	})
}
