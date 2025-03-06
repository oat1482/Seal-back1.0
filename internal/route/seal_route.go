package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupSealRoutes(app *fiber.App, sealController *controller.SealController) {
	api := app.Group("/api")
	seal := api.Group("/seals")

	// ✅ User & Admin สามารถสร้าง Seal ได้
	seal.Post("/", middleware.JWTMiddleware(), sealController.CreateSealHandler)

	// ✅ Admin เท่านั้น สามารถ Generate ซิลชุดใหญ่ได้ (แบบเดิม)
	seal.Post("/generate", middleware.JWTMiddleware(), sealController.GenerateSealsHandler)

	// ✅ ฟีเจอร์ใหม่: Generate ซิลหลายชุด (Batch) ในครั้งเดียว
	seal.Post("/generate-batches", middleware.JWTMiddleware(), sealController.GenerateSealsMultipleBatchesHandler)

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

	// ✅ ดึงซีลตาม ID และ Status (อันนี้เพิ่มเข้ามา)
	seal.Get("/:id/status/:status", middleware.JWTMiddleware(), sealController.GetSealByIDAndStatusHandler)
}
