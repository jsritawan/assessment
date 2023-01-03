//go:build integration

package expense

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const serverPort = 80

func setupIT(t *testing.T) func() {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	db, err := sql.Open("postgres", "postgresql://root:root@db/go-assessment-db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	h := NewHandler(db)
	r.POST("/expenses", h.Create)
	r.GET("/expenses/:id", h.Get)
	r.GET("/expenses", h.GetAll)
	r.PUT("/expenses/:id", h.Update)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: r,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", serverPort), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = srv.Shutdown(ctx)
		assert.NoError(t, err)
	}
}

func TestITCreateExpense(t *testing.T) {
	// Setup server
	teardown := setupIT(t)
	defer teardown()

	// Arrange
	reqBody := `{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath", 
		"tags": ["food", "beverage"]
	}`
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/expenses", serverPort), strings.NewReader(reqBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}

	// Act
	resp, err := client.Do(req)
	assert.NoError(t, err)

	byteBody, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	// Assertion
	expect := Expense{
		ID:     1,
		Title:  "strawberry smoothie",
		Amount: 79,
		Note:   "night market promotion discount 10 bath",
		Tags:   []string{"food", "beverage"},
	}
	byteExpect, err := json.Marshal(expect)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, byteExpect, byteBody)
	}
}

func TestITGetExpenseById(t *testing.T) {
	// Setup server
	teardown := setupIT(t)
	defer teardown()

	// Arrange
	reqCreateBody := `{
		"title": "strawberry smoothie",
		"amount": 79,
		"note": "night market promotion discount 10 bath", 
		"tags": ["food", "beverage"]
	}`
	reqCreate, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/expenses", serverPort), strings.NewReader(reqCreateBody))
	assert.NoError(t, err)
	reqCreate.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	// Act
	respCreate, err := client.Do(reqCreate)
	assert.NoError(t, err)
	byteCreateBody, err := ioutil.ReadAll(respCreate.Body)
	assert.NoError(t, err)
	respCreate.Body.Close()

	var createdExpense Expense
	err = json.Unmarshal(byteCreateBody, &createdExpense)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, respCreate.StatusCode)
	}

	reqGetBody := ``
	reqGet, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/expenses/%d", serverPort, createdExpense.ID), strings.NewReader(reqGetBody))
	assert.NoError(t, err)
	reqGet.Header.Set("Content-Type", "application/json")

	respGet, err := client.Do(reqGet)
	assert.NoError(t, err)
	byteGetBody, err := ioutil.ReadAll(respGet.Body)
	assert.NoError(t, err)
	respCreate.Body.Close()

	// Assertion
	expect := createdExpense
	byteExpect, err := json.Marshal(expect)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, respGet.StatusCode)
		assert.Equal(t, byteExpect, byteGetBody)
	}
}
