package expense

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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
	var expense Expense
	if err := c.BindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := h.DB.QueryRow(`
		INSERT INTO expenses(title, amount, note, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`,
		expense.Title,
		expense.Amount,
		expense.Note,
		pq.Array(&expense.Tags))

	if err := row.Scan(&expense.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func (h *handler) Update(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, errors.New("not implemented yet"))
}
