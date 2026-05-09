package model

import "errors"

var (
	ErrInvalidAmount = errors.New("amount must be greater than zero")
)
