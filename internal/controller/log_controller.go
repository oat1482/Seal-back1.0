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

// ดึง Logs ทั้งหมด
func (lc *LogController) GetAllLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetAllLogs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch logs"})
	}

	return c.JSON(logs)
}
