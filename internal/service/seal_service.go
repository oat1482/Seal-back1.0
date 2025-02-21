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

		// ✅ Log: "User X created seal Y"
		logEntry := model.Log{
			UserID: userID,
			Action: fmt.Sprintf("Created seal %s", seal.SealNumber),
		}
		return s.logRepo.Create(&logEntry)
	})
}

// ✅ สร้างซิลหลายตัวพร้อมกัน (Bulk Insert)
func (s *SealService) GenerateAndCreateSeals(count int, userID uint) ([]model.Seal, error) {
	// ✅ ดึงเลขล่าสุดจากฐานข้อมูล
	latestSealNumber, err := s.GetLatestSealNumber()
	if err != nil {
		return nil, err
	}

	// ✅ สร้างเลขซิลชุดใหม่ (17 หลัก ไม่มีตัวอักษร)
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

		// ✅ Log การสร้างซิล
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

// ✅ ฟังก์ชัน GenerateNextSealNumbers หาเลขซิลล่าสุดแล้วรันต่อ
func GenerateNextSealNumbers(latest string, count int) ([]string, error) {
	// ✅ ถ้ายังไม่มีเลขซิลในระบบ ใช้เลขเริ่มต้น 17 หลัก
	if latest == "" {
		latest = "00000000000000001"
	}

	if len(latest) != 17 {
		return nil, errors.New("invalid seal number format")
	}

	// ✅ แปลง string เป็น int64 เพื่อรองรับเลข 17 หลัก
	lastInt, err := strconv.ParseInt(latest, 10, 64)
	if err != nil {
		return nil, errors.New("invalid seal number format")
	}

	sealNumbers := []string{}
	for i := 1; i <= count; i++ {
		newNum := lastInt + int64(i)
		sealNumbers = append(sealNumbers, fmt.Sprintf("%017d", newNum)) // ✅ ใช้ %017d เพื่อให้ได้ 17 หลักเสมอ
	}

	return sealNumbers, nil
}
