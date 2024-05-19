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
		sql := getTestDatabaseFromConfig(t)

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
		sql := getTestDatabaseFromConfig(t)
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

		wantJsonStr := `[{
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
`
		var want []response
		err := json.Unmarshal([]byte(wantJsonStr), &want)
		if err != nil {
			t.Fatal(err)
		}

		var got []response
		err = json.Unmarshal(rec.Body.Bytes(), &got)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(want), len(got))

		for i := range want {
			assert.Equal(t, want[i].Amount, got[i].Amount)
			assert.Equal(t, want[i].Category, got[i].Category)
			assert.Equal(t, want[i].Date, got[i].Date)
			assert.Equal(t, want[i].ImageUrl, got[i].ImageUrl)
			assert.Equal(t, want[i].Note, got[i].Note)
			assert.Equal(t, want[i].SpenderId, got[i].SpenderId)
			assert.Equal(t, want[i].TransactionType, got[i].TransactionType)
		}
	})
}

func getTestDatabaseFromConfig(t *testing.T) *sql.DB {
	t.Helper()
	cfg := config.Parse("DOCKER")
	sql, err := sql.Open("postgres", cfg.PostgresURI())
	if err != nil {
		t.Fatal(err)
	}
	migration.ApplyMigrations(sql)
	t.Cleanup(func() {
		sql.Query("DELETE FROM transaction")
	})
	return sql
}
