package main

import (
	"log"
	"strconv"
	"github.com/gin-gonic/gin"
)

// --------------- Etiquetas ----------------------------------------

// TO DO: cambiar la consulta (AQUÍ SE SACAN TODAS LAS ETIQUETAS DE UN USUARIO)
func GetTagsByUserId(c *gin.Context) {

	//ID del usuario
	id := c.Param("id")

	rows, err := db.Query(`
		SELECT * FROM EtiquetasRecordatorios 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?)
		`, id)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	defer rows.Close()

	var TagsArray []Tags

	for rows.Next() {
		var Tags Tags

		err := rows.Scan(
			&Tags.N_idUsuario,
			&Tags.N_idRecordatorio,
			&Tags.N_idEtiqueta,
			&Tags.T_nombre,
			&Tags.B_isDeleted,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}

		TagsArray = append(TagsArray, Tags)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	c.JSON(200, TagsArray)
}

// TO DO: FUNCION PARA SACAR LAS ETIQUETAS DE UN RECORDATORIO POR SU NOMBRE
func GetTagsByUserIdAndReminderId(c *gin.Context) {

	//ID del usuario
	id := c.Param("id")

	//ID del recordatorio (se convierte en INT)
	reminderId, err := strconv.Atoi(c.Param("reminderId"))

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid reminder id"})
		return
	}

	rows, err := db.Query(`
		SELECT * FROM EtiquetasRecordatorios 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ? AND N_idRecordatorio = ?)
		`, id, reminderId)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	defer rows.Close()

	var TagsArray []Tags

	for rows.Next() {
		var Tags Tags

		err := rows.Scan(
			&Tags.N_idUsuario,
			&Tags.N_idRecordatorio,
			&Tags.N_idEtiqueta,
			&Tags.T_nombre,
			&Tags.B_isDeleted,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}

		TagsArray = append(TagsArray, Tags)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	c.JSON(200, TagsArray)
}

// TO DO: DELETE TAG
func deleteTag(c *gin.Context) {

	var delTag DelTag

	err := c.BindJSON(&delTag)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	result, err := db.Exec("CALL eliminar_etiqueta(?)", delTag.N_idEtiqueta)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	
	descripcion := "Se eliminó/recuperó etiqueta | ID: " +
		strconv.Itoa(delTag.N_idEtiqueta) +
		" | Usuario ID: " + strconv.Itoa(delTag.P_usuario)

	insertarLog(delTag.P_usuario, "DELETE_ETIQUETA", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Etiqueta alterada correctamente",
		"rowsAffected": rowsAffected,
	})
}
