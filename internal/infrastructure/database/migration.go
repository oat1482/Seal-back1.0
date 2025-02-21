package migration

import (
	"log"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

// CreateStoreTable runs database migrations and seeds initial data
func CreateStoreTable(db *gorm.DB) error {
	log.Println("🚀 Starting AutoMigrate for tables...")

	// ✅ ปิด foreign key constraints ชั่วคราวเพื่อป้องกันปัญหา constraints ที่ไม่มีอยู่จริง
	db.Config.DisableForeignKeyConstraintWhenMigrating = true

	// ✅ เพิ่ม Log เช็คว่ารันถึงแต่ละโมเดลหรือไม่
	log.Println("🔄 Migrating User Table...")
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Printf("❌ Failed to migrate User: %v", err)
		return err
	}
	log.Println("✅ User Table Migrated Successfully!")

	log.Println("🔄 Migrating Seal Table...")
	if err := db.AutoMigrate(&model.Seal{}); err != nil {
		log.Printf("❌ Failed to migrate Seal: %v", err)
		return err
	}
	log.Println("✅ Seal Table Migrated Successfully!")

	log.Println("🔄 Migrating Transaction Table...")
	if err := db.AutoMigrate(&model.Transaction{}); err != nil {
		log.Printf("❌ Failed to migrate Transaction: %v", err)
		return err
	}
	log.Println("✅ Transaction Table Migrated Successfully!")

	log.Println("🔄 Migrating Log Table...")
	if err := db.AutoMigrate(&model.Log{}); err != nil {
		log.Printf("❌ Failed to migrate Log: %v", err)
		return err
	}
	log.Println("✅ Log Table Migrated Successfully!")

	log.Println("✅ Migration successful!")

	// ✅ เปิด foreign key constraints กลับมา หลังจาก migration เสร็จสิ้น
	db.Config.DisableForeignKeyConstraintWhenMigrating = false

	// ✅ Seed users after migration
	err := SeedUsers(db)
	if err != nil {
		log.Printf("⚠️ Warning: Seeding failed: %v", err)
	} else {
		log.Println("✅ Seeding completed successfully!")
	}

	return nil
}
