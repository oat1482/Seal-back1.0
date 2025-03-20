package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupTechnicianRoutes(router fiber.Router, techController *controller.TechnicianController) {
	// 🔹 Group สำหรับ Technician (ใช้ /api/technician)
	tech := router.Group("/api/technician")

	// ✅ Public Routes (ไม่ต้องใช้ JWT)
	tech.Post("/register", techController.RegisterHandler)        // สมัครช่างใหม่
	tech.Post("/login", techController.LoginHandler)              // ล็อกอิน
	tech.Post("/import", techController.ImportTechniciansHandler) // Import รายชื่อช่าง
	tech.Get("/list", techController.GetAllTechniciansHandler)    // ดูรายชื่อช่างทั้งหมด

	tech.Put("/update/:id", techController.UpdateTechnicianHandler)    // อัปเดตข้อมูลช่าง
	tech.Delete("/delete/:id", techController.DeleteTechnicianHandler) // ลบข้อมูลช่าง

	// 🔹 Protected Routes (ต้องใช้ JWT)
	protectedTech := tech.Group("", middleware.TechnicianJWTMiddleware())

	// ✅ Routes ที่เกี่ยวกับการจัดการช่าง
	//protectedTech.Put("/update/:id", techController.UpdateTechnicianHandler)    // อัปเดตข้อมูลช่าง
	//protectedTech.Delete("/delete/:id", techController.DeleteTechnicianHandler) // ลบข้อมูลช่าง

	// ✅ Routes ที่เกี่ยวกับ Seal (เฉพาะช่างที่มีสิทธิ์)
	protectedTech.Get("/seals", techController.GetAssignedSealsHandler)               // ดูซีลที่ได้รับมอบหมาย
	protectedTech.Put("/seals/install", techController.InstallSealHandler)            // ติดตั้งซีล
	protectedTech.Put("/seals/return/:seal_number", techController.ReturnSealHandler) // คืนซีล

}
