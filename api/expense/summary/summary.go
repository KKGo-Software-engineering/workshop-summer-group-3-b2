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

// total amount spent, the average amount spent per day, and the total number of expenses
//
//	{
//		"summary": {
//			"total_income": 2000,
//			"total_expenses": 1000,
//			"current_balance": 1000
//		}
//	}
type Summary struct {
	TotalAmount      float64 `json:"total_amount"`
	AveragePerDay    float64 `json:"average_per_day"`
	CountTransaction int     `json:"count_transaction"`
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
		TotalAmount:      total,
		AveragePerDay:    total / float64(len(data)),
		CountTransaction: count,
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
