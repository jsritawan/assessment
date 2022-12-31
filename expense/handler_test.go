//go:build unit

package expense

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateExpense(t *testing.T) {
	// Arrange
	body := Expense{
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Error("Error marshalling", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO expenses").
		WithArgs(body.Title, body.Amount, body.Note, pq.Array(&body.Tags)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	h := NewHandler(db)
	r := gin.Default()
	r.POST("/expenses", h.CreateExpense)
	b, _ := json.Marshal(Expense{
		ID:     1,
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	})
	expect := string(b)

	// Act
	r.ServeHTTP(rec, req)

	// Assert
	err = mock.ExpectationsWereMet()
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, expect, strings.TrimSpace(rec.Body.String()))
	}
}
