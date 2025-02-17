package dto

// UserDTO ใช้สำหรับรับและส่งข้อมูลผู้ใช้
type UserDTO struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"` // admin, employee
}
