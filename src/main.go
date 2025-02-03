package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Response struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

var (
	receipts = make(map[string]Receipt)
	mu       sync.Mutex
)

func main() {
	r := gin.Default()
	r.POST("/receipts/process", processReceipt)
	r.GET("/receipts/:id/points", getPoints)
	r.Run(":8080")
}

func processReceipt(c *gin.Context) {
	var receipt Receipt
	if err := c.ShouldBindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	id := uuid.New().String()
	mu.Lock()
	receipts[id] = receipt
	mu.Unlock()

	c.JSON(http.StatusOK, Response{ID: id})
}

func getPoints(c *gin.Context) {
	id := c.Param("id")

	mu.Lock()
	receipt, exists := receipts[id]
	mu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	points := calculatePoints(receipt)
	c.JSON(http.StatusOK, PointsResponse{Points: points})
}

func calculatePoints(receipt Receipt) int {
	points := 0
	points += countAlphanumeric(receipt.Retailer)
	points += checkRoundDollar(receipt.Total)
	points += checkMultipleOfQuarter(receipt.Total)
	points += countItems(receipt.Items)
	points += checkItemDescriptions(receipt.Items)
	points += checkPurchaseDate(receipt.PurchaseDate)
	points += checkPurchaseTime(receipt.PurchaseTime)
	return points
}

func countAlphanumeric(s string) int {
	count := 0
	for _, char := range s {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			count++
		}
	}
	return count
}

func checkRoundDollar(total string) int {
	if strings.HasSuffix(total, ".00") {
		return 50
	}
	return 0
}

func checkMultipleOfQuarter(total string) int {
	val, err := strconv.ParseFloat(total, 64)
	if err == nil && math.Mod(val, 0.25) == 0 {
		return 25
	}
	return 0
}

func countItems(items []Item) int {
	return (len(items) / 2) * 5
}

func checkItemDescriptions(items []Item) int {
	points := 0
	for _, item := range items {
		desc := strings.TrimSpace(item.ShortDescription)
		if len(desc)%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err == nil {
				points += int(math.Ceil(price * 0.2))
			}
		}
	}
	return points
}

func checkPurchaseDate(date string) int {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err == nil && parsedDate.Day()%2 == 1 {
		return 6
	}
	return 0
}

func checkPurchaseTime(timeStr string) int {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err == nil && parsedTime.Hour() >= 14 && parsedTime.Hour() < 16 {
		return 10
	}
	return 0
}
