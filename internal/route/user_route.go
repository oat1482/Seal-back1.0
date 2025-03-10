package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes เปลี่ยนให้รับ fiber.Router แทน *fiber.App
// และปรับเป็น /api/users ตามโครงสร้าง group
func SetupUserRoutes(router fiber.Router, userController *controller.UserController) {
	api := router.Group("/api")
	user := api.Group("/users")

	user.Get("/:username", userController.GetUserHandler)
	user.Post("/", userController.CreateUserHandler)
}
