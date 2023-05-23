package models

import "errors"

var (
	ErrorUserAlreadyExists    = errors.New("user already exists")
	ErrorInternal             = errors.New("internal error")
	ErrorOrderNotRegistered   = errors.New("order not registered")
	ErrorRequestLimitExceeded = errors.New("request limit exceeded")
	ErrorServiceInternalError = errors.New("loyalty service internal error")
	ErrorInsufficientFunds    = errors.New("insufficient funds")
)
