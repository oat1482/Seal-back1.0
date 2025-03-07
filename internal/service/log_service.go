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

// ✅ บันทึก Log
func (s *LogService) CreateLog(userID uint, action string) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	if action == "" {
		return errors.New("action is required")
	}

	log := &model.Log{
		UserID: userID,
		Action: action,
	}

	return s.repo.Create(log)
}

// ✅ ดึง Log ทั้งหมด
func (s *LogService) GetAllLogs() ([]model.Log, error) {
	return s.repo.GetAll()
}

// ✅ ดึง Logs พร้อมข้อมูลของ Users
func (s *LogService) GetLogsWithUsers() ([]map[string]interface{}, error) {
	return s.repo.GetLogsWithUsers()
}

// ✅ ดึง Log ตาม ID
func (s *LogService) GetLogByID(logID uint) (*model.Log, error) {
	if logID == 0 {
		return nil, errors.New("logID is required")
	}
	return s.repo.GetByID(logID)
}

// ✅ ดึง Log ตามประเภท (Type)
func (s *LogService) GetLogsByType(logType string) ([]model.Log, error) {
	if logType == "" {
		return nil, errors.New("log type is required")
	}
	return s.repo.GetByType(logType)
}

// ✅ ดึง Log ตาม User ID
func (s *LogService) GetLogsByUser(userID uint) ([]model.Log, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return s.repo.GetByUser(userID)
}

// ✅ ดึง Log ตามช่วงเวลา (Date Range)
func (s *LogService) GetLogsByDateRange(startDate, endDate string) ([]model.Log, error) {
	if startDate == "" || endDate == "" {
		return nil, errors.New("startDate and endDate are required")
	}
	return s.repo.GetByDateRange(startDate, endDate)
}

// ✅ ลบ Log ตาม ID
func (s *LogService) DeleteLog(logID uint) error {
	if logID == 0 {
		return errors.New("logID is required")
	}
	return s.repo.Delete(logID)
}

// ✅ Get logs by specific action type
func (s *LogService) GetLogsByAction(action string) ([]model.Log, error) {
	return s.repo.GetByAction(action)
}
