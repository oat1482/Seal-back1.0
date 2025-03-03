package controller

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

type SealController struct {
	sealService *service.SealService
}

func NewSealController(sealService *service.SealService) *SealController {
	return &SealController{sealService: sealService}
}

// ScanSealHandler สแกนบาร์โค้ดเพื่อตรวจสอบซิลจากเลขบาร์โค้ด
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

// GetSealReportHandler ดึงรายงานสถานะซิลทั้งหมด
func (sc *SealController) GetSealReportHandler(c *fiber.Ctx) error {
	report, err := sc.sealService.GetSealReport()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate report"})
	}
	return c.JSON(report)
}

// GetSealHandler ดึงข้อมูลซิลตามหมายเลข
func (sc *SealController) GetSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	seal, err := sc.sealService.GetSealByNumber(sealNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Seal not found"})
	}
	return c.JSON(seal)
}

// IssueSealHandler ออกซิลให้พนักงาน
func (sc *SealController) IssueSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	if err := sc.sealService.IssueSeal(sealNumber, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seal issued successfully"})
}

// UseSealHandler ใช้ซิล
func (sc *SealController) UseSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	if err := sc.sealService.UseSeal(sealNumber, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seal used successfully"})
}

// ReturnSealHandler คืนซิล
func (sc *SealController) ReturnSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	if err := sc.sealService.ReturnSeal(sealNumber, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seal returned successfully"})
}

// GenerateSealsHandler สำหรับ Admin สร้างซิลจำนวนมากแบบ Bulk
// รับค่า SealNumber ที่ผู้ใช้กำหนด และ count ที่ต้องการสร้าง
// ถ้าหมายเลขที่กรอกเข้ามามีอยู่แล้ว ระบบจะใช้หมายเลขล่าสุดในฐานข้อมูลแทน
// ตัวอย่าง: ถ้ากรอก "00000000000000019" และ count=3 จะได้ "00000000000000020", "00000000000000021", "00000000000000022"
func (sc *SealController) GenerateSealsHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	role, roleOk := c.Locals("role").(string)
	if !ok || !roleOk || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied, admin only"})
	}

	var request struct {
		SealNumber string `json:"seal_number"`
		Count      int    `json:"count"`
	}
	if err := c.BodyParser(&request); err != nil || request.Count <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if request.SealNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seal number is required"})
	}

	// เรียกใช้ฟังก์ชัน GenerateAndCreateSealsFromNumber ที่ตรวจ duplicate key
	seals, err := sc.sealService.GenerateAndCreateSealsFromNumber(request.SealNumber, request.Count, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seals generated successfully", "seals": seals})
}

// CreateSealHandler สำหรับผู้ใช้ทั่วไป สร้างซิลโดยรับเลขซิลที่ผู้ใช้กรอกหรือจากการ Scan
// ระบบจะสร้างหมายเลขซิลต่อจากเลขที่กำหนด
// ตัวอย่าง: ถ้ากรอก "00000000000000019" และ count=3 จะได้ "00000000000000020", "00000000000000021", "00000000000000022"
func (sc *SealController) CreateSealHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var request struct {
		SealNumber string `json:"seal_number"`
		Count      int    `json:"count"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if request.SealNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seal number is required"})
	}
	if request.Count <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Count must be greater than zero"})
	}

	// เรียกใช้ฟังก์ชัน GenerateAndCreateSealsFromNumber ที่ตรวจ duplicate key
	seals, err := sc.sealService.GenerateAndCreateSealsFromNumber(request.SealNumber, request.Count, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seals created successfully", "seals": seals})
}

// ฟังก์ชัน incrementSealNumber เก็บไว้ในกรณีที่ต้องการใช้งานในอนาคต
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

// Add Music You Liked To Typing Musicddddddsfcvsdcds
