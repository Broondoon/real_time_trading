package models

type User struct {
	ID       string `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username string `json:"user_name" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
}
