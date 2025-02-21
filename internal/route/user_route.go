package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App, userController *controller.UserController) {
	api := app.Group("/api") // ✅ แก้เป็น /api
	user := api.Group("/users")

	user.Get("/:username", userController.GetUserHandler)
	user.Post("/", userController.CreateUserHandler)
}
