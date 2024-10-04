package api

import (
	"encoding/json"
	"net/http"
)

type APIResponseError struct {
	// Human-defined error message
	Reason string `json:"reason,omitempty"`
	// Optional error directly from Go output
	Details string `json:"details,omitempty"`
}

func WriteResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	js, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(js)
}
