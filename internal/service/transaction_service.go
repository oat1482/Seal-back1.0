package service

import (
	"errors"

	"github.com/Kev2406/PEA/internal/domain/model"
	"github.com/Kev2406/PEA/internal/domain/repository"
)

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// บันทึก Transaction
func (s *TransactionService) CreateTransaction(transaction *model.Transaction) error {
	if transaction.SealID == 0 || transaction.UserID == 0 || transaction.Action == "" {
		return errors.New("missing required fields")
	}
	return s.repo.Create(transaction)
}

// ดึง Transaction ทั้งหมด
func (s *TransactionService) GetAllTransactions() ([]model.Transaction, error) {
	return s.repo.GetAll()
}

// ดึง Transaction ตาม ID
func (s *TransactionService) GetTransactionByID(id uint) (*model.Transaction, error) {
	return s.repo.GetByID(id)
}
