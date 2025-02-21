package repository

import (
	"errors"

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

// ✅ ดึงเลขซิลล่าสุดจากฐานข้อมูล (แก้ไขให้รองรับเลข 17 หลัก)
func (r *SealRepository) GetLatestSeal() (*model.Seal, error) {
	var seal model.Seal

	// ✅ ใช้ `CAST(seal_number AS BIGINT)` เพื่อรองรับเลข 17 หลัก
	err := r.db.Order("CAST(seal_number AS BIGINT) DESC").First(&seal).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // ✅ ถ้าไม่มีเรคคอร์ด ให้ return nil
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
