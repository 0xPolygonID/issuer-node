package api

import "net/http"

// AuthError is a special error type used to signal an authorization error
type AuthError struct {
	err error
}

// Error satisfies error interface for AuthError
func (a AuthError) Error() string {
	return a.err.Error()
}

// RequestErrorHandlerFunc is a Request Error Handler that can be injected in oapi-codegen to handler errors in requests
func RequestErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

// ResponseErrorHandlerFunc is a Response Error Handler that can be injected in oapi-codegen to handler errors in requests
// We use it to create custom responses to some errors that may occur, like an authentication error.
func ResponseErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Add("Content-Type", "application/json")
	switch err.(type) {
	case AuthError:
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Add("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		_, _ = w.Write([]byte("\"Unauthorized\""))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
}
