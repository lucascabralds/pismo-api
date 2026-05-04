package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/pismo/api/internal/model"
	"github.com/pismo/api/internal/repository"
	"github.com/pismo/api/internal/service"
)

// --- Ajudantes ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}

func decodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// pathID extrai o último segmento do caminho e o analisa como um int64..
// e.g. "/accounts/42" → 42
func pathID(r *http.Request) (int64, error) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	raw := parts[len(parts)-1]
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// --- Account handler ---

type AccountHandler struct {
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

// ServeHTTP envia solicitações GET para /accounts/:id e POST para /accounts.
func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// POST /accounts
	if r.Method == http.MethodPost && strings.TrimSuffix(r.URL.Path, "/") == "/accounts" {
		h.create(w, r)
		return
	}
	// GET /accounts/:id
	if r.Method == http.MethodGet {
		h.get(w, r)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *AccountHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAccountRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	acc, err := h.svc.Create(req.DocumentNumber)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, acc)
}

func (h *AccountHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	acc, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

// --- manipulador de transações ---

type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.CreateTransactionRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Create(req.AccountID, req.OperationTypeID, req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, tx)
}
