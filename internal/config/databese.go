package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	migrations "github.com/Kev2406/PEA/internal/infrastructure/database"
	"gorm.io/driver/sqlserver" // ✅ ใช้เฉพาะ SQL Server
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("❌ Invalid database port: %v", err)
	}

	// ✅ ใช้ SQL Server เท่านั้น
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		port,
		os.Getenv("DB_NAME"),
	)

	// ✅ ปิด Foreign Key Constraint ตอน Migration
	DB, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatalf("❌ Failed to connect to SQL Server: %v", err)
	}

	fmt.Println("✅ Connected to SQL Server successfully!")

	// ✅ เรียกใช้งาน Migration ถ้ามี
	err = migrations.CreateStoreTable(DB)
	if err != nil {
		log.Fatalf("❌ Migration error: %v", err)
	}
}