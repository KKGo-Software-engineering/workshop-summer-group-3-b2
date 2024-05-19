package summary

import (
	"database/sql"
	"errors"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

const (
	typeExpense = "expense"
	typeIncome  = "income"
)

var (
	ErrInvalidSpender = errors.New("invalid spender")
)

type Err struct {
	Message string `json:"message"`
}

type Spender struct {
	ID int `param:"id"`
}

type RawData struct {
	Date          string
	SumAmount     float64
	CountExpenses int
}

type Summary struct {
	Total   float64 `json:"total_amount"`
	Average float64 `json:"average_per_day"`
	Count   int     `json:"count_transaction"`
}

type handler struct {
	flag config.FeatureFlag
	db   *sql.DB
}

const (
	sumSQL = `SELECT
	    date_trunc('day', date)::date AS transaction_date,
	    SUM(amount) AS total_amount,
	    COUNT(*) AS record_count
	FROM
	    "transaction"
	WHERE
	    transaction_type = $1 AND spender_id = $2
	GROUP BY
	    date_trunc('day', date)::date
	ORDER BY
	    transaction_date;`
)

func New(cfg config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfg, db}
}

func summary(data []RawData) Summary {
	if len(data) == 0 {
		return Summary{}
	}

	var total float64
	var count int
	for _, d := range data {
		total += d.SumAmount
		count += d.CountExpenses
	}

	return Summary{
		Total:   total,
		Average: total / float64(len(data)),
		Count:   count,
	}
}

func processSummaryRequest(c echo.Context, db *sql.DB, tnxType string) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	var spender Spender
	err := c.Bind(&spender)
	if err != nil {
		logger.Error(ErrInvalidSpender.Error(), zap.Error(err))
		return c.JSON(http.StatusBadRequest, Err{Message: ErrInvalidSpender.Error()})
	}

	stmt, err := db.PrepareContext(ctx, sumSQL)
	if err != nil {
		logger.Error("prepare statement error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, Err{Message: "prepare statement error"})
	}

	rows, err := stmt.QueryContext(ctx, tnxType, spender.ID)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, Err{Message: "query error"})
	}
	defer rows.Close()

	var raws []RawData
	for rows.Next() {
		var raw RawData
		err := rows.Scan(&raw.Date, &raw.SumAmount, &raw.CountExpenses)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, Err{Message: "scan error"})
		}
		raws = append(raws, raw)
	}

	return c.JSON(http.StatusOK, summary(raws))
}

func (h *handler) GetExpenseSummaryHandler(c echo.Context) error {
	return processSummaryRequest(c, h.db, typeExpense)
}

func (h *handler) GetIncomeSummaryHandler(c echo.Context) error {
	return processSummaryRequest(c, h.db, typeIncome)
}
