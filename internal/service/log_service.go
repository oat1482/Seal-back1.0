package service

import (
	"errors"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
)

type LogService struct {
	repo *repository.LogRepository
}

func NewLogService(repo *repository.LogRepository) *LogService {
	return &LogService{repo: repo}
}

// บันทึก Log
func (s *LogService) CreateLog(userID uint, action string) error {
	if userID == 0 || action == "" {
		return errors.New("missing required fields")
	}
	log := model.Log{
		UserID: userID,
		Action: action,
	}
	return s.repo.Create(&log)
}

// ดึง Log ทั้งหมด
func (s *LogService) GetAllLogs() ([]model.Log, error) {
	return s.repo.GetAll()
}
