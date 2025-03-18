package controller

import (
	"fmt"
	"log"

	"strconv"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

// TechnicianController รับผิดชอบ endpoint สำหรับช่าง (Technician)
type TechnicianController struct {
	technicianService *service.TechnicianService
	sealService       *service.SealService
}

// NewTechnicianController สร้าง instance ของ TechnicianController
func NewTechnicianController(technicianService *service.TechnicianService, sealService *service.SealService) *TechnicianController {
	return &TechnicianController{
		technicianService: technicianService,
		sealService:       sealService,
	}
}

func (tc *TechnicianController) RegisterHandler(c *fiber.Ctx) error {
	var req struct {
		TechnicianCode string `json:"technician_code"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		Email          string `json:"email"`
		PhoneNumber    string `json:"phone_number"`

		// เพิ่มฟิลด์ใหม่
		CompanyName string `json:"company_name"`
		Department  string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// สร้าง Model เพื่อส่งไป Service
	tech := &model.Technician{
		TechnicianCode: req.TechnicianCode,
		Username:       req.Username,
		Password:       req.Password,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		PhoneNumber:    req.PhoneNumber,

		// ใส่ค่านี้ด้วย
		CompanyName: req.CompanyName,
		Department:  req.Department,
	}

	// เรียก Service เพื่อ Register
	if err := tc.technicianService.Register(tech); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Technician registered successfully"})
}

// ✅ LoginHandler สำหรับล็อกอินช่าง
func (tc *TechnicianController) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	token, err := tc.technicianService.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}

// ✅ Technician ดึงรายการ Seal ที่ถูก Assign ให้ตัวเอง
func (tc *TechnicianController) GetAssignedSealsHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	fmt.Println("✅ Technician ID from Token:", techID) // Debug Log

	seals, err := tc.sealService.GetSealsByTechnician(techID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(seals)
}

// ✅ Technician ติดตั้งซีล (เฉพาะที่ได้รับมอบหมาย)
func (tc *TechnicianController) InstallSealHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req struct {
		SealNumber   string `json:"seal_number"`
		SerialNumber string `json:"serial_number,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	log.Println("🔧 Install Seal Request:", req.SealNumber, "by Technician ID:", techID)

	err := tc.technicianService.InstallSeal(req.SealNumber, techID, req.SerialNumber)
	if err != nil {
		log.Println("❌ Install Seal Error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":       "ติดตั้ง Seal เรียบร้อย",
		"seal_number":   req.SealNumber,
		"serial_number": req.SerialNumber,
	})
}

// ✅ Technician คืนซีลที่ติดตั้งแล้ว
func (tc *TechnicianController) ReturnSealHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	sealNumber := c.Params("seal_number")
	var req struct {
		Remarks string `json:"remarks"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := tc.technicianService.ReturnSeal(sealNumber, techID, req.Remarks)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "คืน Seal สำเร็จ",
		"seal_number": sealNumber,
		"remarks":     req.Remarks,
	})
}
func (tc *TechnicianController) UpdateTechnicianHandler(c *fiber.Ctx) error {
	// รับ technician_id จาก URL param
	techIDStr := c.Params("id") // เช่น /api/technician/update/:id
	techID, err := strconv.Atoi(techIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid technician id"})
	}

	var req struct {
		// อันไหนที่อนุญาตให้แก้
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
		CompanyName string `json:"company_name"`
		Department  string `json:"department"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	// ส่งไป Service
	err = tc.technicianService.UpdateTechnician(uint(techID), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Technician updated successfully"})
}
func (tc *TechnicianController) ImportTechniciansHandler(c *fiber.Ctx) error {
	var techList []model.Technician
	if err := c.BodyParser(&techList); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON array"})
	}

	for _, t := range techList {
		// ใส่ default password ถ้าจำเป็น
		if t.Password == "" {
			t.Password = "default123"
		}

		if err := tc.technicianService.Register(&t); err != nil {
			// ถ้า error อาจ return ทันทีหรือสะสม error ไว้
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"message": "Imported successfully", "count": len(techList)})
}
func (tc *TechnicianController) GetAllTechniciansHandler(c *fiber.Ctx) error {
	technicians, err := tc.technicianService.GetAllTechnicians()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch technicians"})
	}
	return c.JSON(technicians)
}
