package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	err := r.Run(":8080")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
}
