package service

import (
	"errors"
	"time"

	"fmt"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var technicianSecretKey = []byte("your-technician-secret-key")

// TechnicianService รับผิดชอบ business logic สำหรับการลงทะเบียนและล็อกอินของช่าง
type TechnicianService struct {
	repo *repository.TechnicianRepository
}

// NewTechnicianService สร้าง instance ของ TechnicianService
func NewTechnicianService(repo *repository.TechnicianRepository) *TechnicianService {
	return &TechnicianService{
		repo: repo,
	}
}

// Register สำหรับลงทะเบียนช่างใหม่
// Register สำหรับลงทะเบียนช่างใหม่
func (s *TechnicianService) Register(tech *model.Technician) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tech.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tech.Password = string(hashedPassword)
	tech.CreatedAt = time.Now()
	tech.UpdatedAt = time.Now()

	// 🔍 Debug Technician Data ก่อนบันทึก
	fmt.Println("🔍 Debug Technician Data:", tech)

	return s.repo.Create(tech)
}

// Login สำหรับช่าง โดยตรวจสอบ credentials และสร้าง JWT token
func (s *TechnicianService) Login(username, password string) (string, error) {
	tech, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(tech.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"tech_id":  tech.ID,
		"username": tech.Username,
		"role":     "technician",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	signedToken, err := token.SignedString(technicianSecretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
