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

// ✅ สแกนบาร์โค้ด (ตรวจสอบซิลจากเลขบาร์โค้ด)
func (sc *SealController) ScanSealHandler(c *fiber.Ctx) error {
	var request struct {
		SealNumber string `json:"seal_number"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	seal, err := sc.sealService.GetSealByNumber(request.SealNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Seal not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Seal scanned successfully",
		"seal":    seal,
	})
}

// ✅ ดึงข้อมูลซิลทั้งหมด
func (sc *SealController) GetSealReportHandler(c *fiber.Ctx) error {
	report, err := sc.sealService.GetSealReport()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate report"})
	}
	return c.JSON(report)
}

// ✅ ดึงข้อมูลซิลตามหมายเลข
func (sc *SealController) GetSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")

	seal, err := sc.sealService.GetSealByNumber(sealNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Seal not found"})
	}

	return c.JSON(seal)
}

// ✅ ออกซิลให้พนักงาน
func (sc *SealController) IssueSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	err := sc.sealService.IssueSeal(sealNumber, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seal issued successfully"})
}

// ✅ ใช้ซิล
func (sc *SealController) UseSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	err := sc.sealService.UseSeal(sealNumber, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seal used successfully"})
}

// ✅ คืนซิล
func (sc *SealController) ReturnSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	err := sc.sealService.ReturnSeal(sealNumber, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seal returned successfully"})
}

// ✅ Generate new seals in bulk (Admin only)
func (sc *SealController) GenerateSealsHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	role, roleOk := c.Locals("role").(string)

	if !ok || !roleOk || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied, admin only"})
	}

	var request struct {
		Count int `json:"count"`
	}
	if err := c.BodyParser(&request); err != nil || request.Count <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid count"})
	}

	seals, err := sc.sealService.GenerateAndCreateSeals(request.Count, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Seals generated successfully", "seals": seals})
}

func (sc *SealController) CreateSealHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	latestSealNumber, err := sc.sealService.GetLatestSealNumber()
	if err != nil {
		log.Println("❌ [CreateSealHandler] Error fetching latest seal:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot fetch latest seal"})
	}

	nextSealNumber := "16200000000000000"
	if latestSealNumber != "" {
		nextSealNumber = incrementSealNumber(latestSealNumber)
	}

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

	num++
	return fmt.Sprintf("%017d", num)
}
