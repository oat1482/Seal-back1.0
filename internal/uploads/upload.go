package uploads

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ✅ เพิ่มตัวแปร AllowedExtensions เพื่อกำหนดประเภทไฟล์ที่อนุญาต
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// SaveImage บันทึกไฟล์รูปภาพและคืนค่าลิงก์ไฟล์
func SaveImage(file *multipart.FileHeader) (string, error) {
	savePath := "internal/uploads/"

	// สร้างโฟลเดอร์ถ้ายังไม่มี
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
			return "", fmt.Errorf("ไม่สามารถสร้างโฟลเดอร์อัปโหลด: %v", err)
		}
	}

	// ✅ ตรวจสอบประเภทไฟล์ (ไฟล์ที่ไม่อยู่ใน AllowedExtensions จะถูกปฏิเสธ)
	extension := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedExtensions[extension] {
		return "", errors.New("ประเภทไฟล์ไม่รองรับ (อนุญาตเฉพาะ .jpg, .jpeg, .png)")
	}

	// ใช้ UUID ตั้งชื่อไฟล์
	filename := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	filePath := filepath.Join(savePath, filename)

	// เปิดไฟล์ที่อัปโหลดมา
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("ไม่สามารถเปิดไฟล์: %v", err)
	}
	defer src.Close()

	// สร้างไฟล์ปลายทางเพื่อบันทึกข้อมูล
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("ไม่สามารถสร้างไฟล์: %v", err)
	}
	defer dst.Close()

	// คัดลอกข้อมูลจากไฟล์ที่อัปโหลดไปยังไฟล์ปลายทาง
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("บันทึกไฟล์ล้มเหลว: %v", err)
	}

	// คืนค่าลิงก์ของไฟล์
	return "/uploads/" + filename, nil
}
