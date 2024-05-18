package summary

import (
	"github.com/KKGo-Software-engineering/workshop-summer/api/expense"
	"time"
)

type Spender struct {
	ID int `param:"id"`
}

type Data struct {
	Date          time.Time
	SumAmount     float64
	CountExpenses int
}

type Summary struct {
	Total   float64 `json:"total"`
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type Storer interface {
	GetExpenses(spenderID int) ([]expense.Expense, error)
}

type Handler struct {
	store Storer
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

func summary(data []Data) Summary {
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

//func (h *Handler) GetExpenseSummaryHandler(c echo.Context) error {
//	var spender Spender
//	if err := c.Bind(&spender); err != nil {
//		return c.JSON(http.StatusBadRequest, err)
//	}
//
//	expenses := h.store.GetExpenses()
//
//}
