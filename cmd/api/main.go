package main

import (
	"log"
	"net/http"

	"github.com/pismo/api/internal/handler"
	"github.com/pismo/api/internal/repository"
	"github.com/pismo/api/internal/service"
)

func main() {
	// Wire up dependencies
	store := repository.NewStore()

	accountSvc := service.NewAccountService(store)
	transactionSvc := service.NewTransactionService(store)

	accountHandler := handler.NewAccountHandler(accountSvc)
	transactionHandler := handler.NewTransactionHandler(transactionSvc)

	mux := http.NewServeMux()

	// Account routes
	mux.Handle("/accounts", accountHandler)
	mux.Handle("/accounts/", accountHandler)

	// Transaction routes
	mux.Handle("/transactions", transactionHandler)

	// OpenAPI spec
	mux.HandleFunc("/swagger.json", swaggerHandler)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(swaggerSpec))
}

const swaggerSpec = `{
  "openapi": "3.0.0",
  "info": {
    "title": "Pismo API",
    "version": "1.0.0",
    "description": "Customer Account & Transactions API"
  },
  "paths": {
    "/accounts": {
      "post": {
        "summary": "Create account",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["document_number"],
                "properties": {
                  "document_number": { "type": "string", "example": "12345678900" }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Account created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Account" }
              }
            }
          },
          "400": { "description": "Bad request" },
          "409": { "description": "Document number already exists" }
        }
      }
    },
    "/accounts/{accountId}": {
      "get": {
        "summary": "Get account by ID",
        "parameters": [
          {
            "in": "path",
            "name": "accountId",
            "required": true,
            "schema": { "type": "integer" }
          }
        ],
        "responses": {
          "200": {
            "description": "Account found",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Account" }
              }
            }
          },
          "400": { "description": "Invalid account ID" },
          "404": { "description": "Account not found" }
        }
      }
    },
    "/transactions": {
      "post": {
        "summary": "Create transaction",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["account_id", "operation_type_id", "amount"],
                "properties": {
                  "account_id":       { "type": "integer", "example": 1 },
                  "operation_type_id": { "type": "integer", "example": 4 },
                  "amount":           { "type": "number",  "example": 123.45 }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Transaction created",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/Transaction" }
              }
            }
          },
          "400": { "description": "Bad request" }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Account": {
        "type": "object",
        "properties": {
          "account_id":      { "type": "integer" },
          "document_number": { "type": "string" }
        }
      },
      "Transaction": {
        "type": "object",
        "properties": {
          "transaction_id":   { "type": "integer" },
          "account_id":       { "type": "integer" },
          "operation_type_id": { "type": "integer" },
          "amount":           { "type": "number" },
          "event_date":       { "type": "string", "format": "date-time" }
        }
      }
    }
  }
}`
