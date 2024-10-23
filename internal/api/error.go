package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorHandlerFunc is an error adapter for the API. It is used to standardize errors happening in the generated code api handlers
func ErrorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	var invalidParamFormatError *InvalidParamFormatError
	switch {
	case errors.As(err, &invalidParamFormatError):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": err.Error()})
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
