package service

import (
	"errors"
	"log"
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
func (s *TechnicianService) InstallSeal(sealNumber string, techID uint, serialNumber string) error {
	// ✅ ค้นหาซิลจากฐานข้อมูล
	seal, err := s.repo.FindSealByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบซีลในระบบ")
	}

	// ✅ ตรวจสอบว่าซิลถูกมอบหมายให้ช่างคนนี้หรือไม่
	if seal.AssignedToTechnician == nil || *seal.AssignedToTechnician != techID {
		return errors.New("คุณไม่มีสิทธิ์ติดตั้งซีลนี้")
	}

	// ✅ ตรวจสอบว่าสถานะของซีลเป็น "จ่าย" เท่านั้น
	if seal.Status != "จ่าย" {
		return errors.New("ซิลต้องอยู่ในสถานะ 'จ่าย' เท่านั้นจึงจะติดตั้งได้")
	}

	now := time.Now()
	seal.Status = "ติดตั้งแล้ว"
	seal.UsedBy = &techID
	seal.UsedAt = &now
	seal.InstalledSerial = serialNumber // ✅ บันทึก Serial Number

	// ✅ บันทึกข้อมูลลงฐานข้อมูล
	if err := s.repo.UpdateSeal(seal); err != nil {
		return err
	}

	// ✅ บันทึก Log
	logEntry := model.Log{
		UserID: techID,
		Action: fmt.Sprintf("ติดตั้งซีล %s (Serial: %s)", sealNumber, serialNumber),
	}
	return s.repo.CreateLog(&logEntry)
}

func (s *TechnicianService) ReturnSeal(sealNumber string, techID uint, remarks string) error {
	// ✅ ค้นหาซิลจากฐานข้อมูล
	seal, err := s.repo.FindSealByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบซีลในระบบ")
	}

	// ✅ ตรวจสอบว่าซิลถูกใช้โดยช่างคนนี้หรือไม่
	if seal.UsedBy == nil || *seal.UsedBy != techID {
		return errors.New("คุณไม่มีสิทธิ์คืนซีลนี้")
	}

	// ✅ ตรวจสอบว่าสถานะของซีลเป็น "ติดตั้งแล้ว" เท่านั้น
	if seal.Status != "ติดตั้งแล้ว" {
		return errors.New("ซิลต้องอยู่ในสถานะ 'ติดตั้งแล้ว' เท่านั้นจึงจะคืนได้")
	}

	now := time.Now()
	seal.Status = "ใช้งานแล้ว"
	seal.ReturnedBy = &techID
	seal.ReturnedAt = &now
	seal.ReturnRemarks = remarks // ✅ บันทึกหมายเหตุ

	// ✅ บันทึกข้อมูลลงฐานข้อมูล
	if err := s.repo.UpdateSeal(seal); err != nil {
		return err
	}

	// ✅ บันทึก Log
	logEntry := model.Log{
		UserID: techID,
		Action: fmt.Sprintf("คืนซีล %s (หมายเหตุ: %s)", sealNumber, remarks),
	}
	return s.repo.CreateLog(&logEntry)
}
func (s *TechnicianService) UpdateTechnician(techID uint, req struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	CompanyName string
	Department  string
}) error {
	log.Println("🔍 [SERVICE] Checking if technician exists: ID =", techID)

	tech, err := s.repo.FindByID(techID)
	if err != nil {
		log.Println("❌ [ERROR] Technician not found:", err)
		return err
	}

	log.Println("✅ [SERVICE] Found Technician:", tech)

	// อัปเดตข้อมูลใหม่
	tech.FirstName = req.FirstName
	tech.LastName = req.LastName
	tech.PhoneNumber = req.PhoneNumber
	tech.CompanyName = req.CompanyName
	tech.Department = req.Department

	log.Println("🛠️ [SERVICE] Updating Technician:", tech)

	err = s.repo.UpdateTechnician(tech)
	if err != nil {
		log.Println("❌ [ERROR] Database update failed:", err)
		return err
	}

	log.Println("✅ [SERVICE] Technician update success!")
	return nil
}

func (s *TechnicianService) GetAllTechnicians() ([]model.Technician, error) {
	return s.repo.GetAllTechnicians()
}

// func (s *TechnicianService) UpdateTechnician(techID uint, req map[string]interface{}) error {
//     technician, err := s.repo.FindByID(techID)
//     if err != nil {
//         return err
//     }

//     // อัปเดตข้อมูลที่ส่งเข้ามา
//     if req["first_name"] != nil {
//         technician.FirstName = req["first_name"].(string)
//     }
//     if req["last_name"] != nil {
//         technician.LastName = req["last_name"].(string)
//     }
//     if req["phone_number"] != nil {
//         technician.PhoneNumber = req["phone_number"].(string)
//     }
//     if req["company_name"] != nil {
//         technician.CompanyName = req["company_name"].(string)
//     }
//     if req["department"] != nil {
//         technician.Department = req["department"].(string)
//     }

//	    return s.repo.UpdateTechnician(technician)
//	}
func (s *TechnicianService) DeleteTechnician(techID uint) error {
	return s.repo.DeleteTechnician(techID)
}
