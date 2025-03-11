package route

import (
	"github.com/Kev2406/PEA/internal/controller"
	"github.com/Kev2406/PEA/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// SetupSealRoutes sets up routes for "seals" under /api/seals
func SetupSealRoutes(router fiber.Router, sealController *controller.SealController) {
	api := router.Group("/api")
	seal := api.Group("/seals")

	// -- 1) POST /api/seals/ : create a new Seal
	seal.Post("/", middleware.JWTMiddleware(), sealController.CreateSealHandler)

	// -- 2) POST /api/seals/generate : admin can generate multiple seals
	seal.Post("/generate", middleware.JWTMiddleware(), sealController.GenerateSealsHandler)

	// -- 3) PUT /api/seals/:seal_number/assign : assign a seal to a technician
	seal.Put("/:seal_number/assign", middleware.JWTMiddleware(), sealController.AssignSealToTechnicianHandler)

	// -- 4) POST /api/seals/scan : scan a barcode
	seal.Post("/scan", middleware.JWTMiddleware(), sealController.ScanSealHandler)

	// -- 5) GET /api/seals/report : admin-only report
	seal.Get("/report", middleware.JWTMiddleware(), sealController.GetSealReportHandler)

	// -- 6) GET /api/seals/check : check multiple seals with query params
	seal.Get("/check", sealController.CheckMultipleSealsHandler)

	// -- 7) GET /api/seals/check/:seal_number : check existence of a single seal
	seal.Get("/check/:seal_number", sealController.CheckSealExistsHandler)

	// -- 8) POST /api/seals/issue-multiple : bulk-issue seals from base number
	seal.Post("/issue-multiple", middleware.JWTMiddleware(), sealController.IssueMultipleSealsHandler)

	// -- 9) GET /api/seals/status/:status : get seals by status
	seal.Get("/status/:status", sealController.GetSealsByStatusHandler)

	// -- 10) GET /api/seals/:id/status/:status : get seal by ID & status
	seal.Get("/:id/status/:status", middleware.JWTMiddleware(), sealController.GetSealByIDAndStatusHandler)

	// -- 11) PUT /api/seals/:seal_number/issue : admin issues a seal to a user
	seal.Put("/:seal_number/issue", middleware.JWTMiddleware(), sealController.IssueSealHandler)

	// -- 12) PUT /api/seals/:seal_number/use : user uses a previously issued seal
	seal.Put("/:seal_number/use", middleware.JWTMiddleware(), sealController.UseSealHandler)

	// -- 13) PUT /api/seals/:seal_number/return : user returns a seal after use
	seal.Put("/:seal_number/return", middleware.JWTMiddleware(), sealController.ReturnSealHandler)

	// -- 14) GET /api/seals/:seal_number : get a single seal by number (wildcard route - put last!)
	seal.Get("/:seal_number", middleware.JWTMiddleware(), sealController.GetSealHandler)
	seal.Post("/check", sealController.CheckSealsHandler)

	// -- Potential routes for technicians commented out:
	// seal.Put("/:seal_number/install", middleware.TechnicianJWTMiddleware(), sealController.InstallSealHandler)
	// seal.Put("/:seal_number/return", middleware.TechnicianJWTMiddleware(), sealController.ReturnSealHandler)
}
