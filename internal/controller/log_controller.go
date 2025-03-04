package controller

import (
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

type LogController struct {
	logService *service.LogService
}

func NewLogController(logService *service.LogService) *LogController {
	return &LogController{logService: logService}
}

func (lc *LogController) GetAllLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetAllLogs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch logs"})
	}

	return c.JSON(logs)
}

func (lc *LogController) GetLogsHandler(c *fiber.Ctx) error {

	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Admin only."})
	}

	logs, err := lc.logService.GetLogsWithUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch logs with users"})
	}

	return c.JSON(logs)
}
