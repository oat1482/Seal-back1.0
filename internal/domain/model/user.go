package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents an employee or admin in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"unique;not null" json:"username"`
	FullName  string         `gorm:"not null" json:"full_name"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Role      string         `gorm:"not null" json:"role"` // admin, employee
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
