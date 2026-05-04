package service

import (
	"errors"

	"github.com/pismo/api/internal/model"
	"github.com/pismo/api/internal/repository"
)

// ErrNotFound é reexportado para uso no manipulador.
var ErrNotFound = repository.ErrNotFound

// ErrDuplicate é reexportado para uso do manipulador.
var ErrDuplicate = repository.ErrDuplicate

// AccountService handles business logic for accounts.
type AccountService struct {
	store *repository.Store
}

func NewAccountService(store *repository.Store) *AccountService {
	return &AccountService{store: store}
}

func (s *AccountService) Create(docNumber string) (*model.Account, error) {
	if docNumber == "" {
		return nil, errors.New("document_number is required")
	}
	return s.store.CreateAccount(docNumber)
}

func (s *AccountService) Get(id int64) (*model.Account, error) {
	return s.store.GetAccount(id)
}

// O TransactionService lida com a lógica de negócios das transações.
type TransactionService struct {
	store *repository.Store
}

func NewTransactionService(store *repository.Store) *TransactionService {
	return &TransactionService{store: store}
}

func (s *TransactionService) Create(accountID, operationTypeID int64, amount float64) (*model.Transaction, error) {
	if amount == 0 {
		return nil, errors.New("amount must not be zero")
	}

	// Verifique se a conta existe.
	if _, err := s.store.GetAccount(accountID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// Validar se o tipo de operação existe
	if _, err := s.store.GetOperationType(operationTypeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, errors.New("operation_type not found")
		}
		return nil, err
	}

	return s.store.CreateTransaction(accountID, operationTypeID, amount)
}
