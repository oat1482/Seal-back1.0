package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupTechnicianRoutes(router fiber.Router, techController *controller.TechnicianController) {
	// 🔹 Group สำหรับ Technician (ใช้ /api/technician)
	tech := router.Group("/api/technician")

	tech.Post("/register", techController.RegisterHandler)
	tech.Post("/login", techController.LoginHandler)

	// 🔹 Routes ที่ต้องการ JWT Middleware
	protectedTech := tech.Group("", middleware.TechnicianJWTMiddleware())
	protectedTech.Get("/seals", techController.GetAssignedSealsHandler)
	protectedTech.Put("/seals/install", techController.InstallSealHandler)
	protectedTech.Put("/seals/return/:seal_number", techController.ReturnSealHandler)
	protectedTech.Put("/update/:id", techController.UpdateTechnicianHandler)

	// ✅ Import Technicians (ไม่ต้องใช้ Token)
	tech.Post("/import", techController.ImportTechniciansHandler)

	// ✅ ดึงรายชื่อช่างทั้งหมด (เปิด Public)
	tech.Get("/list", techController.GetAllTechniciansHandler)
}
