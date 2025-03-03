package repository

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

type SealRepository struct {
	db *gorm.DB
}

func NewSealRepository(db *gorm.DB) *SealRepository {
	return &SealRepository{db: db}
}

func (r *SealRepository) Create(seal *model.Seal) error {
	return r.db.Create(seal).Error
}

func (r *SealRepository) CreateMultiple(seals []model.Seal) error {
	return r.db.Create(&seals).Error
}

func (r *SealRepository) FindByNumber(sealNumber string) (*model.Seal, error) {
	var seal model.Seal
	if err := r.db.Where("seal_number = ?", sealNumber).First(&seal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seal not found")
		}
		return nil, err
	}
	return &seal, nil
}

// ✅ ดึงเลขซิลล่าสุดจากฐานข้อมูล รองรับตัวอักษรนำหน้า และเลขไม่ฟิก 17 หลัก
func (r *SealRepository) GetLatestSeal() (*model.Seal, error) {
	var seals []model.Seal

	// ✅ คิวรีหาซิลทั้งหมดเรียงจากใหม่ -> เก่า
	if err := r.db.Order("seal_number DESC").Find(&seals).Error; err != nil {
		return nil, err
	}

	// ✅ ใช้ Regular Expression แยก Prefix และเลขท้าย
	re := regexp.MustCompile(`^([A-Za-z]*)(\d+)$`)

	var latestSeal *model.Seal
	var maxNumber int64

	// ✅ ค้นหาหมายเลขที่มีตัวเลขท้ายมากที่สุด
	for i := range seals {
		matches := re.FindStringSubmatch(seals[i].SealNumber)
		if len(matches) == 3 {
			num := parseInt64(matches[2])
			if num > maxNumber {
				maxNumber = num
				latestSeal = &seals[i]
			}
		}
	}

	if latestSeal == nil {
		return nil, nil // ✅ ถ้าไม่มีเรคคอร์ด ให้ return nil
	}
	return latestSeal, nil
}

// ✅ ค้นหาหมายเลขซิลล่าสุดที่มี Prefix เดียวกัน
func (r *SealRepository) FindByPrefix(prefix string) (*model.Seal, error) {
	var seal model.Seal
	if err := r.db.Where("seal_number LIKE ?", prefix+"%").Order("LENGTH(seal_number) DESC, seal_number DESC").First(&seal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &seal, nil
}

func (r *SealRepository) Update(seal *model.Seal) error {
	return r.db.Save(seal).Error
}

func (r *SealRepository) Delete(sealID uint) error {
	return r.db.Delete(&model.Seal{}, sealID).Error
}

// ✅ Helper: แปลง string เป็น int64 (กรณีที่มีเลขอยู่ใน Seal Number)
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
