package service

import (
	"errors"
	"fmt"
	"regexp"
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

func NewSealService(
	repo *repository.SealRepository,
	transactionRepo *repository.TransactionRepository,
	logRepo *repository.LogRepository,
	db *gorm.DB,
) *SealService {
	return &SealService{
		repo:            repo,
		transactionRepo: transactionRepo,
		logRepo:         logRepo,
		db:              db,
	}
}

func (s *SealService) GetLatestSealNumber() (string, error) {
	latestSeal, err := s.repo.GetLatestSeal()
	if err != nil {
		return "", err
	}
	if latestSeal == nil {
		return "F000000000001", nil
	}
	return latestSeal.SealNumber, nil
}
func (s *SealService) GetSealByNumber(sealNumber string) (*model.Seal, error) {
	return s.repo.FindByNumber(sealNumber)
}

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

func (s *SealService) GenerateAndCreateSealsFromNumber(startingSealNumber string, count int, userID uint) ([]model.Seal, error) {

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

func (s *SealService) IssueSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "issued", userID)
}

func (s *SealService) UseSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "used", userID)
}

func (s *SealService) ReturnSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "returned", userID)
}

func (s *SealService) UpdateSealStatus(sealNumber string, status string, userID uint) error {
	seal, err := s.repo.FindByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบหมายเลขซิล")
	}

	now := time.Now()
	logAction := ""

	switch status {
	case "ออกให้แล้ว":
		if seal.Status != "พร้อมใช้งาน" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'พร้อมใช้งาน' เท่านั้นจึงจะออกให้ได้")
		}
		seal.Status = "เบิกแล้ว"
		seal.IssuedBy = &userID
		seal.IssuedAt = &now
		logAction = fmt.Sprintf("ออกซิลหมายเลข %s", sealNumber)

	case "จ่ายซิลแล้ว":
		if seal.Status != "เบิกแล้ว" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'เบิกแล้ว' เท่านั้นจึงจะจ่ายซิลได้")
		}
		seal.Status = "จ่ายซิลแล้ว"
		seal.UsedBy = &userID
		seal.UsedAt = &now
		logAction = fmt.Sprintf("จ่ายซิลหมายเลข %s", sealNumber)

	case "คืนแล้ว":
		if seal.Status != "จ่ายซิลแล้ว" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'จ่ายซิลแล้ว' เท่านั้นจึงจะคืนได้")
		}
		seal.Status = "คืนแล้ว"
		seal.ReturnedBy = &userID
		seal.ReturnedAt = &now
		logAction = fmt.Sprintf("คืนซิลหมายเลข %s", sealNumber)

	default:
		return errors.New("สถานะที่ระบุไม่ถูกต้อง")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Update(seal); err != nil {
			return err
		}
		logEntry := model.Log{
			UserID: userID,
			Action: logAction,
		}
		return s.logRepo.Create(&logEntry)
	})
}

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

func GenerateNextSealNumbers(latest string, count int) ([]string, error) {
	if latest == "" {
		latest = "F000000000001"
	}

	re := regexp.MustCompile(`^([A-Za-z]*)(\d+)$`)
	matches := re.FindStringSubmatch(latest)
	if len(matches) != 3 {
		return nil, errors.New("invalid seal number format")
	}

	prefix := matches[1]
	numberPart := matches[2]
	lastInt, err := strconv.ParseInt(numberPart, 10, 64)
	if err != nil {
		return nil, errors.New("invalid numeric part in seal number")
	}

	sealNumbers := make([]string, count)
	numberLength := len(numberPart)

	for i := 0; i < count; i++ {
		newNum := lastInt + int64(i+1)
		sealNumbers[i] = fmt.Sprintf("%s%0*d", prefix, numberLength, newNum)
	}

	return sealNumbers, nil
}
