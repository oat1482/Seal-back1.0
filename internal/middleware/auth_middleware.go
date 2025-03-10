package middleware

import (
	"log"
	"strings"

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

// Middleware สำหรับ Technician
// ---------------------
// TechnicianJWTMiddleware ตรวจสอบ JWT สำหรับ Technician
func TechnicianJWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ** ถ้าเป็น Register หรือ Login ให้ข้ามการตรวจสอบ Token **
		if c.Path() == "/api/technician/register" || c.Path() == "/api/technician/login" {
			log.Println("🔑 Skipping TechnicianJWTMiddleware for:", c.Path())
			return c.Next()
		}

		log.Println("🔍 TechnicianJWTMiddleware is running for:", c.Path())

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Println("🚨 Missing Technician Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing token",
			})
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("🚨 Invalid token format, missing 'Bearer ' prefix")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		log.Println("🔑 Received Technician Token:", tokenString)

		technicianSecretKey := []byte("your-technician-secret-key")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return technicianSecretKey, nil
		})
		if err != nil || !token.Valid {
			log.Println("❌ Technician token invalid:", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("❌ Technician token claims invalid")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		role, ok := claims["role"].(string)
		if !ok || role != "technician" {
			log.Println("🚫 Access denied: not a technician")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied: not technician",
			})
		}

		techIDFloat, ok := claims["tech_id"].(float64)
		if !ok {
			log.Println("🚫 Technician token missing tech_id")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token: missing tech_id",
			})
		}

		log.Println("✅ Technician verified, ID:", uint(techIDFloat))

		c.Locals("tech_id", uint(techIDFloat))
		c.Locals("role", role)

		return c.Next()
	}
}
