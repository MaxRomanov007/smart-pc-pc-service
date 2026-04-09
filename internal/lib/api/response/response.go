package response

import "go/types"

type Response[T any] struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   *T     `json:"data,omitempty"`
}

func New[T any](status string, data *T, error string) *Response[T] {
	return &Response[T]{
		Status: status,
		Error:  error,
		Data:   data,
	}
}

const (
	StatusOK            = "ok"
	StatusBadRequest    = "bad-request"
	StatusUnauthorized  = "unauthorized"
	StatusForbidden     = "forbidden"
	StatusInternalError = "internal-error"
)

func OK[T any](data *T) Response[T] {
	return Response[T]{
		Status: StatusOK,
		Data:   data,
	}
}

func Error(status, msg string) Response[types.Nil] {
	return Response[types.Nil]{
		Status: status,
		Error:  msg,
	}
}

func BadRequest(msg string) Response[types.Nil] {
	return Error(StatusBadRequest, msg)
}

func Unauthorized(msg string) Response[types.Nil] {
	if msg == "" {
		return Error(StatusUnauthorized, "Unauthorized")
	}
	return Error(StatusUnauthorized, "Unauthorized: "+msg)
}

func Forbidden(msg string) Response[types.Nil] {
	if msg == "" {
		return Error(StatusForbidden, "Forbidden")
	}
	return Error(StatusForbidden, "Forbidden: "+msg)
}

func InternalError() Response[types.Nil] {
	return Error(StatusInternalError, "Internal error")
}
