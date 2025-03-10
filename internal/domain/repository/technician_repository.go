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
