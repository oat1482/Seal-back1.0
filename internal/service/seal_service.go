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
	// ถ้ามีซิลเบอร์นี้อยู่แล้ว
	existingSeal, _ := s.repo.FindByNumber(seal.SealNumber)
	if existingSeal != nil {
		return errors.New("มีซิลเบอร์นี้แล้ว")
	}

	now := time.Now()
	seal.Status = "พร้อมใช้งาน" // เดิมคือ "available"
	seal.CreatedAt = now
	seal.UpdatedAt = now

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Create(seal); err != nil {
			return err
		}

		logEntry := model.Log{
			UserID: userID,
			Action: fmt.Sprintf("สร้างซิล %s", seal.SealNumber),
		}
		return s.logRepo.Create(&logEntry)
	})
}

// ✅ สร้างซิลหลายตัว (Bulk Insert) จากเลขล่าสุด
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
			Status:     "พร้อมใช้งาน",
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
			Action: fmt.Sprintf("สร้างซิลใหม่ %d อัน", count),
		}
		return s.logRepo.Create(&logEntry)
	})
	if err != nil {
		return nil, err
	}
	return seals, nil
}

// ✅ สร้างซิลหลายตัว (Bulk Insert) จากเลขที่กำหนด
func (s *SealService) GenerateAndCreateSealsFromNumber(startingSealNumber string, count int, userID uint) ([]model.Seal, error) {
	// ถ้า count == 1 -> เช็คว่ามีซิลนี้แล้วไหม
	if count == 1 {
		existingSeal, _ := s.repo.FindByNumber(startingSealNumber)
		if existingSeal != nil {
			return []model.Seal{*existingSeal}, nil
		}
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
			Status:     "พร้อมใช้งาน",
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
			Action: fmt.Sprintf("สร้างซิล %d อัน จากเลขเริ่ม %s", count, startingSealNumber),
		}
		return s.logRepo.Create(&logEntry)
	})
	if err != nil {
		return nil, err
	}
	return seals, nil
}

// ✅ ฟังก์ชันเปลี่ยนสถานะซิล
func (s *SealService) IssueSeal(sealNumber string, userID uint) error {
	// ส่ง "เบิก"
	return s.UpdateSealStatus(sealNumber, "เบิก", userID)
}
func (s *SealService) UseSeal(sealNumber string, userID uint) error {
	// ส่ง "จ่าย"
	return s.UpdateSealStatus(sealNumber, "จ่าย", userID)
}
func (s *SealService) ReturnSeal(sealNumber string, userID uint) error {
	// ส่ง "คืน"
	return s.UpdateSealStatus(sealNumber, "คืน", userID)
}

func (s *SealService) UpdateSealStatus(sealNumber string, status string, userID uint) error {
	seal, err := s.repo.FindByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบซิลในระบบ")
	}

	now := time.Now()
	logAction := ""

	switch status {
	case "เบิก":
		// เดิมคือ if seal.Status != "available"
		if seal.Status != "พร้อมใช้งาน" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'พร้อมใช้งาน' เท่านั้นจึงจะเบิกได้")
		}
		seal.Status = "เบิก"
		seal.IssuedBy = &userID
		seal.IssuedAt = &now
		logAction = fmt.Sprintf("เบิกซิล %s", sealNumber)

	case "จ่าย":
		if seal.Status != "เบิก" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'เบิก' เท่านั้นจึงจะจ่ายได้")
		}
		seal.Status = "จ่าย"
		seal.UsedBy = &userID
		seal.UsedAt = &now
		logAction = fmt.Sprintf("จ่ายซิล %s", sealNumber)

	case "คืน":
		if seal.Status != "จ่าย" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'จ่าย' เท่านั้นจึงจะคืนได้")
		}
		seal.Status = "คืน"
		seal.ReturnedBy = &userID
		seal.ReturnedAt = &now
		logAction = fmt.Sprintf("คืนซิล %s", sealNumber)

	default:
		return errors.New("สถานะไม่ถูกต้อง")
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

// ✅ รายงานสถานะ (นับจำนวน)
func (s *SealService) GetSealReport() (map[string]interface{}, error) {
	var total, available, issued, used, returned int64

	// นับซิลที่เป็น "พร้อมใช้งาน"
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "พร้อมใช้งาน").Count(&available).Error; err != nil {
		return nil, err
	}
	// นับซิลที่เป็น "เบิก"
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "เบิก").Count(&issued).Error; err != nil {
		return nil, err
	}
	// นับซิลที่เป็น "จ่าย"
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "จ่าย").Count(&used).Error; err != nil {
		return nil, err
	}
	// นับซิลที่เป็น "คืน"
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "คืน").Count(&returned).Error; err != nil {
		return nil, err
	}

	// รวมจำนวนซิลทั้งหมด = ผลรวมของทุกสถานะ
	total = available + issued + used + returned

	report := map[string]interface{}{
		"total_seals": total,
		"พร้อมใช้งาน": available,
		"เบิก":        issued,
		"จ่าย":        used,
		"คืน":         returned,
	}
	return report, nil
}

// ✅ GenerateNextSealNumbers: ใช้ Prefix + เลข, +i เพื่อไม่ให้ Count=1 ขยับเลข
func GenerateNextSealNumbers(latest string, count int) ([]string, error) {
	if latest == "" {
		latest = "F000000000001"
	}

	re := regexp.MustCompile(`^([A-Za-z]*)(\d+)$`)
	matches := re.FindStringSubmatch(latest)
	if len(matches) != 3 {
		return nil, errors.New("รูปแบบเลขซิลไม่ถูกต้อง")
	}

	prefix := matches[1]
	numberPart := matches[2]
	lastInt, err := strconv.ParseInt(numberPart, 10, 64)
	if err != nil {
		return nil, errors.New("เลขซิลไม่ถูกต้อง")
	}

	sealNumbers := make([]string, count)
	numberLength := len(numberPart)

	for i := 0; i < count; i++ {
		newNum := lastInt + int64(i)
		sealNumbers[i] = fmt.Sprintf("%s%0*d", prefix, numberLength, newNum)
	}

	return sealNumbers, nil
}
