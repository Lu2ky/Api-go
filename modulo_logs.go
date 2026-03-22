package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)


func addLog(c *gin.Context) {
	var body struct {
		UsuarioID int    `json:"usuario_id"`
		Accion    string `json:"accion"`
		Descripcion string `json:"descripcion"`
	}

	// Leer JSON
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Datos inválidos"})
		return
	}

	// Ejecutar query
	_, err := db.Exec(`
		INSERT INTO Logs (usuario_id, accion, descripcion)
		VALUES (?, ?, ?)
	`, body.UsuarioID, body.Accion, body.Descripcion)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
