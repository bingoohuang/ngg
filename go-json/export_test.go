package json

import (
	"github.com/bingoohuang/ngg/go-json/internal/errors"
)

var (
	NewSyntaxError    = errors.ErrSyntax
	NewMarshalerError = errors.ErrMarshaler
)
