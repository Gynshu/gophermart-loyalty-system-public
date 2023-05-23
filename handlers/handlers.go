package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gynshu-one/gophermart-loyalty-system/external"
	"github.com/gynshu-one/gophermart-loyalty-system/helpers"
	"github.com/gynshu-one/gophermart-loyalty-system/middlwares"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/gynshu-one/gophermart-loyalty-system/pgadapter"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type Handler interface {
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	AddOrderHandler(w http.ResponseWriter, r *http.Request)
	GetOrderHandler(w http.ResponseWriter, r *http.Request)
	GetBalanceHandler(w http.ResponseWriter, r *http.Request)
	WithdrawBalanceHandler(w http.ResponseWriter, r *http.Request)
	GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
}
type handler struct {
	balance        pgadapter.BalanceAdapter
	order          pgadapter.OrderAdapter
	user           pgadapter.UserAdapter
	accrualAdapter external.AccrualAdapter
	withdrawal     pgadapter.WithdrawalAdapter
}

func NewHandler(balance pgadapter.BalanceAdapter,
	order pgadapter.OrderAdapter,
	user pgadapter.UserAdapter,
	withdrawal pgadapter.WithdrawalAdapter,
	orderService external.AccrualAdapter) Handler {
	return &handler{
		balance:        balance,
		order:          order,
		accrualAdapter: orderService,
		user:           user,
		withdrawal:     withdrawal,
	}
}
func (h *handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user *models.User

	// Read request body
	body, err := helpers.ReadBodyAsBytes(r.Body)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Unmarshal to user struct
	if err = json.Unmarshal(body, &user); err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Basic check
	if user.Login == "" || user.Password == "" {
		log.Debug().Msg("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Create and check user
	user.ID = helpers.GenerateUUID()
	if err = h.user.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, models.ErrorUserAlreadyExists) {
			log.Debug().Msg("User already exists")
			http.Error(w, "User already exists", http.StatusConflict)
			return
		} else {
			log.Debug().Msgf("Internal server error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Authorize user
	jwt, err := middlwares.GenerateJWT(user.ID)
	if err != nil {
		log.Debug().Msgf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    jwt,
		HttpOnly: true,
		Secure:   false,
	})

	// Return response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registered, Please login!"))
}

func (h *handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user *models.User
	defer r.Body.Close()

	// Read request body
	body, err := helpers.ReadBodyAsBytes(r.Body)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Unmarshal to user struct
	if err = json.Unmarshal(body, &user); err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}
	if user.Login == "" || user.Password == "" {
		log.Debug().Msg("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Save password before hashing
	pass := user.Password

	// Read user from db
	user, err = h.user.ReadUser(r.Context(), user.Login)

	// Check password
	if err != nil || !helpers.CheckPasswordHash(pass, user.Password) {
		log.Debug().Msgf("Invalid username or password: %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Authorize user
	jwt, err := middlwares.GenerateJWT(user.ID)
	if err != nil {
		log.Debug().Msgf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    jwt,
		HttpOnly: true,
		Secure:   false,
	})

	// Return response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in!"))
}
func (h *handler) AddOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(models.UserID).(string)
	defer r.Body.Close()

	// Read OrderID as plain text
	OrderID, err := helpers.ReadBodyAsString(r.Body)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Lunar func also checks if order is not empty
	ok := helpers.LunaOrderCheck(OrderID)
	if !ok {
		log.Debug().Msgf("Wrong order id %s", OrderID)
		http.Error(w, "Wrong order id", http.StatusUnprocessableEntity)
		return
	}

	// Check if order already exists
	order, err := h.order.ReadOrder(r.Context(), models.ID.EqualTo(OrderID))
	if err != nil {
		log.Debug().Msgf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If already added by this user
	if order != nil {
		if order[0].UserID == userID {
			log.Debug().Msg("Order already added by this user")
			http.Error(w, "Order already added by this user", http.StatusOK)
			return
		} else {
			log.Debug().Msg("Order already added by another user")
			http.Error(w, "Order already added by another user", http.StatusConflict)
			return
		}
	}

	// Add order to db and fallow it until processed
	err = h.accrualAdapter.FallowOrder(&models.Order{
		ID:     OrderID,
		UserID: userID,
	})
	if err != nil {
		log.Debug().Msgf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Order created and processing"))
}

func (h *handler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(models.UserID).(string)

	// Find orders by user id
	orders, err := h.order.ReadOrder(r.Context(), models.UserID.EqualTo(userID))
	if orders == nil {
		log.Debug().Msgf("No orders for user %s", userID)
		http.Error(w, "No orders", http.StatusNoContent)
		return
	}
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Convert DB orders to more suitable for client
	responseOrders := make([]models.ResponseOrder, 0, len(orders))
	for _, order := range orders {
		o := models.ResponseOrder{
			Number:     order.ID,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}
		if order.Accrual != 0 {
			o.Accrual = order.Accrual
		}
		responseOrders = append(responseOrders, o)
	}

	// Pack
	ordersJSON, err := json.Marshal(responseOrders)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Send
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ordersJSON)
}

func (h *handler) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(models.UserID).(string)

	// Find balance by user id
	balance, err := h.balance.ReadBalance(r.Context(), userID)
	if balance == nil {
		log.Debug().Msgf("No balance for user %s", userID)
		http.Error(w, "No balance", http.StatusNoContent)
		return
	}
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Send
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"current": %f, "withdrawn": %f}`, balance.Amount, balance.Withdrawn)))
}

func (h *handler) WithdrawBalanceHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	userID, _ := r.Context().Value(models.UserID).(string)

	// Parse JSON request body
	var bodyJSON struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	err := json.NewDecoder(r.Body).Decode(&bodyJSON)
	if err != nil || bodyJSON.Sum <= 0 {
		log.Debug().Msgf("Bad request: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Check order validity
	if !helpers.LunaOrderCheck(bodyJSON.Order) {
		log.Debug().Msgf("Wrong order id %s", bodyJSON.Order)
		http.Error(w, "Wrong order id", http.StatusUnprocessableEntity)
		return
	}

	// Check if order is already registered
	order, err := h.order.ReadOrder(r.Context(), models.ID.EqualTo(bodyJSON.Order))
	if err != nil {
		log.Debug().Msgf("Internal server error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if order != nil {
		log.Debug().Msgf("Order already registered %s", bodyJSON.Order)
		http.Error(w, "Wrong order id", http.StatusUnprocessableEntity)
		return
	}

	// Increment withdrawn balance
	err = h.balance.IncrementWithdrawn(r.Context(), userID, bodyJSON.Sum)
	if err != nil {
		if errors.Is(err, models.ErrorInsufficientFunds) {
			log.Debug().Msg("Insufficient funds")
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Fallow order continuously
	err = h.accrualAdapter.FallowOrder(&models.Order{
		ID:     bodyJSON.Order,
		UserID: userID,
	})
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Create withdrawal record
	withdrawal := &models.Withdrawal{
		ID:          uuid.New().String(),
		UserID:      userID,
		OrderID:     bodyJSON.Order,
		Sum:         bodyJSON.Sum,
		ProcessedAt: time.Now(),
	}
	if err = h.withdrawal.CreateWithdrawal(r.Context(), withdrawal); err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(models.UserID).(string)

	// Find withdrawals by user id
	withdrawals, err := h.withdrawal.ReadWithdrawal(r.Context(), userID)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}
	if withdrawals == nil {
		log.Debug().Msgf("No withdrawals for user %s", userID)
		http.Error(w, "No withdrawals", http.StatusNoContent)
		return
	}

	// Remove sensitive data
	for _, withdrawal := range withdrawals {
		withdrawal.ID = ""
		withdrawal.UserID = ""
	}

	// Pack
	withdrawalsJSON, err := json.Marshal(withdrawals)
	if err != nil {
		log.Debug().Msgf("%s: %v", models.ErrorInternal.Error(), err)
		http.Error(w, models.ErrorInternal.Error(), http.StatusInternalServerError)
		return
	}

	// Send
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(withdrawalsJSON)
}
