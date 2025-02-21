package model

import (
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Action    string         `gorm:"not null" json:"action"`
	Timestamp time.Time      `gorm:"autoCreateTime" json:"timestamp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
