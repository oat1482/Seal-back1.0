package service

import (
	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
)

// UserService บริการจัดการ User
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService สร้าง instance ของ UserService
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserByUsername ค้นหาผู้ใช้ตาม username
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// CreateUser เพิ่มผู้ใช้ใหม่
func (s *UserService) CreateUser(user *model.User) error {
	return s.userRepo.Create(user)
}
