package controller

import (
	"strconv"
	"strings"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

type LogController struct {
	logService *service.LogService
}

func NewLogController(logService *service.LogService) *LogController {
	return &LogController{logService: logService}
}

func (lc *LogController) CreateLogHandler(c *fiber.Ctx) error {
	var request struct {
		UserID uint   `json:"user_id"`
		Action string `json:"action"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request",
			"message": err.Error(),
		})
	}

	err := lc.logService.CreateLog(request.UserID, request.Action)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create log",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Log created successfully",
	})
}

func (lc *LogController) GetAllLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetAllLogs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch logs",
			"message": err.Error(),
		})
	}

	groupedLogs := map[string][]model.Log{
		"created":  {},
		"issued":   {},
		"used":     {},
		"returned": {},
		"other":    {},
	}

	for _, log := range logs {
		switch {
		case contains(log.Action, "Created seal"):
			groupedLogs["created"] = append(groupedLogs["created"], log)
		case contains(log.Action, "Issued seal"):
			groupedLogs["issued"] = append(groupedLogs["issued"], log)
		case contains(log.Action, "Used seal"):
			groupedLogs["used"] = append(groupedLogs["used"], log)
		case contains(log.Action, "Returned seal"):
			groupedLogs["returned"] = append(groupedLogs["returned"], log)
		default:
			groupedLogs["other"] = append(groupedLogs["other"], log)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    groupedLogs,
	})
}

// ✅ Get log by ID
func (lc *LogController) GetLogByIDHandler(c *fiber.Ctx) error {
	logID, err := strconv.Atoi(c.Params("log_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid log ID",
			"message": err.Error(),
		})
	}

	log, err := lc.logService.GetLogByID(uint(logID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Log not found",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"log":     log,
	})
}

// ✅ Get logs by type
func (lc *LogController) GetLogsByTypeHandler(c *fiber.Ctx) error {
	logType := c.Params("log_type")
	logs, err := lc.logService.GetLogsByType(logType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch logs",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    logs,
	})
}

// ✅ Get logs by user
func (lc *LogController) GetLogsByUserHandler(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid user ID",
			"message": err.Error(),
		})
	}

	logs, err := lc.logService.GetLogsByUser(uint(userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch logs",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    logs,
	})
}

// ✅ Get logs by date range (Example: ?start=2025-03-01&end=2025-03-07)
func (lc *LogController) GetLogsByDateRangeHandler(c *fiber.Ctx) error {
	startDate := c.Query("start")
	endDate := c.Query("end")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Missing date parameters",
			"message": "Both start and end dates are required",
		})
	}

	logs, err := lc.logService.GetLogsByDateRange(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch logs",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    logs,
	})
}

// ✅ Delete log by ID
func (lc *LogController) DeleteLogHandler(c *fiber.Ctx) error {
	logID, err := strconv.Atoi(c.Params("log_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid log ID",
			"message": err.Error(),
		})
	}

	err = lc.logService.DeleteLog(uint(logID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to delete log",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Log deleted successfully",
	})
}

// ✅ Get logs where action contains "Created seal"
func (lc *LogController) GetCreatedLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetLogsByAction("Created seal")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch created logs",
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "logs": logs})
}

// ✅ Get logs where action contains "Issued seal"
func (lc *LogController) GetIssuedLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetLogsByAction("Issued seal")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch issued logs",
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "logs": logs})
}

// ✅ Get logs where action contains "Used seal"
func (lc *LogController) GetUsedLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetLogsByAction("Used seal")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch used logs",
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "logs": logs})
}

// ✅ Get logs where action contains "Returned seal"
func (lc *LogController) GetReturnedLogsHandler(c *fiber.Ctx) error {
	logs, err := lc.logService.GetLogsByAction("Returned seal")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch returned logs",
			"message": err.Error(),
		})
	}
	return c.JSON(fiber.Map{"success": true, "logs": logs})
}
