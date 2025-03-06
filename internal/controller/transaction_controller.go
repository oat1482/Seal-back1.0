package controller

import (
	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/service"
	"github.com/gofiber/fiber/v2"
)

type TransactionController struct {
	transactionService *service.TransactionService
}

func NewTransactionController(transactionService *service.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

func (tc *TransactionController) CreateTransactionHandler(c *fiber.Ctx) error {
	var transaction model.Transaction
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	err := tc.transactionService.CreateTransaction(&transaction)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create transaction"})
	}

	return c.JSON(transaction)
}

func (tc *TransactionController) GetAllTransactionsHandler(c *fiber.Ctx) error {
	transactions, err := tc.transactionService.GetAllTransactions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch transactions"})
	}

	return c.JSON(transactions)
}
