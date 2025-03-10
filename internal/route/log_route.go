package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// SetupLogRoutes เปลี่ยนให้รับ fiber.Router แทน *fiber.App เพื่อรองรับการใช้ group
func SetupLogRoutes(router fiber.Router, logController *controller.LogController) {
	api := router.Group("/api")
	logs := api.Group("/logs")

	// ✅ ดึง Log แยกตามประเภท (Created, Issued, Used, Returned)
	logs.Get("/created", middleware.JWTMiddleware(), logController.GetCreatedLogsHandler)
	logs.Get("/issued", middleware.JWTMiddleware(), logController.GetIssuedLogsHandler)
	logs.Get("/used", middleware.JWTMiddleware(), logController.GetUsedLogsHandler)
	logs.Get("/returned", middleware.JWTMiddleware(), logController.GetReturnedLogsHandler)

	// ✅ User & Admin สามารถสร้าง Log ได้
	logs.Post("/", middleware.JWTMiddleware(), logController.CreateLogHandler)

	// ✅ ดึง Log ทั้งหมด (Admin เท่านั้น)
	logs.Get("/", middleware.JWTMiddleware(), logController.GetAllLogsHandler)

	// ✅ ดึง Log ตามประเภท (Type)
	logs.Get("/type/:log_type", middleware.JWTMiddleware(), logController.GetLogsByTypeHandler)

	// ✅ ดึง Log ตามผู้ใช้งาน (User)
	logs.Get("/user/:user_id", middleware.JWTMiddleware(), logController.GetLogsByUserHandler)

	// ✅ ดึง Log ตามช่วงเวลา (Date Range)
	logs.Get("/range", middleware.JWTMiddleware(), logController.GetLogsByDateRangeHandler)

	// ✅ ดึง Log ตาม ID (Keep this LAST to avoid conflicts)
	logs.Get("/:log_id", middleware.JWTMiddleware(), logController.GetLogByIDHandler)

	// ✅ ลบ Log (Admin เท่านั้น)
	logs.Delete("/:log_id", middleware.JWTMiddleware(), logController.DeleteLogHandler)
}
