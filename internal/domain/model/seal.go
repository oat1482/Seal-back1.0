package model

import (
	"time"

	"gorm.io/gorm"
)

type Seal struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SealNumber string         `gorm:"unique;not null" json:"seal_number"`
	Status     string         `gorm:"not null" json:"status"` // available, issued, used, returned
	IssuedBy   *uint          `json:"issued_by,omitempty"`    // ใครเป็นคนเบิกซิลให้
	IssuedTo   *uint          `json:"issued_to,omitempty"`    // ใครเป็นคนรับซิล
	ReturnedBy *uint          `json:"returned_by,omitempty"`  // ✅ ใครเป็นคนคืนซิล
	UsedBy     *uint          `json:"used_by,omitempty"`      // ใครเป็นคนใช้ซิล
	IssuedAt   *time.Time     `json:"issued_at,omitempty"`
	UsedAt     *time.Time     `json:"used_at,omitempty"`
	ReturnedAt *time.Time     `json:"returned_at,omitempty"` // ✅ เวลาเมื่อคืนซิล
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
