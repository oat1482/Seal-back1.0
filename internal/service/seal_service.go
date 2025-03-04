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

// -------------------------------------------------------------------
//                            ส่วนนิ่งเดิม
// -------------------------------------------------------------------

// ดึงเลขซิลล่าสุดจากฐานข้อมูล
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

// ค้นหาซิลตามหมายเลข
func (s *SealService) GetSealByNumber(sealNumber string) (*model.Seal, error) {
	return s.repo.FindByNumber(sealNumber)
}

// สร้างซิลใหม่ (กรณีสร้างทีละตัว)
func (s *SealService) CreateSeal(seal *model.Seal, userID uint) error {
	existingSeal, _ := s.repo.FindByNumber(seal.SealNumber)
	if existingSeal != nil {
		return errors.New("มีซิลเบอร์นี้แล้ว")
	}

	now := time.Now()
	// สถานะเริ่มต้นเป็น "เพิ่ม"
	seal.Status = "พร้อมใช้งาน"
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

// สร้างซิลหลายตัว (Bulk Insert) จากเลขล่าสุด
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

// สร้างซิลหลายตัว (Bulk Insert) โดยรับเลขที่กำหนด
func (s *SealService) GenerateAndCreateSealsFromNumber(startingSealNumber string, count int, userID uint) ([]model.Seal, error) {
	if count == 1 {
		existingSeal, _ := s.repo.FindByNumber(startingSealNumber)
		if existingSeal != nil {
			// ถ้ามีซิลนี้แล้ว ไม่ต้องสร้างพร้อมใช้งาน
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

// -------------------------------------------------------------------
//        Mechanics เดิม: "IssueSeal / UseSeal / ReturnSeal"
// -------------------------------------------------------------------

// IssueSeal: เปลี่ยนจาก "เพิ่ม" → "จ่าย"
func (s *SealService) IssueSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "จ่าย", userID)
}

// UseSeal: เปลี่ยนจาก "จ่าย" → "ติดตั้งแล้ว"
func (s *SealService) UseSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "ติดตั้งแล้ว", userID)
}

// ReturnSeal: เปลี่ยนจาก "ติดตั้งแล้ว" → "ใช้งานแล้ว"
func (s *SealService) ReturnSeal(sealNumber string, userID uint) error {
	return s.UpdateSealStatus(sealNumber, "ใช้งานแล้ว", userID)
}

// ฟังก์ชัน UpdateSealStatus เดิม (ไม่เพิ่มพารามิเตอร์ใหม่)
func (s *SealService) UpdateSealStatus(sealNumber string, newStatus string, userID uint) error {
	seal, err := s.repo.FindByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบซิลในระบบ")
	}

	now := time.Now()
	logAction := ""

	switch newStatus {
	case "จ่าย":
		if seal.Status != "พร้อมใช้งาน" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'เพิ่ม' เท่านั้นจึงจะจ่ายได้")
		}
		seal.Status = "จ่าย"
		seal.IssuedBy = &userID
		seal.IssuedAt = &now
		logAction = fmt.Sprintf("จ่ายซิล %s", sealNumber)

	case "ติดตั้งแล้ว":
		if seal.Status != "จ่าย" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'จ่าย' เท่านั้นจึงจะติดตั้งได้")
		}
		seal.Status = "ติดตั้งแล้ว"
		seal.UsedBy = &userID
		seal.UsedAt = &now
		logAction = fmt.Sprintf("ติดตั้งซิล %s", sealNumber)

	case "ใช้งานแล้ว":
		if seal.Status != "ติดตั้งแล้ว" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'ติดตั้งแล้ว' เท่านั้นจึงจะใช้งานได้")
		}
		seal.Status = "ใช้งานแล้ว"
		seal.ReturnedBy = &userID
		seal.ReturnedAt = &now
		logAction = fmt.Sprintf("ซิล %s ถูกตั้งค่าว่าใช้งานแล้ว", sealNumber)

	default:
		return errors.New("สถานะไม่ถูกต้อง")
	}

	// บันทึก DB + Log
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

