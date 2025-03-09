package handlers

import (
	"net/http"
)

func InitializeHealth() {
	http.HandleFunc("/health", healthHandler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//log.Println(w, "OK")
}
