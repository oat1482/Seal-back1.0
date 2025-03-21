package controller

import (
	"fmt"
	"log"

	"strconv"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/Kev2406/PEA/internal/uploads"

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

func (tc *TechnicianController) InstallSealHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// รับค่า seal_number และ serial_number จาก JSON Body
	var req struct {
		SealNumber   string `json:"seal_number"`
		SerialNumber string `json:"serial_number,omitempty"`
	}
	if err := c.BodyParser(&req); err != nil {
		log.Println("❌ [ERROR] Failed to parse request body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	log.Println("🔍 [DEBUG] InstallSealHandler: seal_number =", req.SealNumber, ", serial_number =", req.SerialNumber)

	if req.SealNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seal number is required"})
	}

	err := tc.technicianService.InstallSeal(req.SealNumber, techID, req.SerialNumber)
	if err != nil {
		log.Println("❌ [ERROR] Install Seal Error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":       "ติดตั้งซีลสำเร็จ กรุณาอัปโหลดรูปภาพ",
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
	techIDStr := c.Params("id")
	techID, err := strconv.Atoi(techIDStr)
	if err != nil {
		log.Println("❌ [ERROR] Invalid Technician ID:", techIDStr, "Error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid technician id"})
	}

	var req struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
		CompanyName string `json:"company_name"`
		Department  string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Println("❌ [ERROR] Invalid JSON body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	log.Println("🔍 [DEBUG] Technician Update Request:", req)

	// Convert req to the expected format
	techData := struct {
		FirstName   string
		LastName    string
		PhoneNumber string
		CompanyName string
		Department  string
	}{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: req.PhoneNumber,
		CompanyName: req.CompanyName,
		Department:  req.Department,
	}

	err = tc.technicianService.UpdateTechnician(uint(techID), techData)
	if err != nil {
		log.Println("❌ [ERROR] Failed to update technician:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("✅ [SUCCESS] Technician updated successfully! ID:", techID)
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

func (tc *TechnicianController) DeleteTechnicianHandler(c *fiber.Ctx) error {
	techIDStr := c.Params("id")
	techID, err := strconv.Atoi(techIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid technician ID"})
	}

	err = tc.technicianService.DeleteTechnician(uint(techID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Technician deleted successfully"})
}
func (tc *TechnicianController) UploadSealImagesHandler(c *fiber.Ctx) error {
	techID, ok := c.Locals("tech_id").(uint)
	if !ok {
		log.Println("❌ [ERROR] Unauthorized access: No tech_id found")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	log.Println("🔍 [DEBUG] Technician ID:", techID)

	sealNumber := c.FormValue("seal_number")
	log.Println("📸 [DEBUG] Received seal_number =", sealNumber)

	if sealNumber == "" {
		log.Println("❌ [ERROR] Missing seal_number in request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Seal number is required"})
	}

	var imageURL1, imageURL2 string

	// ✅ Debug ว่ามีไฟล์ถูกอัปโหลดมาจริงไหม + เช็กขนาดไฟล์
	file1, err1 := c.FormFile("image1")
	if err1 == nil {
		log.Println("✅ [DEBUG] Image1 received:", file1.Filename)
		log.Println("📂 [DEBUG] Image1 size:", file1.Size, "bytes")
		log.Println("📂 [DEBUG] Image1 MIME type:", file1.Header.Get("Content-Type"))

		if file1.Size > 0 {
			imageURL1, err1 = uploads.SaveImage(file1)
			if err1 != nil {
				log.Println("❌ [ERROR] Failed to save Image1:", err1)
			}
		} else {
			log.Println("⚠️ [WARNING] Image1 is empty (size = 0)")
		}
	} else {
		log.Println("❌ [ERROR] Image1 FormFile error:", err1)
	}

	file2, err2 := c.FormFile("image2")
	if err2 == nil {
		log.Println("✅ [DEBUG] Image2 received:", file2.Filename)
		log.Println("📂 [DEBUG] Image2 size:", file2.Size, "bytes")
		log.Println("📂 [DEBUG] Image2 MIME type:", file2.Header.Get("Content-Type"))

		if file2.Size > 0 {
			imageURL2, err2 = uploads.SaveImage(file2)
			if err2 != nil {
				log.Println("❌ [ERROR] Failed to save Image2:", err2)
			}
		} else {
			log.Println("⚠️ [WARNING] Image2 is empty (size = 0)")
		}
	} else {
		log.Println("❌ [ERROR] Image2 FormFile error:", err2)
	}

	// ✅ เพิ่ม Debug ก่อนส่งไปยัง Service
	log.Println("📸 [DEBUG] Sending to service -> seal_number:", sealNumber, "image1:", imageURL1, "image2:", imageURL2)

	err := tc.technicianService.UploadSealImages(sealNumber, techID, imageURL1, imageURL2)
	if err != nil {
		log.Println("❌ [ERROR] UploadSealImages Error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	log.Println("✅ [SUCCESS] Image upload completed for seal:", sealNumber)
	return c.JSON(fiber.Map{
		"message":     "อัปโหลดรูปสำเร็จ",
		"seal_number": sealNumber,
		"image1":      imageURL1,
		"image2":      imageURL2,
	})
}
