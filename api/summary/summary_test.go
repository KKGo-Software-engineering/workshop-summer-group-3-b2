package summary

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSummary(t *testing.T) {
	testCases := []struct {
		name string
		data []RawData
		want Summary
	}{
		{
			name: "empty data",
			data: []RawData{},
			want: Summary{Total: 0, Average: 0, Count: 0},
		},
		{
			name: "single data",
			data: []RawData{
				{SumAmount: 10, CountExpenses: 1},
			},
			want: Summary{Total: 10, Average: 10, Count: 1},
		},
		{
			name: "multiple data",
			data: []RawData{
				{SumAmount: 20, CountExpenses: 2},
				{SumAmount: 30, CountExpenses: 3},
			},
			want: Summary{Total: 50, Average: 25, Count: 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := summary(tc.data)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetExpenseSummaryHandler(t *testing.T) {

	t.Run("invalid spender id expect 400", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/expenses/summary")
		c.SetParamNames("id")
		c.SetParamValues("not_int")

		db, _, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		h := New(config.FeatureFlag{}, db)
		_ = h.GetExpenseSummaryHandler(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("get summary succesfully", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/expenses/summary")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"transaction_date", "total_amount", "record_count"}).
			AddRow("2024-04-03", 1000, 10).
			AddRow("2024-04-04", 500, 5)

		mock.ExpectPrepare(sumSQL).ExpectQuery().WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetExpenseSummaryHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"total_amount": 1500, "average_per_day": 750, "count_transaction": 15}`, rec.Body.String())
	})

	t.Run("get summary failed on database", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/expenses/summary")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectPrepare(sumSQL).WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err := h.GetExpenseSummaryHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("query error", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/expenses/summary")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		mock.ExpectPrepare(sumSQL).ExpectQuery().WillReturnError(assert.AnError)

		h := New(config.FeatureFlag{}, db)
		err := h.GetExpenseSummaryHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("scan error", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/spenders/:id/expenses/summary")
		c.SetParamNames("id")
		c.SetParamValues("1")

		db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		defer db.Close()

		rows := sqlmock.NewRows([]string{"transaction_date", "total_amount"}).
			AddRow("2024-04-03", 1000).
			AddRow("2024-04-04", 500)
		mock.ExpectPrepare(sumSQL).ExpectQuery().WillReturnRows(rows)

		h := New(config.FeatureFlag{}, db)
		err := h.GetExpenseSummaryHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

}
