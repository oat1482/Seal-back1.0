package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
)

// ✅ Secret Key สำหรับ Mock JWT (ควรใช้จาก ENV จริง)
var secretKey = []byte("your-secret-key")

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// ✅ VerifyPEAToken ตรวจสอบ Token ผ่าน PEA API หรือ Mock JWT
func (s *AuthService) VerifyPEAToken(tokenString string) (*model.User, error) {
	// ✅ 1️⃣ ลองถอดรหัส Mock JWT ก่อน
	user, err := s.VerifyMockJWT(tokenString)
	if err == nil {
		log.Println("✅ [VerifyPEAToken] ใช้ Mock JWT Token สำเร็จ:", user)
		return user, nil
	}
	log.Println("⚠️ [VerifyPEAToken] Mock JWT ไม่ถูกต้อง ลองตรวจสอบ PEA API...")

	// ✅ 2️⃣ ถ้า JWT ใช้ไม่ได้ ให้ลองตรวจสอบกับ PEA API
	url := "http://localhost:4000/mock-verify" // 👈 ใช้ Mock API แทนของจริงชั่วคราว
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("❌ [VerifyPEAToken] NewRequest Error:", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)
	log.Printf("🔑 [VerifyPEAToken] ส่ง Request ไปที่ PEA API: %s\n", url)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ [VerifyPEAToken] Request Error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// ✅ ตรวจสอบ HTTP Status Code
	if resp.StatusCode != http.StatusOK {
		log.Printf("🚨 [VerifyPEAToken] API ตอบกลับ: Status=%d\n", resp.StatusCode)
		return nil, errors.New("invalid token or unauthorized")
	}

	// ✅ Decode JSON Response
	var userFromAPI model.User
	if err := json.NewDecoder(resp.Body).Decode(&userFromAPI); err != nil {
		log.Println("❌ [VerifyPEAToken] Decode Error:", err)
		return nil, err
	}

	log.Printf("✅ [VerifyPEAToken] รับข้อมูล User จาก PEA API สำเร็จ: %+v\n", userFromAPI)
	return &userFromAPI, nil
}

// ✅ VerifyMockJWT - ถอดรหัส JWT ที่ Mock ขึ้นมา
func (s *AuthService) VerifyMockJWT(tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		log.Println("❌ [VerifyMockJWT] Invalid JWT:", err)
		return nil, errors.New("invalid token")
	}

	// ✅ ดึงค่า Claims ออกมา
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// ✅ ตรวจสอบว่า Claims มีค่าที่ต้องการครบไหม
		empID, empOk := claims["emp_id"].(float64)
		firstName, firstOk := claims["first_name"].(string)
		lastName, lastOk := claims["last_name"].(string)
		email, emailOk := claims["email"].(string)
		role, roleOk := claims["role"].(string)

		// ✅ เพิ่มฟิลด์ใหม่
		peaCode, peaCodeOk := claims["pea_code"].(string)
		peaShort, peaShortOk := claims["pea_short"].(string)
		peaName, peaNameOk := claims["pea_name"].(string)

		if !empOk || !firstOk || !lastOk || !emailOk || !roleOk || !peaCodeOk || !peaShortOk || !peaNameOk {
			log.Println("❌ [VerifyMockJWT] Claims ไม่ครบ:", claims)
			return nil, errors.New("invalid token claims")
		}

		// ✅ สร้าง User Object จาก JWT Claims
		user := &model.User{
			EmpID:     uint(empID),
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Role:      role,
			PeaCode:   peaCode,
			PeaShort:  peaShort,
			PeaName:   peaName,
		}

		log.Println("✅ [VerifyMockJWT] สำเร็จ:", user)
		return user, nil
	}

	return nil, errors.New("invalid token claims")
}
