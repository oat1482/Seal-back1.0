package service

import (
	"errors"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// ✅ ค้นหาผู้ใช้ตาม emp_id
func (s *UserService) GetUserByEmpID(empID uint) (*model.User, error) {
	return s.userRepo.GetByEmpID(empID)
}

// ✅ ค้นหาผู้ใช้ตาม username
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// ✅ เพิ่มผู้ใช้ใหม่
func (s *UserService) CreateUser(user *model.User) error {
	if user.EmpID == 0 || user.Username == "" || user.Email == "" {
		return errors.New("missing required fields")
	}
	return s.userRepo.Create(user)
}
