package database

import (
	"log"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

// MigrateDB ทำการ migrate models ต่างๆ
func MigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.User{}, // ตัวอย่าง model User
		// เพิ่ม model อื่นๆ ตามที่ต้องการ
	)
	if err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migration completed successfully!")
}
