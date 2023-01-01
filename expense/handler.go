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

func (h *handler) Create(c *gin.Context) {
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

func (h *handler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var expense Expense
	row := h.DB.QueryRow(`SELECT id, title, amount, note, tags FROM expenses WHERE id = $1`, id)

	if err := row.Scan(&expense.ID, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *handler) GetAll(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, errors.New("not implemented"))
}

func (h *handler) Update(c *gin.Context) {
	id := c.Param("id")
	var expense Expense
	if err := c.BindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err := h.DB.Prepare("UPDATE expenses SET title=$2, amount=$3, note=$4, tags=$5 WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := stmt.Exec(id, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	row := h.DB.QueryRow("SELECT id, title, amount, note, tags FROM expenses WHERE id=$1 ", id)
	if err := row.Scan(&expense.ID, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}
