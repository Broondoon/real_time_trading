package models

type User struct {
	ID	uint `json:"id" gorm:"primaryKey"`
	Username	string `json:"username" gorm:"unique; not null"`
	Password string `json:"password" gorm:"not null"`
}

