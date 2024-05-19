//go:build integration

package income

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
)

func TestCreateIncomeIT(t *testing.T) {
	t.Run("create income successfully", func(t *testing.T) {
		sql, err := getTestDatabaseFromConfig()
		if err != nil {
			t.Error(err)
		}
		migration.ApplyMigrations(sql)
		defer migration.RollbackMigrations(sql)

		h := New(sql)
		e := echo.New()
		defer e.Close()

		e.POST("/incomes", h.Create)

		payload := mockIncomeRequest()
		body, err := json.Marshal(payload)
		if err != nil {
			log.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/incomes", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"amount":66.6, "category":"Food", "date":"2024-05-18T12:00:00+07:00",
"id":1, "image_url":"/img/income/1.jpg", "note":"Note1234", "spender_id":5, "transaction_type":"Income"}`, rec.Body.String())
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
