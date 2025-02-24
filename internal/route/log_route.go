package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/gofiber/fiber/v2"
)

func AdminOnlyMiddleware(c *fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Admins only"})
	}
	return c.Next()
}

func SetupLogRoutes(app *fiber.App, logController *controller.LogController) {
	logGroup := app.Group("/logs")

	logGroup.Use(AdminOnlyMiddleware)
	logGroup.Get("/", logController.GetLogsHandler)
}
