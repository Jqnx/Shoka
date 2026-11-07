// Package response renders the JSON and file replies shared by every API
// handler, including the machine-readable error envelope clients branch on.
package response

import "net/http"

// Error is the error envelope every failing endpoint returns: a stable
// machine-readable Code plus a human-readable Message.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NotFound replies 404 with the caller-supplied message.
func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Error{Code: "not_found", Message: msg})
}

// InternalError replies 500. Callers should pass a message safe to expose.
func InternalError(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, Error{Code: "internal_error", Message: msg})
}

// BadRequest replies 400 for malformed or missing request input.
func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, Error{Code: "bad_request", Message: msg})
}

// Unauthorized replies 401 with a fixed message; there is no session.
func Unauthorized(w http.ResponseWriter) {
	JSON(w, http.StatusUnauthorized, Error{Code: "unauthorized", Message: "authentication required"})
}

// Forbidden replies 403 when the session exists but lacks the permission.
func Forbidden(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusForbidden, Error{Code: "forbidden", Message: msg})
}

// UnprocessableEntity replies 422 for input that parses but fails validation.
func UnprocessableEntity(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnprocessableEntity, Error{Code: "unsupported", Message: msg})
}

// Conflict replies 409 when the request collides with existing state.
func Conflict(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusConflict, Error{Code: "conflict", Message: msg})
}
