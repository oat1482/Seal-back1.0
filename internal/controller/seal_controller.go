package controller

import (
	"fmt"
	"log"
	"net/url"
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

// เราต้อง decode ค่าที่รับมาจาก URL ก่อนที่จะนำไปค้นหาในฐานข้อมูล เพราะตอนนี้ค่า status ที่ได้รับมานั้นถูก encode อยู่เป็น %E0%B8%9E... แทนที่จะเป็น "พร้อมใช้งาน" แบบปกติ

// ลองแก้ไขฟังก์ชัน GetSealsByStatusHandler ให้ decode โดยใช้ url.QueryUnescape
// ในไฟล์ seal_controller.go
func (sc *SealController) GetSealsByStatusHandler(c *fiber.Ctx) error {
	// ดึง status จาก URL params เช่น /api/seals/status/พร้อมใช้งาน
	rawStatus := c.Params("status")
	status, err := url.QueryUnescape(rawStatus)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status parameter: " + err.Error(),
		})
	}

	seals, err := sc.sealService.GetSealsByStatus(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(seals)
}

// ------------------------- ฟีเจอร์ใหม่ ------------------------- //
//
// POST /api/seals/generate-batches
//
// โครงสร้าง JSON ที่คาดหวัง:
// {
//   "batches": [
//     { "seal_number": "F2499", "count": 3 },
//     { "seal_number": "PEA000002", "count": 2 }
//   ]
// }
//
// เฉพาะ admin เท่านั้น
//
// -------------------------------------------------------------- //

func (sc *SealController) GenerateSealsMultipleBatchesHandler(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	role, roleOk := c.Locals("role").(string)
	if !ok || !roleOk || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied, admin only",
		})
	}

	// โครงสร้างสำหรับรับ request ที่มีหลาย batch
	var request struct {
		Batches []struct {
			SealNumber string `json:"seal_number"`
			Count      int    `json:"count"`
		} `json:"batches"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// เช็คว่าใน batches ต้องมีอย่างน้อย 1 รายการ
	if len(request.Batches) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No batches provided",
		})
	}

	// เตรียม slice สำหรับรวมผลลัพธ์ทั้งหมด
	var allCreatedSeals []interface{}

	// วนลูปในแต่ละ batch
	for _, batch := range request.Batches {
		if batch.SealNumber == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Seal number is required in each batch",
			})
		}
		if batch.Count <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid count (%d) in batch for seal_number=%s", batch.Count, batch.SealNumber),
			})
		}

		// เรียก service เพื่อ generate & create seals
		seals, err := sc.sealService.GenerateAndCreateSealsFromNumber(batch.SealNumber, batch.Count, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		// เก็บ seals ที่สร้างได้ในผลลัพธ์รวม
		allCreatedSeals = append(allCreatedSeals, seals)
	}

	// ตอบกลับเป็น JSON รวมทั้งหมด
	return c.JSON(fiber.Map{
		"message": "All batches generated successfully",
		"results": allCreatedSeals, // จะเป็น array ของ array seals (ถ้าอยากแบนให้อยู่ใน array เดียว อาจ loop รวมกันได้)
	})
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

	var request struct {
		IssuedTo     uint   `json:"issued_to"`     // รหัสพนักงานที่รับซิล
		EmployeeCode string `json:"employee_code"` // รหัสพนักงาน
		Remark       string `json:"remark"`        // หมายเหตุเพิ่มเติม
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if request.IssuedTo == 0 || request.EmployeeCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing required fields"})
	}

	if err := sc.sealService.IssueSealWithDetails(sealNumber, request.IssuedTo, request.EmployeeCode, request.Remark); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":       "จ่าย Seal เรียบร้อย",
		"seal_number":   sealNumber,
		"issued_to":     request.IssuedTo,
		"employee_code": request.EmployeeCode,
		"remark":        request.Remark,
	})
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

	var request struct {
		SerialNumber string `json:"serial_number,omitempty"` // ✅ รับ Serial Number
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ✅ ส่ง Serial Number ไปยัง Service
	if err := sc.sealService.UseSealWithSerial(sealNumber, userID, request.SerialNumber); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":       "ติดตั้ง Seal เรียบร้อย",
		"serial_number": request.SerialNumber, // ✅ Return Serial Number ที่รับมา
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

	var request struct {
		Remarks string `json:"remarks,omitempty"` // ✅ รับ Remarks
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ✅ ส่ง Remarks ไปยัง Service
	if err := sc.sealService.ReturnSealWithRemarks(sealNumber, userID, request.Remarks); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "บันทึกเป็น 'ใช้งานแล้ว' เรียบร้อย",
		"remarks": request.Remarks, // ✅ Return Remarks ที่รับมา
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
