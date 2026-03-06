package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		validAPIKey := os.Getenv("API_KEY")
		if validAPIKey == "" {
			log.Fatal("API_KEY no configurada .env")
		}
		if apiKey == "" {
			c.JSON(401, gin.H{"error": "API Key necesaria para uso"})
			c.Abort()
			return
		}
		if apiKey != validAPIKey {
			c.JSON(403, gin.H{"error": "API Key invalida"})
			c.Abort()
			return
		}

		c.Next()
	}
}
