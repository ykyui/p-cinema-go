package service

import (
	"errors"
	"net/http"
)

var (
	ErrVarNotFound   = errors.New("ErrVarNotFound")
	ErrPathDuplicate = errors.New("ErrPathDuplicate")
	ErrCantBuy       = errors.New("ErrCantBuy")
)

const (
	ErrPathDuplicateResponseStatusCode = iota + http.StatusUnavailableForLegalReasons + 1
)
