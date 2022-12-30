package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Please use server.go for main file")
	fmt.Printf("start at port: %q\n", os.Getenv("PORT"))
	fmt.Printf("DB URL: %q\n", os.Getenv("DATABASE_URL"))

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":" + os.Getenv("PORT"))
}
