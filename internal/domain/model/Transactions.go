package model

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	SealID    uint           `gorm:"not null" json:"seal_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	IssuedTo  *uint          `json:"issued_to,omitempty"`    // ✅ เพิ่มฟิลด์นี้
	Action    string         `gorm:"not null" json:"action"` // issued, used, returned
	Timestamp time.Time      `gorm:"autoCreateTime" json:"timestamp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
