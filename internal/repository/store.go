package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/pismo/api/internal/model"
)

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a unique constraint is violated.
var ErrDuplicate = errors.New("document number already exists")

// Store holds all in-memory data.
type Store struct {
	mu sync.RWMutex

	accounts       map[int64]*model.Account
	accountsByDoc  map[string]int64
	transactions   map[int64]*model.Transaction
	operationTypes map[int64]*model.OperationType

	nextAccountID     int64
	nextTransactionID int64
}

// NewStore creates a new Store pre-seeded with the four operation types.
func NewStore() *Store {
	s := &Store{
		accounts:       make(map[int64]*model.Account),
		accountsByDoc:  make(map[string]int64),
		transactions:   make(map[int64]*model.Transaction),
		operationTypes: make(map[int64]*model.OperationType),
		nextAccountID:     1,
		nextTransactionID: 1,
	}

	// Seed operation types
	for id, desc := range map[int64]string{
		1: "PURCHASE",
		2: "INSTALLMENT PURCHASE",
		3: "WITHDRAWAL",
		4: "PAYMENT",
	} {
		s.operationTypes[id] = &model.OperationType{OperationTypeID: id, Description: desc}
	}

	return s
}

// --- Accounts ---

func (s *Store) CreateAccount(docNumber string) (*model.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accountsByDoc[docNumber]; exists {
		return nil, ErrDuplicate
	}

	acc := &model.Account{
		AccountID:      s.nextAccountID,
		DocumentNumber: docNumber,
	}
	s.accounts[acc.AccountID] = acc
	s.accountsByDoc[docNumber] = acc.AccountID
	s.nextAccountID++
	return acc, nil
}

func (s *Store) GetAccount(id int64) (*model.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, ok := s.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return acc, nil
}

// --- Operation Types ---

func (s *Store) GetOperationType(id int64) (*model.OperationType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ot, ok := s.operationTypes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return ot, nil
}

// --- Transactions ---

// debtOperations holds operation type IDs that should be stored as negative amounts.
var debtOperations = map[int64]bool{1: true, 2: true, 3: true}

func (s *Store) CreateTransaction(accountID, operationTypeID int64, amount float64) (*model.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Enforce sign convention
	if debtOperations[operationTypeID] && amount > 0 {
		amount = -amount
	} else if !debtOperations[operationTypeID] && amount < 0 {
		amount = -amount
	}

	tx := &model.Transaction{
		TransactionID:   s.nextTransactionID,
		AccountID:       accountID,
		OperationTypeID: operationTypeID,
		Amount:          amount,
		EventDate:       time.Now().UTC(),
	}
	s.transactions[tx.TransactionID] = tx
	s.nextTransactionID++
	return tx, nil
}
