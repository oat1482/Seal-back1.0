package model

import "time"

type Technician struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	TechnicianCode string `gorm:"not null;unique" json:"technician_code"` // ✅ เพิ่มฟิลด์นี้
	Username       string `gorm:"unique;not null" json:"username"`
	Password       string `gorm:"not null" json:"-"` // เก็บเป็น hashed password
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `gorm:"unique;not null" json:"email"`
	ElectricCode   string `json:"electric_code"`                // รหัสของการไฟฟ้า (ถ้ามี)
	PhoneNumber    string `gorm:"not null" json:"phone_number"` // ✅ เพิ่มเบอร์โทร

	// --- เพิ่มฟิลด์ใหม่ตามที่ต้องการ ---
	CompanyName string `json:"company_name"` // ชื่อบริษัท
	Department  string `json:"department"`   // ชื่อหน่วยงาน

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
