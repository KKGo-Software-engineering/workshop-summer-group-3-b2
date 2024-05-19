package transaction

import (
	"database/sql"
	"errors"
	"fmt"
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

type request struct {
	Date            time.Time `json:"date"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	TransactionType string    `json:"transaction_type"`
	Note            string    `json:"note"`
	ImageUrl        string    `json:"image_url"`
	SpenderId       int       `json:"spender_id"`
}

type response struct {
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
	var req request
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

func validateTransaction(req request) error {
	if req.Amount < 0.0 {
		return errors.New("amount is lower than 0.0")
	} else if req.Category == "" {
		return errors.New("category is required")
	}
	return nil
}

func (h handler) GetAll(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()
	tranType := c.QueryParam("transaction_type")
	fmt.Println(tranType)
	if tranType != "EXPENSE" && tranType != "INCOME" {
		return c.JSON(http.StatusBadRequest, transactionError{Message: "invalid transaction type"})
	}
	rows, err := h.db.QueryContext(ctx, `SELECT id, date, amount, category, note, image_url, spender_id, transaction_type FROM transaction where transaction_type = $1`, tranType)
	if err != nil {
		logger.Error("query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	var res []response
	for rows.Next() {
		var t response
		err := rows.Scan(&t.Id, &t.Date, &t.Amount, &t.Category, &t.Note, &t.ImageUrl, &t.SpenderId, &t.TransactionType)
		if err != nil {
			logger.Error("scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		res = append(res, t)
	}
	return c.JSON(http.StatusOK, res)
}
