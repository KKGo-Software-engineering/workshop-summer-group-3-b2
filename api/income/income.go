package income

import (
	"database/sql"
	"errors"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	insertStatement = `INSERT INTO transaction (date, amount, category, transaction_type, note, image_url, spender_id)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	transactionIncome = "Income"
)

type incomeError struct {
	Message string `json:"message"`
}

type income struct {
	Id              int       `json:"id"`
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	TransactionType string    `json:"transaction_type"`
	Note            string    `json:"note"`
	ImageUrl        string    `json:"image_url"`
	SpenderId       int       `json:"spender_id"`
}
type handler struct {
	db *sql.DB
}

func New(db *sql.DB) *handler {
	return &handler{db}
}

func (h *handler) Create(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()
	var inc income
	err := c.Bind(&inc)
	if err != nil {
		logger.Error("error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, incomeError{Message: "invalid request body"})
	}
	if err = validateIncome(inc); err != nil {
		return c.JSON(http.StatusBadRequest, incomeError{Message: err.Error()})
	}
	inc.TransactionType = transactionIncome
	var lastInsertId int
	err = h.db.QueryRowContext(ctx, insertStatement, inc.Date, inc.Amount, inc.Category,
		inc.TransactionType, inc.Note, inc.ImageUrl, inc.SpenderId).Scan(&lastInsertId)
	if err != nil {
		logger.Error("insert income into transaction table error:", zap.Error(err))
		return c.NoContent(http.StatusInternalServerError)
	}
	inc.Id = lastInsertId
	return c.JSON(http.StatusCreated, inc)
}

func validateIncome(inc income) error {
	if inc.SpenderId == 0 {
		return errors.New("spender id is required")
	} else if inc.Amount < 0.0 {
		return errors.New("amount is lower than 0.0")
	} else if inc.Category == "" {
		return errors.New("category is required")
	}
	return nil
}
