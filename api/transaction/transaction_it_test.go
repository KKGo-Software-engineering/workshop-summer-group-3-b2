//go:build integration

package transaction

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/migration"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateTransactionIT(t *testing.T) {
	t.Run("create transaction successfully", func(t *testing.T) {
		sql, err := getTestDatabaseFromConfig()
		if err != nil {
			t.Error(err)
		}
		migration.ApplyMigrations(sql)
		defer migration.RollbackMigrations(sql)

		h := New(sql)
		e := echo.New()
		defer e.Close()

		e.POST("/transactions", h.Create)

		payload := mockTransactionRequest()
		body, err := json.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

func TestGetTransactionIT(t *testing.T) {
	t.Run("create get transactions successfully", func(t *testing.T) {
		sql, err := getTestDatabaseFromConfig()
		if err != nil {
			t.Error(err)
		}
		migration.ApplyMigrations(sql)
		defer migration.RollbackMigrations(sql)
		h := New(sql)
		e := echo.New()
		defer e.Close()
		date1, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		date2, _ := time.Parse(time.RFC3339, "2024-05-18T15:51:49.673703Z")
		sql.Exec(insertStatement, date1, 66.6, "Food", "EXPENSE", "Note1234", "/img/transaction/1.jpg", 5)
		sql.Exec(insertStatement, date2, 70.6, "Food", "EXPENSE", "Note555", "/img/transaction/2.jpg", 5)
		e.GET("/transactions", h.GetAll)

		req := httptest.NewRequest(http.MethodGet, "/transactions?transaction_type=EXPENSE", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[{
    "id": 1,
    "date": "2024-05-18T11:51:49.673703Z",
    "amount": 66.6,
    "category": "Food",
    "transaction_type": "EXPENSE",
    "note": "Note1234",
    "image_url": "/img/transaction/1.jpg",
    "spender_id": 5
  },
  {
    "id": 2,
    "date": "2024-05-18T15:51:49.673703Z",
    "amount": 70.6,
    "category": "Food",
    "transaction_type": "EXPENSE",
    "note": "Note555",
    "image_url": "/img/transaction/2.jpg",
    "spender_id": 5
  }]
`, rec.Body.String())
	})
}

func getTestDatabaseFromConfig() (*sql.DB, error) {
	cfg := config.Parse("DOCKER")
	sql, err := sql.Open("postgres", cfg.PostgresURI())
	if err != nil {
		return nil, err
	}
	return sql, nil
}
