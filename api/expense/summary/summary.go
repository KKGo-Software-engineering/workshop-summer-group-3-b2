package summary

import (
	"database/sql"
	"errors"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
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
	Date          time.Time
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

// /api/v1/spenders/{id}/expenses/summary
func (h *handler) GetExpenseSummaryHandler(c echo.Context) error {
	logger := mlog.L(c)
	//ctx := c.Request().Context()
	_ = c.Request().Context()

	var spender Spender
	err := c.Bind(&spender)
	if err != nil {
		logger.Error(ErrInvalidSpender.Error(), zap.Error(err))
		return c.JSON(http.StatusBadRequest, Err{Message: ErrInvalidSpender.Error()})
	}

	return c.JSON(http.StatusOK, Summary{})
}
