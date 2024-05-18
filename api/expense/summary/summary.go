package summary

import (
	"database/sql"
	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"time"
)

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

//func (h *handler) GetExpenseSummaryHandler(c echo.Context) error {
//	var spender Spender
//	if err := c.Bind(&spender); err != nil {
//		return c.JSON(http.StatusBadRequest, err)
//	}
//
//	expenses := h.store.GetExpenses()
//
//}
