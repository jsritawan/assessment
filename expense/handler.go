package expense

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *handler {
	return &handler{
		DB: db,
	}
}

func (h *handler) CreateExpense(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, errors.New("not implemented yet"))
}
