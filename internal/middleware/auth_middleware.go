package middleware

import (
	"log"
	"regexp"
	"strings"

	"fmt"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ✅ JWTMiddleware ใช้ตรวจสอบ Token (ทั้ง PEA API และ Mock JWT)
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Println("🚨 Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing token",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Println("🔑 Received Token:", tokenString)

		authService := service.NewAuthService()

		// ✅ ลองตรวจสอบกับ Mock JWT ก่อน
		user, err := authService.VerifyMockJWT(tokenString)
		if err == nil {
			log.Println("✅ [JWTMiddleware] Token Verified using Mock JWT")
			setUserContext(c, user)
			return c.Next()
		}

		// ❌ ถ้า Mock JWT ไม่ผ่าน ลองตรวจสอบกับ PEA API
		log.Println("⚠️ [JWTMiddleware] Mock JWT verification failed. Trying PEA API...")
		user, err = authService.VerifyPEAToken(tokenString)
		if err != nil {
			log.Println("❌ [JWTMiddleware] Token Verification Failed:", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// ✅ Ensure role is set in Context
		if user.Role == "" {
			log.Println("🚨 [JWTMiddleware] Role is missing from token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token: missing role"})
		}

		log.Printf("✅ [JWTMiddleware] User Verified: ID=%d, Role=%s, Name=%s %s, PEA Code=%s\n",
			user.EmpID, user.Role, user.FirstName, user.LastName, user.PeaCode)

		setUserContext(c, user)
		return c.Next()
	}
}

func AdminOnlyMiddleware(c *fiber.Ctx) error {
	log.Printf("Role from c.Locals('role'): %v", c.Locals("role"))
	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied. Admins only"})
	}
	return c.Next()
}

// ✅ ฟังก์ชันช่วยเซ็ตค่า User Context ใน Fiber
func setUserContext(c *fiber.Ctx, user *model.User) {
	c.Locals("user_id", user.EmpID)
	c.Locals("role", user.Role) // ✅ Role is always set now
	c.Locals("first_name", user.FirstName)
	c.Locals("last_name", user.LastName)

	// ✅ เพิ่มข้อมูลของ กฟฟ. ลงใน Context
	c.Locals("pea_code", user.PeaCode)
	c.Locals("pea_short", user.PeaShort)
	c.Locals("pea_name", user.PeaName)
}

func TechnicianJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("🔍 [TechnicianJWTMiddleware] Checking Path:", c.Path())

		// ✅ เพิ่ม Path ที่ต้องการ Skip JWT ตรวจสอบ
		skipPaths := []string{
			"/api/technician/register",
			"/api/technician/login",
			"/api/technician/import",
			"/api/technician/list",
		}

		// ✅ ใช้ Regular Expression เพื่อตรวจสอบ /api/technician/update/:id และ /api/technician/delete/:id
		regexPatterns := []string{
			`^/api/technician/update/\d+$`, // รองรับ /api/technician/update/9
			`^/api/technician/delete/\d+$`, // รองรับ /api/technician/delete/9
		}

		// ✅ ตรวจสอบ Skip Path แบบปกติ
		for _, path := range skipPaths {
			if c.Path() == path {
				fmt.Println("✅ [TechnicianJWTMiddleware] Skipping JWT check for:", c.Path())
				return c.Next()
			}
		}

		// ✅ ตรวจสอบ Path ด้วย Regular Expression
		for _, pattern := range regexPatterns {
			match, _ := regexp.MatchString(pattern, c.Path())
			if match {
				fmt.Println("✅ [TechnicianJWTMiddleware] Skipping JWT check for:", c.Path())
				return c.Next()
			}
		}

		// 🔑 ตรวจสอบ Authorization Header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
		}

		// เอา Token จริงๆ ออกมา
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		technicianSecretKey := []byte("your-technician-secret-key")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return technicianSecretKey, nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// ✅ ตรวจสอบ Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		role, _ := claims["role"].(string)
		if role != "technician" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied: not technician"})
		}

		techIDFloat, ok := claims["tech_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token: missing tech_id"})
		}

		techID := uint(techIDFloat)
		c.Locals("tech_id", techID)
		c.Locals("role", role)

		return c.Next()
	}
}
