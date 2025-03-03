package route

import (
	"log"

	"github.com/Kev2406/PEA/internal/controller"
	"github.com/gofiber/fiber/v2"
)

func AdminOnlyMiddleware(c *fiber.Ctx) error {
	// Log path ของ request
	log.Printf("🔍 [AdminOnlyMiddleware] Request Path: %s", c.Path())

	// ดึง role จาก context
	role, ok := c.Locals("role").(string)
	log.Printf("🔍 [AdminOnlyMiddleware] Role in context: %v (ok: %v)", role, ok)

	if !ok || role != "admin" {
		log.Printf("🚫 [AdminOnlyMiddleware] Access denied. Required role: admin, Found: %v", role)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Admins only"})
	}

	log.Println("✅ [AdminOnlyMiddleware] Access granted for admin")
	return c.Next()
}

func SetupLogRoutes(app *fiber.App, logController *controller.LogController) {
	// สร้าง group /api ก่อน
	api := app.Group("/api")
	// จากนั้นสร้าง group /logs ภายใต้ /api → เส้นทางจริงคือ /api/logs
	logGroup := api.Group("/logs")
	log.Println("🔧 [SetupLogRoutes] Setting up /api/logs routes")

	// ใช้ middleware ที่จำกัดให้เฉพาะ admin
	logGroup.Use(AdminOnlyMiddleware)

	// กำหนด route GET /api/logs พร้อม log เพิ่มเติม
	logGroup.Get("/", func(c *fiber.Ctx) error {
		log.Printf("🔍 [SetupLogRoutes] GET /api/logs invoked")
		return logController.GetLogsHandler(c)
	})
}
