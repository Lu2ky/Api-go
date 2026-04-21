package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func insertarLog(usuarioID int, accion string, descripcion string) {
	var query string
	var args []interface{}

	if usuarioID == 0 {

		query = `
		INSERT INTO Logs (T_accion, T_Descripcion, Dt_fecha)
		VALUES (?, ?, NOW())
		`
		args = []interface{}{accion, descripcion}
	} else {

		query = `
		INSERT INTO Logs (N_idUsuario, T_accion, T_Descripcion, Dt_fecha)
		VALUES (?, ?, ?, NOW())
		`
		args = []interface{}{usuarioID, accion, descripcion}
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Println("Error al insertar log:", err)
	}
}

func insertLog(c *gin.Context) {
	var log Log

	err := c.BindJSON(&log)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	// Llamamos a la función que hace el INSERT
	insertarLog(log.N_idUsuario, log.Accion, log.Descripcion)

	c.JSON(200, gin.H{"status": "log insertado"})
}
