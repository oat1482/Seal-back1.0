// user.model.go
package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents an employee or admin in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	EmpID     uint           `gorm:"uniqueIndex;not null" json:"emp_id"`
	Title     string         `gorm:"size:50;not null" json:"title_s_desc"`
	FirstName string         `gorm:"size:100;not null" json:"first_name"`
	LastName  string         `gorm:"size:100;not null" json:"last_name"`
	Username  string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Role      string         `gorm:"size:20;not null;default:'user'" json:"role"`
	PeaCode   string         `gorm:"size:10;not null" json:"pea_code"`  // รหัสกฟฟ.
	PeaShort  string         `gorm:"size:10;not null" json:"pea_short"` // ตัวย่อ
	PeaName   string         `gorm:"size:255;not null" json:"pea_name"` // ชื่อกฟฟ.
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
