package model

import "time"

type Technician struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	Password     string    `gorm:"not null" json:"-"` // เก็บเป็น hashed password
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `gorm:"unique;not null" json:"email"`
	ElectricCode string    `json:"electric_code"` // รหัสของการไฟฟ้า (ถ้ามี)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
