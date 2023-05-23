package models

import "github.com/gynshu-one/gophermart-loyalty-system/pgadapter/composer"

const (
	ID     = composer.Field("id")
	UserID = composer.Field("user_id")
	Status = composer.Field("status")
)
