package service

import (
	"errors"
	"net/http"
)

var (
	ErrVarNotFound   = errors.New("ErrVarNotFound")
	ErrPathDuplicate = errors.New("ErrPathDuplicate")
)

const (
	ErrPathDuplicateResponseStatusCode = iota + http.StatusUnavailableForLegalReasons + 1
)
