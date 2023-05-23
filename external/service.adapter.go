package external

import (
	"context"
	"github.com/gammazero/workerpool"
	resty "github.com/go-resty/resty/v2"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/gynshu-one/gophermart-loyalty-system/pgadapter"
	"github.com/gynshu-one/gophermart-loyalty-system/pgadapter/composer"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"time"
)

type AccrualAdapter interface {
	FallowOrder(order *models.Order) error
}

// orderService is an independent service that continuously updates state of orders in db
type orderService struct {
	addr     string
	orders   pgadapter.OrderAdapter
	balances pgadapter.BalanceAdapter
	client   *resty.Client
}

func Start(addr string, orders pgadapter.OrderAdapter, balances pgadapter.BalanceAdapter) *orderService {
	client := resty.New()
	e := &orderService{
		addr:     addr,
		orders:   orders,
		balances: balances,
		client:   client,
	}

	// Read all orders that are not processed yet from DB
	// In case of service restart, we will continue from the point we stopped
	all, err := e.orders.ReadOrder(context.Background(), composer.CondGroup(
		models.Status.NotEqualTo(models.OrderStatusProcessed), composer.And(),
		models.Status.NotEqualTo(models.OrderStatusInvalid),
	))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read orders")
	}

	// Fallow each again
	if len(all) > 0 {
		for _, order := range all {
			e.FallowOrder(order)
			time.Sleep(200 * time.Millisecond)
		}
	}
	return e
}

type responseStruct struct {
	OrderID    string  `json:"order"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	RetryAfter int     `json:"retry_after"`
}

func (e *orderService) FallowOrder(order *models.Order) error {
	ctx := context.Background()

	// If order passed from Start() function, it will have status
	// Otherwise order is new
	if order.Status == "" {
		order.UploadedAt = time.Now()
		order.Status = models.OrderStatusNew
		order.UpdatedAt = time.Now()
		ctx_, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		err := e.orders.CreateOrder(ctx_, order)
		cancel()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create order")
			return err
		}
	}

	// Pass order to worker pool
	workerPool.Submit(func() {
		e.worker(ctx, order)
	})
	return nil
}

var workerPool = workerpool.New(50)

func (e *orderService) worker(ctx context.Context, order *models.Order) {
	response := e.check(*order)

	// If response is nil - we will try again later (internal errors)
	if response == nil {
		time.Sleep(300 * time.Millisecond)
		e.FallowOrder(order)
		return
	}

	// If rate limit is reached - we will try again later
	if response.RetryAfter > 0 {
		ctx_, cancel := context.WithTimeout(ctx, time.Duration(response.RetryAfter)*time.Second)
		defer cancel()
		workerPool.Pause(ctx_)
		e.FallowOrder(order)
		return
	}

	// No changes
	if response.Status == order.Status {
		time.Sleep(300 * time.Millisecond)
		e.FallowOrder(order)
		return
	}

	// Update order status and increment balance
	if response.Status == models.OrderStatusProcessed {
		ctx_, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		if response.Accrual > 0 {
			e.updateBalance(ctx_, order.UserID, response.Accrual)
		}
		cancel()
		order.Accrual = response.Accrual
		order.Status = models.OrderStatusProcessed
	}
	order.UpdatedAt = time.Now()
	err := e.orders.UpdateOrders(context.Background(), order)
	if err != nil {
		log.Error().Err(err).Msg("failed to update order")
		return
	}

	// If order is Invalid or Processed - we will not check it again
	if response.Status == models.OrderStatusInvalid || response.Status == models.OrderStatusProcessed {
		return
	}
	time.Sleep(300 * time.Millisecond)
	e.FallowOrder(order)
}
func (e *orderService) check(order models.Order) *responseStruct {
	var response *responseStruct
	res, err := e.client.R().
		SetResult(&response).
		Get(e.addr + "/api/orders/" + order.ID)

	if err != nil {
		log.Error().Err(models.ErrorInternal).Err(err).Msgf("error while checking order %s", order.ID)
		return nil
	}

	switch res.StatusCode() {
	case http.StatusNoContent:
		log.Warn().Err(models.ErrorOrderNotRegistered).Err(err).Msgf("error while checking order %s", order.ID)
		return nil
	case http.StatusTooManyRequests:
		retryAfter, err_ := strconv.ParseInt(res.Header().Get("Retry-After"), 10, 64)

		if err_ != nil {
			log.Info().Msg("Retry-After header is not set")
		} else {
			response.RetryAfter = int(retryAfter)
			return response
		}
		log.Warn().Err(models.ErrorRequestLimitExceeded).Msgf("error while checking order %s", order.ID)
		return nil
	case http.StatusInternalServerError:
		log.Debug().Err(models.ErrorServiceInternalError).Msgf("error while checking order %s", order.ID)
		return nil
	}
	return response
}

func (e *orderService) updateBalance(ctx context.Context, userID string, addAmount float64) {
	err := e.balances.IncrementBalance(ctx, userID, addAmount)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update balance")
		return
	}
}
