package main

import (
	"Shared/network"
	"fmt"
	"log"
)

func main() {
	client := network.NewHttpClient("http://server:8000")
	client.SetAuthToken("my-secret-token-123")

	// GET
	resp, err := client.Get("/hello", nil)
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}
	fmt.Println("GET Response:", string(resp))

	// POST
	postPayload := map[string]string{"key": "username", "value": "ivan"}
	resp, err = client.Post("/data", postPayload)
	if err != nil {
		log.Fatalf("POST failed: %v", err)
	}
	fmt.Println("POST Response:", string(resp))

	// PUT
	putPayload := map[string]string{"key": "username", "value": "ivan-updated"}
	resp, err = client.Put("/data", putPayload)
	if err != nil {
		log.Fatalf("PUT failed: %v", err)
	}
	fmt.Println("PUT Response:", string(resp))

	//DELETE
	deletePayload := map[string]string{"key": "username"}
	resp, err = client.Delete("/data", deletePayload)
	if err != nil {
		log.Fatalf("DELETE failed: %v", err)
	}
	fmt.Println("DELETE Response:", string(resp))
}
