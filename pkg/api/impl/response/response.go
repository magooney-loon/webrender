package response

import (
	"encoding/json"
	"net/http"

	types "github.com/magooney-loon/webserver/types/api"
)

type responder struct{}

// New creates a new responder
func New() types.Responder {
	return &responder{}
}

// JSON sends a JSON response
func (r *responder) JSON(w http.ResponseWriter, status int, data interface{}) {
	response := types.Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Error sends an error response
func (r *responder) Error(w http.ResponseWriter, err *types.Error) {
	response := types.Response{
		Success: false,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(response)
}

// Stream sends a streaming response
func (r *responder) Stream(w http.ResponseWriter, stream chan interface{}) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for data := range stream {
		response := types.Response{
			Success: true,
			Data:    data,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
		flusher.Flush()
	}
}
