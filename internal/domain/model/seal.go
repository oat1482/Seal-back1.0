package model

import (
	"time"

	"gorm.io/gorm"
)

type Seal struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	SealNumber           string         `gorm:"unique;not null" json:"seal_number"`
	Status               string         `gorm:"not null" json:"status"`
	IssuedBy             *uint          `json:"issued_by,omitempty"`
	IssuedTo             *uint          `json:"issued_to,omitempty"`
	ReturnedBy           *uint          `json:"returned_by,omitempty"`
	UsedBy               *uint          `json:"used_by,omitempty"`
	IssuedAt             *time.Time     `json:"issued_at,omitempty"`
	UsedAt               *time.Time     `json:"used_at,omitempty"`
	ReturnedAt           *time.Time     `json:"returned_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
	InstalledSerial      string         `json:"installed_serial,omitempty"`
	ReturnRemarks        string         `json:"return_remarks,omitempty"`
	EmployeeCode         string         `json:"employee_code,omitempty"`
	IssueRemark          string         `json:"issue_remark,omitempty"`
	AssignedToTechnician *uint          `json:"assigned_to_technician,omitempty"`

	// ✅ เพิ่มฟิลด์เก็บลิงก์รูปภาพ (อัปโหลด 2 รูป)
	Image1 string `json:"image1,omitempty"`
	Image2 string `json:"image2,omitempty"`
}