// -------------------------------------------------------------------
//       เมธอดใหม่สำหรับรองรับ SerialNumber / Remarks เพิ่มเติม
// -------------------------------------------------------------------

// UseSealWithSerial: ติดตั้งซิล พร้อมบันทึก serialNumber
func (s *SealService) UseSealWithSerial(sealNumber string, userID uint, deviceSerial string) error {
	return s.UpdateSealStatusWithExtra(sealNumber, "ติดตั้งแล้ว", userID, deviceSerial, "")
}

// ReturnSealWithRemarks: ส่งคืน (ใช้งานแล้ว) พร้อมบันทึก remarks
func (s *SealService) ReturnSealWithRemarks(sealNumber string, userID uint, remarks string) error {
	return s.UpdateSealStatusWithExtra(sealNumber, "ใช้งานแล้ว", userID, "", remarks)
}

// UpdateSealStatusWithExtra: version ใหม่ รองรับ deviceSerial, remarks
func (s *SealService) UpdateSealStatusWithExtra(
	sealNumber string,
	newStatus string,
	userID uint,
	deviceSerial string,
	remarks string,
) error {
	seal, err := s.repo.FindByNumber(sealNumber)
	if err != nil {
		return errors.New("ไม่พบซิลในระบบ")
	}

	now := time.Now()
	logAction := ""

	switch newStatus {
	case "ติดตั้งแล้ว":
		if seal.Status != "จ่าย" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'จ่าย' เท่านั้นจึงจะติดตั้งได้")
		}
		seal.Status = "ติดตั้งแล้ว"
		seal.UsedBy = &userID
		seal.UsedAt = &now

		// สมมติว่าคุณเพิ่มฟิลด์ `InstalledSerial` ใน model.Seal (ต้องแก้ model/Seal.go และ Migrate DB)
		// seal.InstalledSerial = deviceSerial

		logAction = fmt.Sprintf("ติดตั้งซิล %s (Serial: %s)", sealNumber, deviceSerial)

	case "ใช้งานแล้ว":
		if seal.Status != "ติดตั้งแล้ว" {
			return errors.New("ซิลต้องอยู่ในสถานะ 'ติดตั้งแล้ว' เท่านั้นจึงจะใช้งานได้")
		}
		seal.Status = "ใช้งานแล้ว"
		seal.ReturnedBy = &userID
		seal.ReturnedAt = &now

		// สมมติว่าคุณเพิ่มฟิลด์ `ReturnRemarks` ใน model.Seal หรือเก็บลง Transaction/Log
		// seal.ReturnRemarks = remarks

		logAction = fmt.Sprintf("ซิล %s ถูกตั้งค่าว่าใช้งานแล้ว (หมายเหตุ: %s)", sealNumber, remarks)

	default:
		return errors.New("สถานะไม่ถูกต้อง (version Extra)")
	}

	// บันทึก DB + Log
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

// -------------------------------------------------------------------
//                     ฟังก์ชัน GenerateNextSealNumbers
// -------------------------------------------------------------------

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

	// ไล่เลขต่อเนื่อง
	for i := 0; i < count; i++ {
		newNum := lastInt + int64(i)
		sealNumbers[i] = fmt.Sprintf("%s%0*d", prefix, numberLength, newNum)
	}

	return sealNumbers, nil
}

// -------------------------------------------------------------------
//           รายงานสถานะ (GetSealReport) เวอร์ชัน 4 ขั้น
// -------------------------------------------------------------------

func (s *SealService) GetSealReport() (map[string]interface{}, error) {
	var total, added, paid, installed, used int64

	if err := s.db.Model(&model.Seal{}).Where("status = ?", "เพิ่ม").Count(&added).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "จ่าย").Count(&paid).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "ติดตั้งแล้ว").Count(&installed).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Seal{}).Where("status = ?", "ใช้งานแล้ว").Count(&used).Error; err != nil {
		return nil, err
	}

	total = added + paid + installed + used
	report := map[string]interface{}{
		"total_seals": total,
		"เพิ่ม":       added,
		"จ่าย":        paid,
		"ติดตั้งแล้ว": installed,
		"ใช้งานแล้ว":  used,
	}
	return report, nil
}
