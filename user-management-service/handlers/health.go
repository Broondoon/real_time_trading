package handlers

import (
	"fmt"
	"net/http"
)

// InitializeHealth sets up the health check endpoint
func InitializeHealth() {
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
