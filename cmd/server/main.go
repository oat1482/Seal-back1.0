package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kev2406/PEA/internal/config"
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
	migration "github.com/Kev2406/PEA/internal/infrastructure/database"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/Kev2406/PEA/internal/route"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var secretKey = []byte("your-secret-key")

func generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"emp_id":     user.EmpID,
		"role":       user.Role,
		"title":      user.Title,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"username":   user.Username,
		"email":      user.Email,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"pea_code":   user.PeaCode,
		"pea_short":  user.PeaShort,
		"pea_name":   user.PeaName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: No .env file found. Using system environment variables.")
	}

	// Initialize the database
	config.InitDB()

	log.Println("🔧 Running database migrations...")
	if err := migration.CreateStoreTable(config.DB); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Migrations completed!")

	// Create Fiber app
	app := fiber.New()

	// Apply authentication middleware globally (Every API requires JWT)
	app.Use(middleware.JWTMiddleware())

	// Initialize repositories
	userRepo := repository.NewUserRepository(config.DB)
	sealRepo := repository.NewSealRepository(config.DB)
	transactionRepo := repository.NewTransactionRepository(config.DB)
	logRepo := repository.NewLogRepository(config.DB)

	// Initialize services
	userService := service.NewUserService(userRepo)
	sealService := service.NewSealService(sealRepo, transactionRepo, logRepo, config.DB)
	logService := service.NewLogService(logRepo)

	// Admin user to be created or verified
	adminUser := &model.User{
		EmpID:     998877,
		Title:     "Mr.",
		FirstName: "Admin",
		LastName:  "Test",
		Username:  "admin_test",
		Email:     "admin_test@pea.co.th",
		Role:      "admin",
		PeaCode:   "F01101",
		PeaShort:  "FNRM",
		PeaName:   "กฟจ.นครราชสีมา",
	}

	// Check if admin already exists
	existingAdmin, _ := userService.GetUserByUsername(adminUser.Username)
	if existingAdmin == nil {
		if err := userService.CreateUser(adminUser); err != nil {
			log.Println("❌ Failed to create admin user:", err)
		} else {
			log.Println("✅ Admin user created!")
		}
	} else {
		adminUser = existingAdmin
		log.Println("🔹 Admin user already exists!")
	}

	// Normal user to be created or verified
	normalUser := &model.User{
		EmpID:     123456,
		Title:     "Mr.",
		FirstName: "User",
		LastName:  "Test",
		Username:  "user_test",
		Email:     "user_test@pea.co.th",
		Role:      "user",
		PeaCode:   "F02101",
		PeaShort:  "FCYP",
		PeaName:   "กฟจ.ชัยภูมิ",
	}

	// Check if normal user already exists
	existingUser, _ := userService.GetUserByUsername(normalUser.Username)
	if existingUser == nil {
		if err := userService.CreateUser(normalUser); err != nil {
			log.Println("❌ Failed to create normal user:", err)
		} else {
			log.Println("✅ Normal user created!")
		}
	} else {
		normalUser = existingUser
		log.Println("🔹 Normal user already exists!")
	}

	// Generate & log the JWT tokens for testing
	adminToken, _ := generateToken(adminUser)
	userToken, _ := generateToken(normalUser)
	log.Println("🛡️ Admin Token (ใช้ใน Postman):", adminToken)
	log.Println("👤 User Token (ใช้ใน Postman):", userToken)

	// Create controllers
	userController := controller.NewUserController(userService)
	sealController := controller.NewSealController(sealService)
	logController := controller.NewLogController(logService)

	// Setup routes (Protected by JWT)
	route.SetupUserRoutes(app, userController)
	route.SetupSealRoutes(app, sealController)

	// ✅ Log API is restricted to Admins only
	app.Use("/logs", middleware.AdminOnlyMiddleware)
	route.SetupLogRoutes(app, logController)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("🚀 Server is running on http://localhost:%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
