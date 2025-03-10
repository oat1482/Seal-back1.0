package controller

import (
	"fmt"
	"log"

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

// ✅ RegisterHandler สำหรับลงทะเบียนช่างใหม่
func (tc *TechnicianController) RegisterHandler(c *fiber.Ctx) error {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		ElectricCode string `json:"electric_code"`
		PhoneNumber  string `json:"phone_number"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tech := &model.Technician{
		Username:     req.Username,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		ElectricCode: req.ElectricCode,
		PhoneNumber:  req.PhoneNumber,
	}

	if err := tc.technicianService.Register(tech); err != nil {
		log.Println("Registration error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Registration successful"})
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
