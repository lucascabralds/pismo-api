# Pismo Code Assessment — Phase 1

REST API for **Customer Account & Transactions** built with Go (standard library only — no external dependencies).

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.22 |
| HTTP | `net/http` (stdlib) |
| Storage | In-memory (thread-safe) |
| Tests | `testing` package (stdlib) |
| Docs | OpenAPI 3.0 JSON served at `/swagger.json` |
| Container | Docker / Docker Compose |

---

## Running locally

### Prerequisites
- Go 1.22+

```bash
go run ./cmd/api
```

Server starts on **http://localhost:8080**.

---

### Running with Docker

```bash
docker compose up --build
```

---

### Running tests

```bash
go test ./...
```

---

## API Endpoints

### `POST /accounts` — Create an account

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number": "12345678900"}'
```

Response `201 Created`:
```json
{
  "account_id": 1,
  "document_number": "12345678900"
}
```

---

### `GET /accounts/:accountId` — Retrieve an account

```bash
curl http://localhost:8080/accounts/1
```

Response `200 OK`:
```json
{
  "account_id": 1,
  "document_number": "12345678900"
}
```

---

### `POST /transactions` — Create a transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "operation_type_id": 4, "amount": 123.45}'
```

Response `201 Created`:
```json
{
  "transaction_id": 1,
  "account_id": 1,
  "operation_type_id": 4,
  "amount": 123.45,
  "event_date": "2024-01-01T10:00:00Z"
}
```

**Operation types:**

| ID | Description | Amount sign |
|---|---|---|
| 1 | PURCHASE | negative |
| 2 | INSTALLMENT PURCHASE | negative |
| 3 | WITHDRAWAL | negative |
| 4 | PAYMENT | positive |

---

### `GET /swagger.json` — OpenAPI 3.0 specification

```bash
curl http://localhost:8080/swagger.json
```

---

## Business rules

- `document_number` must be unique per account.
- `account_id` and `operation_type_id` must reference existing records; otherwise a `400` is returned.
- Debt operations (purchase, installment purchase, withdrawal) are always stored with **negative** amounts.
- Payments are always stored with **positive** amounts.
- The `event_date` is automatically set to the UTC time of the request.

---

## Project structure

```
.
├── cmd/api/main.go               # Entry point + OpenAPI spec
├── internal/
│   ├── handler/handler.go        # HTTP handlers
│   ├── handler/handler_test.go   # Tests
│   ├── service/service.go        # Business logic
│   ├── repository/store.go       # In-memory storage
│   └── model/model.go            # Data types & DTOs
├── Dockerfile
├── docker-compose.yml
└── README.md
```
