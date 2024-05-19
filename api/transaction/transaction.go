package transaction

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
)

type transactionError struct {
	Message string `json:"message"`
}

type transactionRequest struct {
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
	var req transactionRequest
	err := c.Bind(&req)
	if err != nil {
		logger.Error("error", zap.Error(err))
		return c.JSON(http.StatusBadRequest, transactionError{Message: "invalid request body"})
	}
	if err = validateTransaction(req); err != nil {
		return c.JSON(http.StatusBadRequest, transactionError{Message: err.Error()})
	}
	var lastInsertId int
	err = h.db.QueryRowContext(ctx, insertStatement, req.Date, req.Amount, req.Category,
		req.TransactionType, req.Note, req.ImageUrl, req.SpenderId).Scan(&lastInsertId)
	if err != nil {
		logger.Error("insert transaction into transaction table error:", zap.Error(err))
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

func validateTransaction(req transactionRequest) error {
	if req.Amount < 0.0 {
		return errors.New("amount is lower than 0.0")
	} else if req.Category == "" {
		return errors.New("category is required")
	}
	return nil
}
