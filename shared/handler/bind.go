// Package handler provides common utilities for HTTP handlers.
package handler

import (
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/shared/errors"
)

// BindRequest reads the request body, parses it into the target type using gopantic,
// and validates it. Returns the parsed value or a user-friendly error.
//
// Usage:
//
//	req, bindErr := handler.BindRequest[CreateTransferRequest](r)
//	if bindErr != nil {
//	    response.Error(w, bindErr)
//	    return
//	}
func BindRequest[T any](r *http.Request) (T, *errors.Error) {
	var zero T

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return zero, errors.BadRequest("failed to read request body")
	}
	defer func() { _ = r.Body.Close() }()

	result, parseErr := model.ParseInto[T](body)
	if parseErr != nil {
		return zero, errors.Validation(parseErr.Error())
	}

	return result, nil
}
