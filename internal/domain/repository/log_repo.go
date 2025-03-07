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

// ✅ บันทึก Log
func (r *LogRepository) Create(log *model.Log) error {
	return r.db.Create(log).Error
}

// ✅ ดึง Log ทั้งหมด
func (r *LogRepository) GetAll() ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Find(&logs).Error
	return logs, err
}

// ✅ ดึง Logs พร้อมข้อมูลของ Users
func (r *LogRepository) GetLogsWithUsers() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			logs.id AS log_id, logs.user_id, 
			users.first_name, users.last_name, users.email, users.role, 
			logs.action, logs.timestamp
		FROM logs
		JOIN users ON logs.user_id = users.emp_id
		ORDER BY logs.timestamp DESC
	`

	rows, err := r.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var logID, userID int
		var firstName, lastName, email, role, action string
		var timestamp string // Change if your DB type differs

		err = rows.Scan(
			&logID, &userID,
			&firstName, &lastName, &email, &role,
			&action, &timestamp,
		)
		if err != nil {
			return nil, err
		}

		// ✅ ใช้ map แทน interface{} เพื่อให้ถูกต้อง
		logEntry := map[string]interface{}{
			"log_id":     logID,
			"user_id":    userID,
			"first_name": firstName,
			"last_name":  lastName,
			"email":      email,
			"role":       role,
			"action":     action,
			"timestamp":  timestamp,
		}

		results = append(results, logEntry)
	}

	return results, nil
}

// ✅ ดึง Log ตาม ID
func (r *LogRepository) GetByID(logID uint) (*model.Log, error) {
	var log model.Log
	err := r.db.Where("id = ?", logID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// ✅ ดึง Log ตามประเภท (Type)
func (r *LogRepository) GetByType(logType string) ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Where("action = ?", logType).Find(&logs).Error
	return logs, err
}

// ✅ ดึง Log ตาม User ID
func (r *LogRepository) GetByUser(userID uint) ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Where("user_id = ?", userID).Find(&logs).Error
	return logs, err
}

// ✅ ดึง Log ตามช่วงเวลา (Date Range)
func (r *LogRepository) GetByDateRange(startDate, endDate string) ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Find(&logs).Error
	return logs, err
}

// ✅ ลบ Log ตาม ID
func (r *LogRepository) Delete(logID uint) error {
	return r.db.Where("id = ?", logID).Delete(&model.Log{}).Error
}

// ✅ Fetch logs by action type
func (r *LogRepository) GetByAction(action string) ([]model.Log, error) {
	var logs []model.Log
	err := r.db.Where("action LIKE ?", "%"+action+"%").Find(&logs).Error
	return logs, err
}
