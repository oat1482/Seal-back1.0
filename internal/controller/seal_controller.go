package controller

import (
	"fmt"
	"log"
	"regexp"
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

// --------------- ส่วนฟังก์ชันเดิม ๆ ที่ไม่เปลี่ยน --------------- //

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

func (sc *SealController) GetSealReportHandler(c *fiber.Ctx) error {
	report, err := sc.sealService.GetSealReport()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate report"})
	}
	return c.JSON(report)
}

func (sc *SealController) GetSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	seal, err := sc.sealService.GetSealByNumber(sealNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Seal not found"})
	}
	return c.JSON(seal)
}

// ------------------------------------------------------------------- //
//  1. “จ่าย Seal” (IssueSealHandler) ยังเหมือนเดิม แค่ปรับข้อความ   //
//
// ------------------------------------------------------------------- //
func (sc *SealController) IssueSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	if err := sc.sealService.IssueSeal(sealNumber, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	// ข้อความเป็นภาษาไทยตามสถานะใหม่
	return c.JSON(fiber.Map{"message": "จ่าย Seal เรียบร้อย"})
}

// ------------------------------------------------------------------- //
//  2. “ติดตั้ง (UseSeal)” + เพิ่ม Serial Number จาก Request Body     //
//
// ------------------------------------------------------------------- //
func (sc *SealController) UseSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เพิ่ม struct รับ serial_number (ถ้าไม่ส่งมา ก็จะเป็นค่าว่าง)
	var request struct {
		SerialNumber string `json:"serial_number,omitempty"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// เรียก Service โดยส่ง serialNumber ไปด้วย
	if err := sc.sealService.UseSealWithSerial(sealNumber, userID, request.SerialNumber); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"message":       "ติดตั้ง Seal เรียบร้อย",
		"serial_number": request.SerialNumber,
	})
}

// ------------------------------------------------------------------- //
//  3. “ใช้งานแล้ว (ReturnSeal)” + เพิ่ม Remarks / หมายเหตุ           //
//
// ------------------------------------------------------------------- //
func (sc *SealController) ReturnSealHandler(c *fiber.Ctx) error {
	sealNumber := c.Params("seal_number")
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เพิ่ม struct รับ remarks
	var request struct {
		Remarks string `json:"remarks,omitempty"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// เรียก Service โดยส่ง remarks ไปด้วย
	if err := sc.sealService.ReturnSealWithRemarks(sealNumber, userID, request.Remarks); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"message": "บันทึกเป็น 'ใช้งานแล้ว' เรียบร้อย",
		"remarks": request.Remarks,
	})
}

// ------------------------------------------------------------------- //
//
//	ส่วนการ Generate / Create Seal เดิม ๆ                             //
//
// ------------------------------------------------------------------- //
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

	seals, err := sc.sealService.GenerateAndCreateSealsFromNumber(request.SealNumber, request.Count, userID)
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

	seals, err := sc.sealService.GenerateAndCreateSealsFromNumber(request.SealNumber, request.Count, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Seals created successfully", "seals": seals})
}

// ส่วน incrementSealNumber เดิม
func incrementSealNumber(current string) string {
	if len(current) < 5 {
		log.Println("❌ Error: Invalid seal number format")
		return current
	}

	re := regexp.MustCompile(`^([A-Za-z]*)(\d+)$`)
	matches := re.FindStringSubmatch(current)
	if len(matches) != 3 {
		log.Println("❌ Error: Invalid seal number format")
		return current
	}

	prefix := matches[1]
	numberPart := matches[2]

	num, err := strconv.ParseInt(numberPart, 10, 64)
	if err != nil {
		log.Println("❌ Error parsing seal number:", err)
		return current
	}
	num++
	return fmt.Sprintf("%s%0*d", prefix, len(numberPart), num)
}
