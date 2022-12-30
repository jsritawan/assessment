package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Please use server.go for main file")
	fmt.Printf("start at port: %q\n", os.Getenv("PORT"))
	fmt.Printf("DB URL: %q\n", os.Getenv("DATABASE_URL"))
}
