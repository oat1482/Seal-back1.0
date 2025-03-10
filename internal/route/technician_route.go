package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupTechnicianRoutes(app *fiber.App, techController *controller.TechnicianController) {
	// สร้างกลุ่มเส้นทางสำหรับ technician
	tech := app.Group("/api/technician")

	// ✅ Register & Login **ไม่ต้องใช้ Token**
	tech.Post("/register", techController.RegisterHandler)
	tech.Post("/login", techController.LoginHandler)

	// ✅ ใช้ Middleware เฉพาะเส้นทางที่ต้องใช้ Token
	protectedTech := tech.Group("", middleware.TechnicianJWTMiddleware())

	// ตัวอย่าง: ดูซีลที่ assign ให้ช่าง ต้องมี Token
	protectedTech.Get("/seals", techController.GetAssignedSealsHandler)
}
