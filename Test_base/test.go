package main

import (
	"Shared/entities/entity"
	"Shared/entities/user"
	"fmt"
	"time"
)

func main() {
	// Create a new User
	u := user.NewUser(user.NewUserParams{
		NewEntityParams: entity.NewEntityParams{
			Id:           "1",
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		},
		Username: "test",
		Password: "password",
	})

	// Print the User
	fmt.Println(u.GetId())
	fmt.Println(u.GetDateCreated())
	fmt.Println(u.GetDateModified())
	fmt.Println(u.GetUsername())
	fmt.Println(u.GetPassword())
}
