package models

type User struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Username string `json:"user_name" gorm:"not null"`
	Password string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
}
