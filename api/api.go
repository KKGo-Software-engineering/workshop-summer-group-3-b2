package api

import (
	"database/sql"
	"github.com/KKGo-Software-engineering/workshop-summer/api/summary"
	"github.com/KKGo-Software-engineering/workshop-summer/api/transaction"

	"github.com/KKGo-Software-engineering/workshop-summer/api/config"
	"github.com/KKGo-Software-engineering/workshop-summer/api/eslip"
	"github.com/KKGo-Software-engineering/workshop-summer/api/health"
	"github.com/KKGo-Software-engineering/workshop-summer/api/mlog"
	"github.com/KKGo-Software-engineering/workshop-summer/api/spender"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	*echo.Echo
}

func New(db *sql.DB, cfg config.Config, logger *zap.Logger) *Server {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(mlog.Middleware(logger))

	v1 := e.Group("/api/v1")

	v1.GET("/slow", health.Slow)
	v1.GET("/health", health.Check(db))
	v1.POST("/upload", eslip.Upload)

	v1.Use(middleware.BasicAuth(AuthCheck))

	{
		h := spender.New(cfg.FeatureFlag, db)
		v1.GET("/spenders", h.GetAll)
		v1.POST("/spenders", h.Create)
	}

	{
		h := transaction.New(db)
		v1.POST("/transactions", h.Create)
		v1.GET("/spenders/:spenderId/transactions", h.GetAllBySpender)
		v1.PUT("/spenders/:spenderId/transactions/:transId", h.Update)
		v1.DELETE("/spenders/:spenderId/transactions/:transId", h.Delete)
	}

	{
		h := summary.New(cfg.FeatureFlag, db)
		v1.GET("/spenders/:id/expenses/summary", h.GetExpenseSummaryHandler)
		v1.GET("/spenders/:id/incomes/summary", h.GetIncomeSummaryHandler)
	}

	return &Server{e}
}
