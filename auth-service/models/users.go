package models

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"user_name" gorm:"unique; not null"`
	Password string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
}
