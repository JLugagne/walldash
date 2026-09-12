package inbound

import (
	"encoding/json"
	"net/http"

	rootDomain "github.com/JLugagne/walldash/domain"
	"github.com/JLugagne/walldash/internal/dashboard/domain"
	"github.com/JLugagne/walldash/internal/pkg/logger"
)

// ResponseSuccess represents the JSend success response payload.
type ResponseSuccess struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

// ResponseFail represents the JSend fail response payload for validation or client errors.
type ResponseFail struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

// ResponseError represents the JSend error response payload for server errors.
type ResponseError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Error details included in fail/error responses.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Controller provides helper methods to send standardized JSON responses.
type Controller struct{}

// NewController creates a new Controller instance.
func NewController() *Controller {
	return &Controller{}
}

// SendSuccess sends a success response with 200 OK status code.
func (c *Controller) SendSuccess(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResponseSuccess{
		Status: "success",
		Data:   data,
	})
}

// SendFail sends a failure response for client/pre-condition errors with 400 Bad Request.
func (c *Controller) SendFail(w http.ResponseWriter, r *http.Request, data any, err error) {
	log := logger.LoggerFromContext(r.Context())
	if err != nil {
		log.WithError(err).Warn("request validation or pre-condition failed")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	res := ResponseFail{
		Status: "fail",
		Data:   data,
	}

	if err != nil {
		if domainErr, ok := domain.AsDomainError(err); ok {
			res.Error = &Error{
				Code:    domainErr.Code,
				Message: domainErr.Message,
			}
		} else if rootErr, ok := rootDomain.AsDomainError(err); ok {
			res.Error = &Error{
				Code:    rootErr.Code,
				Message: rootErr.Message,
			}
		} else {
			res.Error = &Error{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			}
		}
	}

	_ = json.NewEncoder(w).Encode(res)
}

// SendError sends an error response for internal/server errors with 500 status code.
func (c *Controller) SendError(w http.ResponseWriter, r *http.Request, err error) {
	log := logger.LoggerFromContext(r.Context())
	if err != nil {
		log.WithError(err).Error("unhandled server error during request processing")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	msg := "internal error"
	code := "INTERNAL_ERROR"
	if err != nil {
		if domainErr, ok := domain.AsDomainError(err); ok {
			code = domainErr.Code
		} else if rootErr, ok := rootDomain.AsDomainError(err); ok {
			code = rootErr.Code
		}
	}

	_ = json.NewEncoder(w).Encode(ResponseError{
		Status:  "error",
		Message: msg,
		Code:    code,
	})
}
