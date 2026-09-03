// Package response provides standardized JSON envelopes for every API
// handler so success/error shapes stay consistent across the whole app.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-cp-boilerplate/internal/pkg/apperror"
)

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

type successBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type errorBody struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Success writes a 200 OK standardized success response.
func Success(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, successBody{Status: "success", Message: message, Data: data})
}

// Created writes a 201 Created standardized success response.
func Created(c *gin.Context, data any, message string) {
	c.JSON(http.StatusCreated, successBody{Status: "success", Message: message, Data: data})
}

// SuccessWithMeta writes a success response including pagination metadata.
func SuccessWithMeta(c *gin.Context, data any, message string, meta Meta) {
	c.JSON(http.StatusOK, successBody{Status: "success", Message: message, Data: data, Meta: &meta})
}

// Error writes a standardized error response with an explicit status code.
func Error(c *gin.Context, statusCode int, message string, errs map[string]string) {
	c.JSON(statusCode, errorBody{Status: "error", Message: message, Errors: errs})
}

// FromAppError translates a domain *apperror.AppError into a standardized
// error response, falling back to 500 for unrecognized errors.
func FromAppError(c *gin.Context, err error) {
	if ae, ok := err.(*apperror.AppError); ok {
		Error(c, ae.HTTPStatus, ae.Message, ae.Errors)
		return
	}
	Error(c, http.StatusInternalServerError, "Internal server error", nil)
}
