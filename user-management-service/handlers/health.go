package handlers

import (
	"fmt"
	"net/http"
)

func InitializeHealth() {
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
