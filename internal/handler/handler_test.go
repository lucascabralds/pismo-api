package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pismo/api/internal/handler"
	"github.com/pismo/api/internal/model"
	"github.com/pismo/api/internal/repository"
	"github.com/pismo/api/internal/service"
)

// helpers

func newDeps() (*handler.AccountHandler, *handler.TransactionHandler) {
	store := repository.NewStore()
	accSvc := service.NewAccountService(store)
	txSvc := service.NewTransactionService(store)
	return handler.NewAccountHandler(accSvc), handler.NewTransactionHandler(txSvc)
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		t.Fatal(err)
	}
	return buf
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// --- Account tests ---

func TestCreateAccount_Success(t *testing.T) {
	accH, _ := newDeps()

	body := toJSON(t, model.CreateAccountRequest{DocumentNumber: "12345678900"})
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	w := httptest.NewRecorder()

	accH.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var acc model.Account
	decodeResponse(t, w, &acc)

	if acc.AccountID != 1 {
		t.Errorf("expected account_id=1, got %d", acc.AccountID)
	}
	if acc.DocumentNumber != "12345678900" {
		t.Errorf("unexpected document_number: %s", acc.DocumentNumber)
	}
}

func TestCreateAccount_DuplicateDocument(t *testing.T) {
	accH, _ := newDeps()

	body := toJSON(t, model.CreateAccountRequest{DocumentNumber: "99999999999"})

	// first request
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	w := httptest.NewRecorder()
	accH.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// second request with same document
	body2 := toJSON(t, model.CreateAccountRequest{DocumentNumber: "99999999999"})
	req2 := httptest.NewRequest(http.MethodPost, "/accounts", body2)
	w2 := httptest.NewRecorder()
	accH.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w2.Code)
	}
}

func TestCreateAccount_MissingDocumentNumber(t *testing.T) {
	accH, _ := newDeps()

	body := toJSON(t, model.CreateAccountRequest{DocumentNumber: ""})
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	w := httptest.NewRecorder()

	accH.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetAccount_Success(t *testing.T) {
	accH, _ := newDeps()

	// create account first
	body := toJSON(t, model.CreateAccountRequest{DocumentNumber: "11111111111"})
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	w := httptest.NewRecorder()
	accH.ServeHTTP(w, req)

	// retrieve it
	req2 := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	w2 := httptest.NewRecorder()
	accH.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	var acc model.Account
	decodeResponse(t, w2, &acc)
	if acc.AccountID != 1 {
		t.Errorf("expected account_id=1, got %d", acc.AccountID)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	accH, _ := newDeps()

	req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
	w := httptest.NewRecorder()
	accH.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetAccount_InvalidID(t *testing.T) {
	accH, _ := newDeps()

	req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
	w := httptest.NewRecorder()
	accH.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Transaction tests ---

func createAccountForTest(t *testing.T, accH *handler.AccountHandler, doc string) model.Account {
	t.Helper()
	body := toJSON(t, model.CreateAccountRequest{DocumentNumber: doc})
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	w := httptest.NewRecorder()
	accH.ServeHTTP(w, req)
	var acc model.Account
	decodeResponse(t, w, &acc)
	return acc
}

func TestCreateTransaction_Payment(t *testing.T) {
	accH, txH := newDeps()
	acc := createAccountForTest(t, accH, "22222222222")

	body := toJSON(t, model.CreateTransactionRequest{
		AccountID:       acc.AccountID,
		OperationTypeID: 4, // PAYMENT
		Amount:          123.45,
	})
	req := httptest.NewRequest(http.MethodPost, "/transactions", body)
	w := httptest.NewRecorder()
	txH.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var tx model.Transaction
	decodeResponse(t, w, &tx)

	if tx.Amount <= 0 {
		t.Errorf("payment should be stored as positive, got %f", tx.Amount)
	}
}

func TestCreateTransaction_Purchase_NegativeAmount(t *testing.T) {
	accH, txH := newDeps()
	acc := createAccountForTest(t, accH, "33333333333")

	body := toJSON(t, model.CreateTransactionRequest{
		AccountID:       acc.AccountID,
		OperationTypeID: 1, // PURCHASE
		Amount:          50.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/transactions", body)
	w := httptest.NewRecorder()
	txH.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var tx model.Transaction
	decodeResponse(t, w, &tx)

	if tx.Amount >= 0 {
		t.Errorf("purchase should be stored as negative, got %f", tx.Amount)
	}
}

func TestCreateTransaction_InvalidAccount(t *testing.T) {
	_, txH := newDeps()

	body := toJSON(t, model.CreateTransactionRequest{
		AccountID:       999,
		OperationTypeID: 1,
		Amount:          50.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/transactions", body)
	w := httptest.NewRecorder()
	txH.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_InvalidOperationType(t *testing.T) {
	accH, txH := newDeps()
	acc := createAccountForTest(t, accH, "44444444444")

	body := toJSON(t, model.CreateTransactionRequest{
		AccountID:       acc.AccountID,
		OperationTypeID: 99, // invalid
		Amount:          50.0,
	})
	req := httptest.NewRequest(http.MethodPost, "/transactions", body)
	w := httptest.NewRecorder()
	txH.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
