package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var (
	data  = make(map[string]string)
	mutex = sync.Mutex{}
)

type RequestData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Hello from Service B"}
	json.NewEncoder(w).Encode(response)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var request RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	data[request.Key] = request.Value
	mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data added"})
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	var request RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	if _, exists := data[request.Key]; !exists {
		mutex.Unlock()
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	data[request.Key] = request.Value
	mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"message": "Data updated"})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	var request RequestData
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	if _, exists := data[request.Key]; !exists {
		mutex.Unlock()
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	delete(data, request.Key)
	mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"message": "Data deleted"})
}

func main() {
	http.HandleFunc("/hello", getHandler)
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postHandler(w, r)
		case http.MethodPut:
			putHandler(w, r)
		case http.MethodDelete:
			deleteHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Service B running on port 8000")
	http.ListenAndServe(":8000", nil)
}
