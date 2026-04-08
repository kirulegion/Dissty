package domain

import "errors"

var (
    ErrUnauthorized    = errors.New("unauthorized")
    ErrInvalidAction   = errors.New("invalid action")
    ErrInvalidRole     = errors.New("invalid role")
    ErrInvalidContext  = errors.New("invalid context")
)
