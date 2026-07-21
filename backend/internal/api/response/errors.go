package response

import "net/http"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, Error{Code: "not_found", Message: msg})
}

func InternalError(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusInternalServerError, Error{Code: "internal_error", Message: msg})
}

func BadRequest(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusBadRequest, Error{Code: "bad_request", Message: msg})
}

func Unauthorized(w http.ResponseWriter) {
	JSON(w, http.StatusUnauthorized, Error{Code: "unauthorized", Message: "authentication required"})
}

func Forbidden(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusForbidden, Error{Code: "forbidden", Message: msg})
}

func UnprocessableEntity(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnprocessableEntity, Error{Code: "unsupported", Message: msg})
}

func Conflict(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusConflict, Error{Code: "conflict", Message: msg})
}
