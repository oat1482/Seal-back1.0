package model

import (
	"time"

	"gorm.io/gorm"
)

type Seal struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	SealNumber string `gorm:"unique;not null" json:"seal_number"`
	// เดิม: // available, issued, used, returned
	// ใหม่: // "เพิ่ม", "จ่าย", "ติดตั้งแล้ว", "ใช้งานแล้ว"
	Status          string         `gorm:"not null" json:"status"`
	IssuedBy        *uint          `json:"issued_by,omitempty"`
	IssuedTo        *uint          `json:"issued_to,omitempty"`
	ReturnedBy      *uint          `json:"returned_by,omitempty"`
	UsedBy          *uint          `json:"used_by,omitempty"`
	IssuedAt        *time.Time     `json:"issued_at,omitempty"`
	UsedAt          *time.Time     `json:"used_at,omitempty"`
	ReturnedAt      *time.Time     `json:"returned_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	InstalledSerial string         `json:"installed_serial,omitempty"`
	ReturnRemarks   string         `json:"return_remarks,omitempty"`
	EmployeeCode    string         `json:"employee_code,omitempty"` // รหัสพนักงาน
	IssueRemark     string         `json:"issue_remark,omitempty"`  // หมายเหตุตอนจ่าย
}
