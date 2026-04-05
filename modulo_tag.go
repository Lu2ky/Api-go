package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// --------------- Etiquetas ----------------------------------------

func GetTagsByUserId(c *gin.Context) {

	//ID del usuario
	id := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "TagsByUser:"+id).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var tagsArray []Tags

		err := json.Unmarshal([]byte(val), &tagsArray)

		if err == nil {
			c.JSON(200, tagsArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

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

	// Convertir a formato apto para redis
	data, err := json.Marshal(TagsArray)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err2 := rdb.Set(ctx, "TagsByUser:"+id, data, 48*time.Hour).Err()

	if err2 != nil {
		log.Printf("Error al guardar en Redis: %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, TagsArray)
}

// FUNCION PARA SACAR LAS ETIQUETAS DE UN RECORDATORIO POR SU NOMBRE
func GetTagsByUserIdAndReminderId(c *gin.Context) {

	//ID del usuario
	id := c.Param("id")

	//ID del recordatorio (se convierte en INT)
	reminderId, err := strconv.Atoi(c.Param("reminderId"))

	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid reminder id"})
		return
	}

	//	Consulta a redis
	key := fmt.Sprintf("TagsByUser&Reminder:%s-%d", id, reminderId)
	val, err := rdb.Get(c.Request.Context(), key).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var tagsArray []Tags

		err := json.Unmarshal([]byte(val), &tagsArray)

		if err == nil {
			c.JSON(200, tagsArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

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

	// Convertir a formato apto para redis
	data, err := json.Marshal(TagsArray)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err2 := rdb.Set(ctx, key, data, 48*time.Hour).Err()

	if err2 != nil {
		log.Printf("Error al guardar en Redis: %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, TagsArray)
}

// DELETE TAG
func deleteTag(c *gin.Context) {

	var delTag DelTag

	err := c.BindJSON(&delTag)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	// Borrar registro de datos de usuario de redis
	deleted, err2 := rdb.Del(ctx, "TagsByUser:"+*delTag.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Llamado al procedimiento
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
