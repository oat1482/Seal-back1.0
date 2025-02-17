package repository

import (
	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

// UserRepository จัดการข้อมูลของ User
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository สร้าง instance ของ UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername ค้นหาผู้ใช้ตาม username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create เพิ่มผู้ใช้ใหม่
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update อัปเดตข้อมูลผู้ใช้
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}
