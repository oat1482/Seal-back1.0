package repository

import (
	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

type LogRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(log *model.Log) error {
	return r.db.Create(log).Error
}

func (r *LogRepository) GetAll() ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Find(&logs).Error
	return logs, err
}
