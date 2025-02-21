package controller

import (
	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

// ✅ ค้นหาผู้ใช้ตาม username
func (uc *UserController) GetUserHandler(c *fiber.Ctx) error {
	username := c.Params("username")

	user, err := uc.userService.GetUserByUsername(username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// ✅ เพิ่มผู้ใช้ใหม่
func (uc *UserController) CreateUserHandler(c *fiber.Ctx) error {
	var request struct {
		EmpID     uint   `json:"emp_id"`
		Title     string `json:"title_s_desc"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		Role      string `json:"role"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	existingUser, _ := uc.userService.GetUserByEmpID(request.EmpID)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Employee ID already exists"})
	}

	// ✅ สร้าง User ใหม่
	user := model.User{
		EmpID:     request.EmpID,
		Title:     request.Title,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Username:  request.Username,
		Email:     request.Email,
		Role:      request.Role,
	}

	err := uc.userService.CreateUser(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created successfully", "user": user})
}
