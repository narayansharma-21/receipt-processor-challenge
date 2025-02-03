package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCountAlphanumeric(t *testing.T) {
	assert.Equal(t, 14, countAlphanumeric("M&M Corner Market"))
	assert.Equal(t, 10, countAlphanumeric("Best-Buy123"))
	assert.Equal(t, 0, countAlphanumeric("!!!***"))
}

func TestCheckRoundDollar(t *testing.T) {
	assert.Equal(t, 50, checkRoundDollar("10.00"))
	assert.Equal(t, 0, checkRoundDollar("10.99"))
}

func TestCheckMultipleOfQuarter(t *testing.T) {
	assert.Equal(t, 25, checkMultipleOfQuarter("10.25"))
	assert.Equal(t, 25, checkMultipleOfQuarter("5.50"))
	assert.Equal(t, 0, checkMultipleOfQuarter("7.33"))
}

func TestCountItems(t *testing.T) {
	assert.Equal(t, 5, countItems([]Item{{"A", "1.00"}, {"B", "2.00"}}))
	assert.Equal(t, 10, countItems([]Item{{"A", "1.00"}, {"B", "2.00"}, {"C", "3.00"}, {"D", "4.00"}}))
	assert.Equal(t, 0, countItems([]Item{{"A", "1.00"}}))
}

func TestCheckPurchaseDate(t *testing.T) {
	assert.Equal(t, 6, checkPurchaseDate("2023-07-15")) // Odd day
	assert.Equal(t, 0, checkPurchaseDate("2023-07-16")) // Even day
}

func TestCheckPurchaseTime(t *testing.T) {
	assert.Equal(t, 10, checkPurchaseTime("14:30"))
	assert.Equal(t, 0, checkPurchaseTime("13:59"))
	assert.Equal(t, 0, checkPurchaseTime("16:00"))
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/receipts/process", processReceipt)
	r.GET("/receipts/:id/points", getPoints)
	return r
}

func TestProcessReceipt(t *testing.T) {
	r := setupRouter()

	reqBody := `{"retailer":"M&M Corner Market","purchaseDate":"2023-07-15","purchaseTime":"14:30","items":[{"shortDescription":"Milk","price":"3.00"}],"total":"6.00"}`
	req := httptest.NewRequest("POST", "/receipts/process", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.ID)
}

func TestGetPoints(t *testing.T) {
	r := setupRouter()
	// Create a receipt first
	reqBody := `{"retailer":"M&M Corner Market","purchaseDate":"2023-07-15","purchaseTime":"14:30","items":[{"shortDescription":"Milk","price":"3.00"}],"total":"6.00"}`
	req := httptest.NewRequest("POST", "/receipts/process", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Now test GET /receipts/:id/points
	req = httptest.NewRequest("GET", "/receipts/"+resp.ID+"/points", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
