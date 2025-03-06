package main

import (
	"fmt"
	"log"
	"os"
	"sync"
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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var secretKey = []byte("your-secret-key")

func generateToken(user *model.User, wg *sync.WaitGroup, tokenChan chan<- string) {
	defer wg.Done()
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
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		tokenChan <- ""
		return
	}
	tokenChan <- signedToken
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: No .env file found. Using system environment variables.")
	}

	config.InitDB()

	log.Println("🔧 Running database migrations...")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := migration.CreateStoreTable(config.DB); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		log.Println("✅ Migrations completed!")
	}()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://192.168.2.19:5173, https://192.168.2.19:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, Authorization, Accept",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
		MaxAge:           12 * 3600,
	}))

	app.Options("*", func(c *fiber.Ctx) error {
		if c.Get("Origin") != "" {
			c.Set("Access-Control-Allow-Origin", c.Get("Origin"))
		}
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		c.Set("Access-Control-Allow-Credentials", "true")
		return c.SendStatus(fiber.StatusOK)
	})

	app.Use(middleware.JWTMiddleware())

	userRepo := repository.NewUserRepository(config.DB)
	sealRepo := repository.NewSealRepository(config.DB)
	transactionRepo := repository.NewTransactionRepository(config.DB)
	logRepo := repository.NewLogRepository(config.DB)

	userService := service.NewUserService(userRepo)
	sealService := service.NewSealService(sealRepo, transactionRepo, logRepo, config.DB)
	logService := service.NewLogService(logRepo)

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

	adminChan := make(chan string, 1)
	userChan := make(chan string, 1)
	wg.Add(2)
	go generateToken(adminUser, &wg, adminChan)
	go generateToken(normalUser, &wg, userChan)

	wg.Wait()
	close(adminChan)
	close(userChan)

	adminToken := <-adminChan
	userToken := <-userChan

	log.Println("🛡️ Admin Token (ใช้ใน Postman):", adminToken)
	log.Println("👤 User Token (ใช้ใน Postman):", userToken)

	userController := controller.NewUserController(userService)
	sealController := controller.NewSealController(sealService)
	logController := controller.NewLogController(logService)

	route.SetupUserRoutes(app, userController)
	route.SetupSealRoutes(app, sealController)

	app.Use("/logs", middleware.AdminOnlyMiddleware)
	route.SetupLogRoutes(app, logController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("🚀 Server is running on http://localhost:%s\n", port)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
