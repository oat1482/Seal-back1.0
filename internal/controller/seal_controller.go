package controller

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

type SealController struct {
	sealService *service.SealService
}

func NewSealController(sealService *service.SealService) *SealController {
	return &SealController{sealService: sealService}
}

// ✅ Generate new seals in bulk (Admin only)
func (sc *SealController) GenerateSealsHandler(c *fiber.Ctx) error {
	// ✅ Extract user_id and role from authentication middleware
	userID, ok := c.Locals("user_id").(uint)
	role, roleOk := c.Locals("role").(string)

	if !ok || !roleOk || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied, admin only"})
	}

	// ✅ Extract count from request body
	var request struct {
		Count int `json:"count"`
	}
	if err := c.BodyParser(&request); err != nil || request.Count <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid count"})
	}

	// ✅ Call service to generate and create seals
	seals, err := sc.sealService.GenerateAndCreateSeals(request.Count, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seals generated successfully", "seals": seals})
}

func (sc *SealController) CreateSealHandler(c *fiber.Ctx) error {
	// ✅ Extract user_id (Allow User & Admin)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// ✅ หาเลข Seal ล่าสุดใน Database
	latestSealNumber, err := sc.sealService.GetLatestSealNumber()
	if err != nil {
		log.Println("❌ [CreateSealHandler] Error fetching latest seal:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot fetch latest seal"})
	}

	// ✅ ถ้ายังไม่มีเลข Seal ในระบบ ให้เริ่มจาก `16200000000000000`
	nextSealNumber := "16200000000000000"
	if latestSealNumber != "" {
		// ✅ สร้างเลขใหม่โดยบวกค่าเพิ่มทีละ 1
		nextSealNumber = incrementSealNumber(latestSealNumber)
	}

	// ✅ สร้าง Seal ใหม่
	newSeal := &model.Seal{
		SealNumber: nextSealNumber,
		Status:     "available",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = sc.sealService.CreateSeal(newSeal, userID)
	if err != nil {
		log.Println("❌ [CreateSealHandler] Failed to create seal:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seal created successfully", "seal": newSeal})
}

func incrementSealNumber(current string) string {
	if len(current) != 17 {
		log.Println("❌ Error: Invalid seal number format")
		return current
	}

	num, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		log.Println("❌ Error parsing seal number:", err)
		return current
	}

	num++                            // ✅ บวกค่าเพิ่ม 1
	return fmt.Sprintf("%017d", num) // ✅ ใช้ %017d เพื่อให้ได้ 17 หลักเสมอ
}
