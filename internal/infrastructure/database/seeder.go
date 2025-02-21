package migration

import (
	"log"

	"github.com/Kev2406/PEA/internal/domain/model"
	"gorm.io/gorm"
)

// SeedUsers creates default users for testing before integrating with PEA API
func SeedUsers(db *gorm.DB) error {
	users := []model.User{
		{
			EmpID:     498143,
			Title:     "นาย",
			FirstName: "สมชาย",
			LastName:  "ใจดี",
			Username:  "somchai.j",
			Email:     "somchai.j@pea.co.th",
			PeaCode:   "F01101",
			PeaShort:  "FNRM",
			PeaName:   "กฟจ.นครราชสีมา",
		},
		{
			EmpID:     500112,
			Title:     "นาง",
			FirstName: "สมหญิง",
			LastName:  "ใจดี",
			Username:  "somying.j",
			Email:     "somying.j@pea.co.th",
			PeaCode:   "F01201",
			PeaShort:  "FNOS",
			PeaName:   "กฟส.โนนสูง",
		},
	}

	log.Println("🌱 Seeding users...")

	for _, user := range users {
		log.Printf("🔍 Checking user: %s %s (EmpID: %d, PeaCode: %s, PeaShort: %s, PeaName: %s)",
			user.FirstName, user.LastName, user.EmpID, user.PeaCode, user.PeaShort, user.PeaName)

		result := db.Where("emp_id = ?", user.EmpID).FirstOrCreate(&user)

		if result.Error != nil {
			log.Printf("❌ Error seeding user %s %s: %v", user.FirstName, user.LastName, result.Error)
			return result.Error
		}

		if result.RowsAffected > 0 {
			log.Printf("✅ Created mock user: %s %s (%s)", user.FirstName, user.LastName, user.PeaName)
		} else {
			log.Printf("⚠️ User already exists: %s %s (%s)", user.FirstName, user.LastName, user.PeaName)
		}
	}

	log.Println("✅ User seeding completed successfully!")
	return nil
}
