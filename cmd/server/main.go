package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Kev2406/PEA/internal/config"
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/domain/repository"
	"github.com/Kev2406/PEA/internal/route"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// ✅ โหลดค่า Environment Variables จากไฟล์ .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: No .env file found. Using system environment variables.")
	}

	// ✅ เชื่อมต่อ Database
	config.InitDB()

	// ✅ สร้าง Fiber App
	app := fiber.New()

	// ✅ สร้าง Repository และ Service
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	// ✅ กำหนด Routes
	route.SetupUserRoutes(app, userController)

	// ✅ กำหนด PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // ค่าเริ่มต้น
	}

	// ✅ เริ่มต้นเซิร์ฟเวอร์
	fmt.Println("🚀 Server is running on http://localhost:" + port)
	log.Fatal(app.Listen(":" + port))
}
