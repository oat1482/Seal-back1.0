package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// SetupSealRoutes เปลี่ยนให้รับ fiber.Router แทน *fiber.App เพื่อรองรับการใช้ group
// SetupSealRoutes เปลี่ยนให้รับ fiber.Router แทน *fiber.App เพื่อรองรับการใช้ group
func SetupSealRoutes(router fiber.Router, sealController *controller.SealController) {
	api := router.Group("/api")
	seal := api.Group("/seals")

	// ✅ User & Admin สามารถสร้าง Seal ได้
	seal.Post("/", middleware.JWTMiddleware(), sealController.CreateSealHandler)

	// ✅ Admin เท่านั้น สามารถ Generate ซิลชุดใหญ่ได้ (แบบเดิม)
	seal.Post("/generate", middleware.JWTMiddleware(), sealController.GenerateSealsHandler)

	// ✅ เพิ่ม API Assign Seal ให้ช่าง (เฉพาะพนักงานไฟฟ้า)
	// ✅ ใช้ PUT เพื่อให้ตรงกับ Postman
	seal.Put("/:seal_number/assign", middleware.JWTMiddleware(), sealController.AssignSealToTechnicianHandler)

	// ✅ สแกนบาร์โค้ดเพื่อดึงข้อมูลซิล
	seal.Post("/scan", middleware.JWTMiddleware(), sealController.ScanSealHandler)

	// ✅ รายงานสถานะซิล (Admin เท่านั้น)
	seal.Get("/report", middleware.JWTMiddleware(), sealController.GetSealReportHandler)

	// ✅ ทุกคนสามารถอ่านข้อมูลซิลได้
	seal.Get("/:seal_number", middleware.JWTMiddleware(), sealController.GetSealHandler)

	// ✅ Admin เท่านั้น สามารถออกซิลให้ User ได้
	seal.Put("/:seal_number/issue", middleware.JWTMiddleware(), sealController.IssueSealHandler)

	// ✅ User ใช้ซิลที่ออกให้แล้ว
	seal.Put("/:seal_number/use", middleware.JWTMiddleware(), sealController.UseSealHandler)

	// ✅ User คืนซิลหลังจากใช้งาน
	seal.Put("/:seal_number/return", middleware.JWTMiddleware(), sealController.ReturnSealHandler)

	// ✅ ดึงซีลตาม Status
	seal.Get("/status/:status", sealController.GetSealsByStatusHandler)

	// ✅ ดึงซีลตาม ID และ Status
	seal.Get("/:id/status/:status", middleware.JWTMiddleware(), sealController.GetSealByIDAndStatusHandler)

	// ✅ ตรวจสอบการมีอยู่ของ Seal ตามเลข
	seal.Get("/check/:seal_number", sealController.CheckSealExistsHandler)
	// ✅ ช่างติดตั้ง Seal (เฉพาะที่ได้รับมอบหมาย)
	//seal.Put("/:seal_number/install", middleware.TechnicianJWTMiddleware(), sealController.InstallSealHandler)

	// ✅ ช่างคืน Seal
	//seal.Put("/:seal_number/return", middleware.TechnicianJWTMiddleware(), sealController.ReturnSealHandler)

}
