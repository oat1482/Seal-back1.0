package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupTechnicianRoutes(router fiber.Router, techController *controller.TechnicianController) {
	tech := router.Group("/api/technician")

	// ✅ Register & Login **ไม่ต้องใช้ Token**
	tech.Post("/register", techController.RegisterHandler)
	tech.Post("/login", techController.LoginHandler)

	// ✅ ใช้ Middleware เฉพาะเส้นทางที่ต้องใช้ Token
	protectedTech := tech.Group("", middleware.TechnicianJWTMiddleware())

	// ✅ ดึงซีลที่ถูก Assign ให้ช่าง
	protectedTech.Get("/seals", techController.GetAssignedSealsHandler)

	// ✅ **เพิ่ม API ติดตั้ง Seal**
	protectedTech.Put("/seals/install", techController.InstallSealHandler)

	// ✅ **เพิ่ม API คืน Seal**
	protectedTech.Put("/seals/return/:seal_number", techController.ReturnSealHandler)
}
