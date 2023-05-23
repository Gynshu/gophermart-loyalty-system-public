package handlers

import (
	"context"
	"errors"
	"github.com/gynshu-one/gophermart-loyalty-system/external"
	"github.com/gynshu-one/gophermart-loyalty-system/models"
	"github.com/gynshu-one/gophermart-loyalty-system/pgadapter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_handler_RegisterHandler(t *testing.T) {
	type fields struct {
		user pgadapter.UserAdapter
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name: "Invalid request body",
			fields: fields{
				user: mockUserAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/register", strings.NewReader("invalid json")),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Empty login or password",
			fields: fields{
				user: mockUserAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/register", strings.NewReader(`{"Login": "", "Password": ""}`)),
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "User already exists",
			fields: fields{
				user: mockUserAdapter{
					err: models.ErrorUserAlreadyExists,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/register", strings.NewReader(`{"Login": "test", "Password": "test"}`)),
			},
			wantStatusCode: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				user: tt.fields.user,
			}
			h.RegisterHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.RegisterHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func Test_handler_LoginHandler(t *testing.T) {
	type fields struct {
		user pgadapter.UserAdapter
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name: "Invalid request body",
			fields: fields{
				user: mockUserAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/login", strings.NewReader("invalid json")),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Empty login or password",
			fields: fields{
				user: mockUserAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/login", strings.NewReader(`{"Login": "", "Password": ""}`)),
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "Invalid login or password",
			fields: fields{
				user: mockUserAdapter{
					err: errors.New("invalid login or password"),
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/login", strings.NewReader(`{"Login": "test", "Password": "test"}`)),
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				user: tt.fields.user,
			}
			h.LoginHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.LoginHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func Test_handler_AddOrderHandler(t *testing.T) {
	type fields struct {
		order          pgadapter.OrderAdapter
		accrualAdapter external.AccrualAdapter
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name: "Incorrect order ID",
			fields: fields{
				order: mockOrderAdapter{},
				accrualAdapter: mockAccrualAdapter{
					err: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/add_order", strings.NewReader(`"inc"`)).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Internal error",
			fields: fields{
				order: mockOrderAdapter{},
				accrualAdapter: mockAccrualAdapter{
					err: errors.New("internal error"),
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/add_order", strings.NewReader(`12345678903`)),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Order already added by the this user",
			fields: fields{
				order: mockOrderAdapter{
					order: &models.Order{
						ID:     "12345678903",
						UserID: "user_id",
					},
				},
				accrualAdapter: mockAccrualAdapter{
					err: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/add_order", strings.NewReader(`12345678903`)).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Order already added by another user",
			fields: fields{
				order: mockOrderAdapter{
					order: &models.Order{
						ID:     "12345678903",
						UserID: "another_user_id",
					},
				},
				accrualAdapter: mockAccrualAdapter{
					err: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/add_order", strings.NewReader(`12345678903`)).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				order:          tt.fields.order,
				accrualAdapter: tt.fields.accrualAdapter,
			}
			h.AddOrderHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.AddOrderHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func Test_handler_GetOrderHandler(t *testing.T) {
	type fields struct {
		order pgadapter.OrderAdapter
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name: "No orders for user",
			fields: fields{
				order: mockOrderAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "Internal server error",
			fields: fields{
				order: mockOrderAdapter{
					order: &models.Order{
						ID:     "order_id",
						UserID: "user_id",
						Status: "status",
					},
					err: errors.New("internal error"),
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Successfully retrieved orders",
			fields: fields{
				order: mockOrderAdapter{
					order: &models.Order{
						ID:     "order_id",
						UserID: "user_id",
						Status: "status",
					},
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				order: tt.fields.order,
			}
			h.GetOrderHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.GetOrderHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}
		})
	}
}

func Test_handler_GetBalanceHandler(t *testing.T) {
	type fields struct {
		balance pgadapter.BalanceAdapter
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedStatus int
	}{
		{
			name: "Test success case",
			fields: fields{
				balance: mockBalanceAdapter{
					balance: &models.Balance{Amount: 1000.0, Withdrawn: 500.0},
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user1")),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Test no balance case",
			fields: fields{
				balance: mockBalanceAdapter{
					balance: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user1")),
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Test internal error case",
			fields: fields{
				balance: mockBalanceAdapter{
					balance: &models.Balance{Amount: 1000.0, Withdrawn: 500.0},
					err:     errors.New("internal error"),
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/get_order", nil).WithContext(context.WithValue(context.Background(), models.UserID, "user1")),
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				balance: tt.fields.balance,
			}
			h.GetBalanceHandler(tt.args.w, tt.args.r)
			res := tt.args.w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.expectedStatus {
				t.Errorf("handler.GetBalanceHandler() expected status = %v, got = %v", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}

func Test_handler_WithdrawBalanceHandler(t *testing.T) {
	type fields struct {
		balance    pgadapter.BalanceAdapter
		order      pgadapter.OrderAdapter
		withdrawal pgadapter.WithdrawalAdapter
		accrual    external.AccrualAdapter
	}
	type args struct {
		w    *httptest.ResponseRecorder
		r    *http.Request
		code int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name:   "Invalid json body",
			fields: fields{},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/withdraw", strings.NewReader(`"}{""`)).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "Incorrect order ID",
			fields: fields{},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/withdraw", strings.NewReader(`{
																								"order": "237722562",
																								"sum": 751
																								} `)).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Order Already registered",
			fields: fields{
				order: mockOrderAdapter{
					order: &models.Order{
						ID:     "order_id",
						UserID: "user_id",
					},
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/withdraw", strings.NewReader(`{
																								"order": "order_id",
																								"sum": 751
																								} `)).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Insufficient funds",
			fields: fields{
				order: mockOrderAdapter{
					order: nil,
				},
				balance: mockBalanceAdapter{
					err: models.ErrorInsufficientFunds,
				},
				accrual: mockAccrualAdapter{
					err: nil,
				},
				withdrawal: mockWithdrawalAdapter{
					err: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/withdraw", strings.NewReader(`{
																								"order": "2377225624",
																								"sum": 751
																								} `)).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
		{
			name: "OK 200",
			fields: fields{
				order: mockOrderAdapter{
					order: nil,
				},
				balance: mockBalanceAdapter{
					balance: nil,
					err:     models.ErrorInsufficientFunds,
				},
				accrual: mockAccrualAdapter{
					err: nil,
				},
				withdrawal: mockWithdrawalAdapter{
					err: nil,
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "/withdraw", strings.NewReader(`{
																								"order": "2377225624",
																								"sum": 751
																								} `)).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusPaymentRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				balance:        tt.fields.balance,
				order:          tt.fields.order,
				withdrawal:     tt.fields.withdrawal,
				accrualAdapter: tt.fields.accrual,
			}
			h.WithdrawBalanceHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.WithdrawBalanceHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}

		})
	}
}

func TestGetWithdrawalsHandler(t *testing.T) {
	type fields struct {
		withdrawal pgadapter.WithdrawalAdapter
	}
	type args struct {
		w    *httptest.ResponseRecorder
		r    *http.Request
		code int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name: "Show withdrawals",
			fields: fields{
				withdrawal: mockWithdrawalAdapter{
					withdrawal: []*models.Withdrawal{
						{
							Sum:         100,
							OrderID:     "order_id",
							ProcessedAt: time.Now(),
						},
					},
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/withdraw", nil).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "Noting to show",
			fields: fields{
				withdrawal: mockWithdrawalAdapter{},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/withdraw", nil).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "Internal error",
			fields: fields{
				withdrawal: mockWithdrawalAdapter{
					err: errors.New("error"),
				},
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/withdraw", nil).
					WithContext(context.WithValue(context.Background(), models.UserID, "user_id")),
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				withdrawal: tt.fields.withdrawal,
			}
			h.GetWithdrawalsHandler(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.wantStatusCode {
				t.Errorf("handler.GetWithdrawalsHandler() error = %v, wantErr %v", tt.args.w.Code, tt.wantStatusCode)
			}

		})
	}
}
