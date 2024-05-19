package income

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type anyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func mockIncomeRequest() income {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}
	return income{
		Date:      time.Date(2024, time.May, 18, 12, 0, 0, 0, loc),
		Amount:    66.6,
		Category:  "Food",
		Note:      "Note1234",
		ImageUrl:  "/img/income/1.jpg",
		SpenderId: 5,
	}
}

func setupTest(inc income) (echo.Context, *httptest.ResponseRecorder) {
	body, err := json.Marshal(inc)
	if err != nil {
		log.Fatal(err)
	}
	e := echo.New()
	defer e.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/income", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestCreateIncome(t *testing.T) {
	t.Run("Create Income Successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockIncomeRequest()
		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		c, rec := setupTest(req)
		mock.ExpectQuery(insertStatement).WithArgs(anyTime{}, req.Amount, req.Category,
			transactionIncome, req.Note, req.ImageUrl, req.SpenderId).WillReturnRows(row)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"amount":66.6, "category":"Food", "date":"2024-05-18T12:00:00+07:00",
"id":1, "image_url":"/img/income/1.jpg", "note":"Note1234", "spender_id":5, "transaction_type":"Income"}`, rec.Body.String())
	})
	t.Run("Create Income Successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockIncomeRequest()
		c, rec := setupTest(req)
		mock.ExpectQuery(insertStatement).WithArgs(anyTime{}, req.Amount, req.Category,
			transactionIncome, req.Note, req.ImageUrl, req.SpenderId).WillReturnError(errors.New("error db"))
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("Create Income fail request body is invalid", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		e := echo.New()
		defer e.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/income", strings.NewReader("test"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"invalid request body"}`, rec.Body.String())
	})
	t.Run("Create Income fail spender id is 0", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockIncomeRequest()
		req.SpenderId = 0
		c, rec := setupTest(req)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"spender id is required"}`, rec.Body.String())
	})
	t.Run("Create Income fail amount is lower than 0.0", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockIncomeRequest()
		req.Amount = -1
		c, rec := setupTest(req)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"amount is lower than 0.0"}`, rec.Body.String())
	})
	t.Run("Create Income fail category is empty", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockIncomeRequest()
		req.Category = ""
		c, rec := setupTest(req)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"category is required"}`, rec.Body.String())
	})
}
