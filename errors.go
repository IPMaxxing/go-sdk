package ipmax

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ErrorCode int

const (
	ErrorCodeInvalidRequest         ErrorCode = 1000
	ErrorCodeInvalidPayload         ErrorCode = 1001
	ErrorCodeInvalidIP              ErrorCode = 1002
	ErrorCodeIPNotFound             ErrorCode = 1003
	ErrorCodeRouteNotFound          ErrorCode = 1004
	ErrorCodeInvalidAPIKey          ErrorCode = 1010
	ErrorCodeRateLimitExceeded      ErrorCode = 1011
	ErrorCodeServiceUnavailable     ErrorCode = 1400
	ErrorCodeInternalError          ErrorCode = 1401
	ErrorCodeInsufficientBalance    ErrorCode = 1500
	ErrorCodeIdempotencyKeyConflict ErrorCode = 1501
)

var (
	ErrInvalidRequest      = errors.New("ipmax: invalid request")
	ErrUnauthorized        = errors.New("ipmax: unauthorized")
	ErrInsufficientBalance = errors.New("ipmax: insufficient balance")
	ErrNotFound            = errors.New("ipmax: not found")
	ErrConflict            = errors.New("ipmax: conflict")
	ErrRateLimited         = errors.New("ipmax: rate limited")
	ErrServer              = errors.New("ipmax: server error")
)

type APIError struct {
	StatusCode int           `json:"-"`
	Code       ErrorCode     `json:"code"`
	Message    string        `json:"message"`
	RequestID  string        `json:"request_id"`
	Retryable  bool          `json:"retryable"`
	RetryAfter time.Duration `json:"-"`
}

func (e *APIError) Error() string {
	text := fmt.Sprintf("ipmax: %d %s", e.StatusCode, e.Message)
	if e.Code != 0 {
		text += fmt.Sprintf(" (code %d)", e.Code)
	}
	if e.RequestID != "" {
		text += " [request " + e.RequestID + "]"
	}
	return text
}

func (e *APIError) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusUnsupportedMediaType:
		return target == ErrInvalidRequest
	case http.StatusUnauthorized:
		return target == ErrUnauthorized
	case http.StatusPaymentRequired:
		return target == ErrInsufficientBalance
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusConflict:
		return target == ErrConflict
	case http.StatusTooManyRequests:
		return target == ErrRateLimited
	}
	return e.StatusCode >= 500 && target == ErrServer
}

func newAPIError(status int, header http.Header, body []byte) *APIError {
	var envelope ErrorResponse
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Error.Message == "" {
		envelope.Error = APIError{Message: http.StatusText(status), Retryable: retryableStatus(status)}
	}
	apiErr := envelope.Error
	apiErr.StatusCode = status
	if apiErr.RequestID == "" {
		apiErr.RequestID = header.Get("X-Request-Id")
	}
	if delay, ok := retryAfter(header); ok {
		apiErr.RetryAfter = delay
	}
	return &apiErr
}

func retryableStatus(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500
}
