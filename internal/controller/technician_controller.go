package controller

import (
	"log"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

// TechnicianController รับผิดชอบ endpoint สำหรับช่าง (Technician)
type TechnicianController struct {
	techService *service.TechnicianService
	sealService *service.SealService
}

// NewTechnicianController สร้าง instance ของ TechnicianController
func NewTechnicianController(techService *service.TechnicianService, sealService *service.SealService) *TechnicianController {
	return &TechnicianController{
		techService: techService,
		sealService: sealService,
	}
}

// RegisterHandler สำหรับลงทะเบียนช่างใหม่
func (tc *TechnicianController) RegisterHandler(c *fiber.Ctx) error {
	var req struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Email        string `json:"email"`
		ElectricCode string `json:"electric_code"`
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
	}
	if err := tc.techService.Register(tech); err != nil {
		log.Println("Registration error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Registration successful"})
}

// LoginHandler สำหรับล็อกอินช่าง
func (tc *TechnicianController) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	token, err := tc.techService.Login(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"token": token})
}

// GetAssignedSealsHandler สำหรับดึงข้อมูล Seal ที่ถูก assign ให้ช่าง
func (tc *TechnicianController) GetAssignedSealsHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	seals, err := tc.sealService.GetSealsByTechnician(techID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(seals)
}
