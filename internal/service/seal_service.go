package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
	"gorm.io/gorm"
)

type SealService struct {
	repo            *repository.SealRepository
	transactionRepo *repository.TransactionRepository
	logRepo         *repository.LogRepository
	db              *gorm.DB
}

func NewSealService(repo *repository.SealRepository, transactionRepo *repository.TransactionRepository, logRepo *repository.LogRepository, db *gorm.DB) *SealService {
	return &SealService{
		repo:            repo,
		transactionRepo: transactionRepo,
		logRepo:         logRepo,
		db:              db,
	}
}

// ✅ ดึงเลขซิลล่าสุดจากฐานข้อมูล
func (s *SealService) GetLatestSealNumber() (string, error) {
	latestSeal, err := s.repo.GetLatestSeal()
	if err != nil {
		return "", err
	}
	if latestSeal == nil {
		return "00000000000000001", nil
	}
	return latestSeal.SealNumber, nil
}

// ✅ ค้นหาซิลตามหมายเลข
func (s *SealService) GetSealByNumber(sealNumber string) (*model.Seal, error) {
	return s.repo.FindByNumber(sealNumber)
}

// ✅ สร้างซิลใหม่
func (s *SealService) CreateSeal(seal *model.Seal, userID uint) error {
	existingSeal, _ := s.repo.FindByNumber(seal.SealNumber)
	if existingSeal != nil {
		return errors.New("seal number already exists")
	}

	now := time.Now()
	seal.Status = "available"
	seal.CreatedAt = now
	seal.UpdatedAt = now

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(seal); err != nil {
			return err
		}

		// ✅ เพิ่ม Log สำหรับการสร้างซิล
		logEntry := model.Log{
			UserID: userID,
			Action: fmt.Sprintf("Created seal %s", seal.SealNumber),
		}
		return s.logRepo.Create(&logEntry)
	})
}

// ✅ สร้างซิลหลายตัวพร้อมกัน (Bulk Insert) แบบ Auto-Generate จากเลขล่าสุด
func (s *SealService) GenerateAndCreateSeals(count int, userID uint) ([]model.Seal, error) {
	latestSealNumber, err := s.GetLatestSealNumber()
	if err != nil {
		return nil, err
	}

	sealNumbers, err := GenerateNextSealNumbers(latestSealNumber, count)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	seals := make([]model.Seal, count)

	for i, sn := range sealNumbers {
		seals[i] = model.Seal{
			SealNumber: sn,
			Status:     "available",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateMultiple(seals); err != nil {
			return err
		}

		// ✅ เพิ่ม Log สำหรับการสร้างซิลแบบ Bulk
		logEntry := model.Log{
			UserID: userID,
			Action: fmt.Sprintf("Generated %d seals", count),
		}
		return s.logRepo.Create(&logEntry)
	})

	if err != nil {
		return nil, err
	}

	return seals, nil
}

// ✅ สร้างซิลหลายตัวพร้อมกัน (Bulk Insert) โดยรับเลขเริ่มต้นจากผู้ใช้หรือจากการ Scan
func (s *SealService) GenerateAndCreateSealsFromNumber(startingSealNumber string, count int, userID uint) ([]model.Seal, error) {
	if len(startingSealNumber) != 17 {
		return nil, errors.New("invalid starting seal number format")
	}

	sealNumbers, err := GenerateNextSealNumbers(startingSealNumber, count)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	seals := make([]model.Seal, count)

	for i, sn := range sealNumbers {
		seals[i] = model.Seal{
			SealNumber: sn,
			Status:     "available",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateMultiple(seals); err != nil {
			return err
		}

		// ✅ เพิ่ม Log สำหรับการสร้างซิลแบบ Bulk จากเลขที่กำหนด
		logEntry := model.Log{
			UserID: userID,
			Action: fmt.Sprintf("Generated %d seals from starting number %s", count, startingSealNumber),
		}
		return s.logRepo.Create(&logEntry)
	})

	if err != nil {
		return nil, err
	}

	return seals, nil
}

// ✅ ออกซิลให้พนักงาน
func (s *SealService) IssueSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "issued", userID)
}

// ✅ ใช้ซิล
func (s *SealService) UseSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "used", userID)
}

// ✅ คืนซิล
func (s *SealService) ReturnSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "returned", userID)
}

// ✅ อัปเดตสถานะของซิล (Issue, Use, Return)
func (s *SealService) UpdateSealStatus(sealNumber string, status string, userID uint) error {
	seal, err := s.repo.FindByNumber(sealNumber)
	if err != nil {
		return errors.New("seal not found")
	}

	now := time.Now()
	logAction := ""

	switch status {
	case "issued":
		if seal.Status != "available" {
			return errors.New("only available seals can be issued")
		}
		seal.Status = "issued"
		seal.IssuedBy = &userID
		seal.IssuedAt = &now
		logAction = fmt.Sprintf("Issued seal %s", sealNumber)
	case "used":
		if seal.Status != "issued" {
			return errors.New("only issued seals can be used")
		}
		seal.Status = "used"
		seal.UsedBy = &userID
		seal.UsedAt = &now
		logAction = fmt.Sprintf("Used seal %s", sealNumber)
	case "returned":
		if seal.Status != "used" {
			return errors.New("only used seals can be returned")
		}
		seal.Status = "returned"
		seal.ReturnedBy = &userID
		seal.ReturnedAt = &now
		logAction = fmt.Sprintf("Returned seal %s", sealNumber)
	default:
		return errors.New("invalid status update")
	}

	// ✅ บันทึก Log และอัปเดตซิลใน transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(seal); err != nil {
			return err
		}

		// ✅ เพิ่ม Log สำหรับการเปลี่ยนแปลงสถานะ
		logEntry := model.Log{
			UserID: userID,
			Action: logAction,
		}
		return s.logRepo.Create(&logEntry)
	})
}

// ✅ รายงานสถานะของซิลทั้งหมด
func (s *SealService) GetSealReport() (map[string]interface{}, error) {
	var total, available, issued, used, returned int64

	if err := s.db.Model(&model.Seal{}).Count(&total).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "available").Count(&available).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "issued").Count(&issued).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "used").Count(&used).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "returned").Count(&returned).Error; err != nil {
		return nil, err
	}

	report := map[string]interface{}{
		"total_seals": total,
		"available":   available,
		"issued":      issued,
		"used":        used,
		"returned":    returned,
	}

	return report, nil
}

// ✅ ฟังก์ชัน GenerateNextSealNumbers หาเลขซิลล่าสุดหรือเลขเริ่มต้นแล้วรันต่อ
func GenerateNextSealNumbers(latest string, count int) ([]string, error) {
	if latest == "" {
		latest = "00000000000000001"
	}

	if len(latest) != 17 {
		return nil, errors.New("invalid seal number format")
	}

	lastInt, err := strconv.ParseInt(latest, 10, 64)
	if err != nil {
		return nil, errors.New("invalid seal number format")
	}

	sealNumbers := []string{}
	for i := 1; i <= count; i++ {
		newNum := lastInt + int64(i)
		sealNumbers = append(sealNumbers, fmt.Sprintf("%017d", newNum))
	}

	return sealNumbers, nil
}
