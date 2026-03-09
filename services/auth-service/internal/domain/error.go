package domain

import "errors"

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrWrongOTP           = errors.New("otp does not match")
	ErrExpiredOTP         = errors.New("otp has expired")
	ErrOTPAlreadyUsed     = errors.New("otp has been used")
	ErrProviderLinkExists = errors.New("provider  is already linked")
	ErrProviderNotFound   = errors.New("provider not found")
)
