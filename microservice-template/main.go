package main

import (
	"Shared/entities/entity"
	"Shared/entities/user"
	"fmt"
	"time"
	//"Shared/network"
)

func main() {
	fmt.Println("Hello, World!")

	u := user.New(user.NewUserParams{
		NewEntityParams: entity.NewEntityParams{
			ID:           "u1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Username: "test",
		Password: "password",
	})

	// Print the User
	fmt.Print("User: ")
	fmt.Println(u.GetId())
	fmt.Println(u.GetDateCreated())
	fmt.Println(u.GetDateModified())
	fmt.Println(u.GetUsername())
	fmt.Println(u.GetPassword())

}
