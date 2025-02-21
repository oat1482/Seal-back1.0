// 📂 mock/mock_server.go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/mock-verify", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id":    9,
			"emp_id":     123456,
			"first_name": "User",
			"last_name":  "Test",
			"role":       "user",
			"email":      "user_test@pea.co.th",
			"pea_code":   "F01101",
			"pea_short":  "FNRM",
			"pea_name":   "กฟจ.นครราชสีมา",
		})
	})

	log.Println("🚀 Mock API running on http://localhost:4000")
	log.Fatal(app.Listen(":4000"))
}
