package transaction

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
	_, ok = v.(int)
	_, ok = v.(string)
	_, ok = v.(float64)
	return ok
}

func mockTransactionRequest() request {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal(err)
	}
	return request{
		Date:            time.Date(2024, time.May, 18, 12, 0, 0, 0, loc),
		Amount:          66.6,
		Category:        "Food",
		Note:            "Note1234",
		ImageUrl:        "/img/transaction/1.jpg",
		TransactionType: "INCOME",
		SpenderId:       5,
	}
}

func setupTest(transaction request) (echo.Context, *httptest.ResponseRecorder) {
	body, err := json.Marshal(transaction)
	if err != nil {
		log.Fatal(err)
	}
	e := echo.New()
	defer e.Close()
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func setupUpdateOrDeleteTest(method string, transaction request) (echo.Context, *httptest.ResponseRecorder) {
	body, err := json.Marshal(transaction)
	if err != nil {
		log.Fatal(err)
	}
	e := echo.New()
	defer e.Close()
	req := httptest.NewRequest(method, "/spenders/1/transactions/1", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestCreateTransaction(t *testing.T) {
	t.Run("Create Transaction Successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		row := sqlmock.NewRows([]string{"id"}).AddRow(1)
		c, rec := setupTest(req)
		mock.ExpectQuery(insertStatement).WillReturnRows(row)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})
	t.Run("Create Transaction fail request body is invalid", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		e := echo.New()
		defer e.Close()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transaction", strings.NewReader("test"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"invalid request body"}`, rec.Body.String())
	})
	t.Run("Create Transaction fail amount is lower than 0.0", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		req.Amount = -1
		c, rec := setupTest(req)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"amount is lower than 0.0"}`, rec.Body.String())
	})
	t.Run("Create Transaction fail category is empty", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		req.Category = ""
		c, rec := setupTest(req)
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"category is required"}`, rec.Body.String())
	})
	t.Run("Create Transaction fail insert into db error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		c, rec := setupTest(req)
		mock.ExpectQuery(insertStatement).WithArgs(anyTime{}, req.Amount, req.Category,
			req.TransactionType, req.Note, req.ImageUrl, req.SpenderId).WillReturnError(errors.New("error"))
		h := New(db)
		err = h.Create(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetAllExpense(t *testing.T) {
	t.Run("get all expense successfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/spenders/1/transaction?transaction_type=EXPENSE", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		date1, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		date2, _ := time.Parse(time.RFC3339, "2024-05-18T15:51:49.673703Z")
		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "note", "image_url", "spender_id", "transaction_type"}).
			AddRow(1, date1, 1000, "Lunch", "MOCK", "location/on/s3/bucket/eslip1", 1, "EXPENSE").
			AddRow(2, date2, 2000, "Dinner", "MOCK", "location/on/s3/bucket/eslip2", 2, "EXPENSE")
		mock.ExpectQuery(`SELECT id, date, amount, category, note, image_url, spender_id, transaction_type FROM transaction where transaction_type = $1 and spender_id = $2`).WillReturnRows(rows)
		h := New(db)
		err := h.GetAllBySpender(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[{"id":1,"date":"2024-05-18T11:51:49.673703Z","amount":1000,"category":"Lunch","note":"MOCK","image_url":"location/on/s3/bucket/eslip1","spender_id":1,"transaction_type":"EXPENSE"},
{"id":2,"date":"2024-05-18T15:51:49.673703Z","amount":2000,"category":"Dinner","note":"MOCK","image_url":"location/on/s3/bucket/eslip2","spender_id":2,"transaction_type":"EXPENSE"}]`, rec.Body.String())
	})
	t.Run("get all expense fail incorrect transaction_type", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/spenders/1/transaction?transaction_type=TEST", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		date1, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		date2, _ := time.Parse(time.RFC3339, "2024-05-18T15:51:49.673703Z")
		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "note", "image_url", "spender_id", "transaction_type"}).
			AddRow(1, date1, 1000, "Lunch", "MOCK", "location/on/s3/bucket/eslip1", 1, "EXPENSE").
			AddRow(2, date2, 2000, "Dinner", "MOCK", "location/on/s3/bucket/eslip2", 2, "EXPENSE")
		mock.ExpectQuery(`SELECT id, date, amount, category, note, image_url, spender_id, transaction_type FROM transaction where transaction_type = $1`).WithArgs("EXPENSE").WillReturnRows(rows)
		h := New(db)
		err := h.GetAllBySpender(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"invalid transaction type"}`, rec.Body.String())
	})
	t.Run("get all expense failed on database", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/spenders/1/transaction?transaction_type=EXPENSE", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectQuery(`SELECT id, date, amount, category, note, image_url, spender_id, transaction_type FROM transaction where transaction_type = 'EXPENSE'`).WillReturnError(assert.AnError)

		h := New(db)
		err := h.GetAllBySpender(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("get all expense successfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/spenders/1/transaction?transaction_type=EXPENSE", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "date", "amount", "category", "note", "image_url", "spender_id", "transaction_type"}).
			AddRow("", "", 1000, "Lunch", "MOCK", "location/on/s3/bucket/eslip1", 1, "EXPENSE").
			AddRow("", "date2", 2000, "Dinner", "MOCK", "location/on/s3/bucket/eslip2", 2, "EXPENSE")
		mock.ExpectQuery(`SELECT id, date, amount, category, note, image_url, spender_id, transaction_type FROM transaction where transaction_type = $1`).WithArgs("EXPENSE").WillReturnRows(rows)
		h := New(db)
		err := h.GetAllBySpender(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUpdateTransaction(t *testing.T) {
	t.Run("Update Transaction Successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		date, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		req := mockTransactionRequest()
		req.Date = date
		c, rec := setupUpdateOrDeleteTest(http.MethodPut, req)
		mock.ExpectExec(updateStatment).WillReturnResult(sqlmock.NewResult(1, 1))
		h := New(db)
		err = h.Update(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("Update Transaction fail request body is invalid", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		e := echo.New()
		defer e.Close()
		req := httptest.NewRequest(http.MethodPut, "/spenders/1/transactions/1", strings.NewReader("123123123"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := New(db)
		err = h.Update(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("Update Transaction fail amount is lower than 0.0", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		date, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		req := mockTransactionRequest()
		req.Amount = -1
		req.Date = date
		c, rec := setupUpdateOrDeleteTest(http.MethodPut, req)
		h := New(db)
		err = h.Update(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"amount is lower than 0.0"}`, rec.Body.String())
	})
	t.Run("Update Transaction fail category is empty", func(t *testing.T) {
		db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		req.Category = ""
		c, rec := setupUpdateOrDeleteTest(http.MethodPut, req)
		h := New(db)
		err = h.Update(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"message":"category is required"}`, rec.Body.String())
	})
	t.Run("Update Transaction fail db error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		c, rec := setupUpdateOrDeleteTest(http.MethodPut, req)
		mock.ExpectExec(updateStatment).WillReturnError(errors.New("error"))
		h := New(db)
		err = h.Update(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteTransaction(t *testing.T) {
	t.Run("Delete Transaction Successfully", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		date, _ := time.Parse(time.RFC3339, "2024-05-18T11:51:49.673703Z")
		req := mockTransactionRequest()
		req.Date = date
		c, rec := setupUpdateOrDeleteTest(http.MethodDelete, req)
		mock.ExpectExec(deleteStatment).WillReturnResult(sqlmock.NewResult(1, 1))
		h := New(db)
		err = h.Delete(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("Delete Transaction fail db error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			log.Fatal(err)
		}
		req := mockTransactionRequest()
		c, rec := setupUpdateOrDeleteTest(http.MethodDelete, req)
		mock.ExpectExec(deleteStatment).WillReturnError(errors.New("error"))
		h := New(db)
		err = h.Delete(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
