package repository

import (
	"errors"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

type TechnicianRepository struct {
	db *gorm.DB
}

func NewTechnicianRepository(db *gorm.DB) *TechnicianRepository {
	return &TechnicianRepository{db: db}
}

func (r *TechnicianRepository) Create(tech *model.Technician) error {
	return r.db.Create(tech).Error
}

func (r *TechnicianRepository) FindByUsername(username string) (*model.Technician, error) {
	var tech model.Technician
	if err := r.db.Where("username = ?", username).First(&tech).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("technician not found")
		}
		return nil, err
	}
	return &tech, nil
}
func (r *TechnicianRepository) FindByID(techID uint) (*model.Technician, error) {
	var tech model.Technician
	if err := r.db.Where("id = ?", techID).First(&tech).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ไม่พบช่างในระบบ")
		}
		return nil, err
	}
	return &tech, nil
}
func (r *TechnicianRepository) FindSealByNumber(sealNumber string) (*model.Seal, error) {
	var seal model.Seal
	if err := r.db.Where("seal_number = ?", sealNumber).First(&seal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ไม่พบซีลในระบบ")
		}
		return nil, err
	}
	return &seal, nil
}

// ✅ อัปเดตข้อมูลซีล
func (r *TechnicianRepository) UpdateSeal(seal *model.Seal) error {
	return r.db.Save(seal).Error
}

// ✅ บันทึก Log
func (r *TechnicianRepository) CreateLog(log *model.Log) error {
	return r.db.Create(log).Error
}
func (r *TechnicianRepository) UpdateTechnician(tech *model.Technician) error {
	return r.db.Save(tech).Error
}
func (r *TechnicianRepository) FindByTechCode(code string) (*model.Technician, error) {
	var tech model.Technician
	if err := r.db.Where("technician_code = ?", code).First(&tech).Error; err != nil {
		return nil, err
	}
	return &tech, nil
}
func (r *TechnicianRepository) GetAllTechnicians() ([]model.Technician, error) {
	var technicians []model.Technician
	if err := r.db.Find(&technicians).Error; err != nil {
		return nil, err
	}
	return technicians, nil
}
