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
	t.Run("Create Expense With Invalid Request Shoud Return Bad Request", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader("invalid-request"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.POST("/expenses", h.Create)

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Create Expense Should Return OK", func(t *testing.T) {

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

		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectQuery(`
		INSERT INTO expenses(title, amount, note, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id`).
			WithArgs(body.Title, body.Amount, body.Note, pq.Array(&body.Tags)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.POST("/expenses", h.Create)
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
	})
}

func TestGetExpenseDetailById(t *testing.T) {
	t.Run("Get Expense Detail By Invalid Id Should Return Bad Request", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/expenses/invalid-id", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.GET("/expenses/:id", h.Get)

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Get Expense Detail By Id Should Return OK", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/expenses/1", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectQuery("SELECT (.+) FROM expenses").
			WithArgs("1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
				AddRow(1, "strawberry smoothie", 79, "night market promotion discount 10 bath", pq.Array(&[]string{"food", "beverage"})))

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.GET("/expenses/:id", h.Get)
		expect := "{\"id\":1,\"title\":\"strawberry smoothie\",\"amount\":79,\"note\":\"night market promotion discount 10 bath\",\"tags\":[\"food\",\"beverage\"]}"

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expect, strings.TrimSpace(rec.Body.String()))
	})
}

func TestGetAllExpenses(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT (.+) FROM expenses").
		ExpectQuery().
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
			AddRow(1, "strawberry smoothie", 79, "night market promotion discount 10 bath", pq.Array(&[]string{"food", "beverage"})).
			AddRow(2, "apple smoothie", 89, "no discount", pq.Array(&[]string{"beverage"})))

	gin.SetMode(gin.TestMode)
	h := NewHandler(db)
	r := gin.Default()
	r.GET("/expenses", h.GetAll)
	expect := []Expense{
		{
			ID:     1,
			Title:  "strawberry smoothie",
			Amount: 79,
			Note:   "night market promotion discount 10 bath",
			Tags:   []string{"food", "beverage"},
		},
		{
			ID:     2,
			Title:  "apple smoothie",
			Amount: 89,
			Note:   "no discount",
			Tags:   []string{"beverage"},
		},
	}
	expectBytes, err := json.Marshal(expect)
	if err != nil {
		t.Errorf("an error '%s' was not expected when marshalling expenses", err)
	}

	// Act
	r.ServeHTTP(rec, req)

	// Assert
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(expectBytes), strings.TrimSpace(rec.Body.String()))
}

func TestUpdateExpense(t *testing.T) {
	t.Run("Update Expense Detail By Invalid Id Should Return Bad Request", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodPut, "/expenses/invalid-id", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.PUT("/expenses/:id", h.Update)

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Update Expense With Invalid Request Shoud Return Bad Request", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodPut, "/expenses/1", strings.NewReader("invalid-request"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, _, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.PUT("/expenses/:id", h.Update)

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Update Expense Should Return OK", func(t *testing.T) {

		// Arrange
		body := `
		{
			"title": "apple smoothie",
			"amount": 89,
			"note": "no discount",
			"tags": ["beverage"]
		}`
		req := httptest.NewRequest(http.MethodPut, "/expenses/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		mock.ExpectPrepare("UPDATE expenses").
			ExpectExec().
			WithArgs("1", "apple smoothie", 89.0, "no discount", pq.Array([]string{"beverage"})).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM expenses").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
				AddRow("1", "apple smoothie", 89.0, "no discount", pq.Array([]string{"beverage"})))

		gin.SetMode(gin.TestMode)
		h := NewHandler(db)
		r := gin.Default()
		r.PUT("/expenses/:id", h.Update)

		expect := `{"id":1,"title":"apple smoothie","amount":89,"note":"no discount","tags":["beverage"]}`

		// Act
		r.ServeHTTP(rec, req)

		// Assert
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expect, strings.TrimSpace(rec.Body.String()))
	})
}
