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

// --------- COMENTARIOS -----------------------
func getPersonalCommentsByUserIdAndCourseId(c *gin.Context) {

	id_User := c.Param("id")
	id_course := c.Param("idCourse")

	//	Consulta a redis
	key := fmt.Sprintf("PersonalComments:%s-%s", id_User, id_course)
	val, err := rdb.Get(c.Request.Context(), key).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var ofcCommentsArray []ofcComments

		err := json.Unmarshal([]byte(val), &ofcCommentsArray)

		if err == nil {
			c.JSON(200, ofcCommentsArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

	rows, err := db.Query(`SELECT * FROM ComentariosOficiales 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?)
		AND N_idHorario = ?`, id_User, id_course)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()
	var ofcCommentsArray []ofcComments
	for rows.Next() {
		var ofcComment ofcComments
		err := rows.Scan(
			&ofcComment.N_idHorario,
			&ofcComment.N_idUsuario,
			&ofcComment.N_idCurso,
			&ofcComment.Curso,
			&ofcComment.N_idComentarios,
			&ofcComment.T_comentario,
			&ofcComment.B_isDeleted,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		ofcCommentsArray = append(ofcCommentsArray, ofcComment)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	// Convertir a formato apto para redis
	data, err := json.Marshal(ofcCommentsArray)
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
	c.JSON(200, ofcCommentsArray)
}

func getPersonalCommentsByUserId(c *gin.Context) {
	id_User := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "PersonalComments:"+id_User).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var ofcCommentsArray []ofcComments

		err := json.Unmarshal([]byte(val), &ofcCommentsArray)

		if err == nil {
			c.JSON(200, ofcCommentsArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

	rows, err := db.Query(`SELECT * FROM ComentariosOficiales WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ? )`, id_User)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()
	var ofcCommentsArray []ofcComments
	for rows.Next() {
		var ofcComment ofcComments
		err := rows.Scan(
			&ofcComment.N_idHorario,
			&ofcComment.N_idUsuario,
			&ofcComment.N_idCurso,
			&ofcComment.Curso,
			&ofcComment.N_idComentarios,
			&ofcComment.T_comentario,
			&ofcComment.B_isDeleted,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		ofcCommentsArray = append(ofcCommentsArray, ofcComment)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	// Convertir a formato apto para redis
	data, err := json.Marshal(ofcCommentsArray)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err2 := rdb.Set(ctx, "PersonalComments:"+id_User, data, 48*time.Hour).Err()

	if err2 != nil {
		log.Printf("Error al guardar en Redis: %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, ofcCommentsArray)
}

// Insertar comentario personal en actividad oficial
func addPersonalComment(c *gin.Context) {
	var newComment new_ofcComments
	err := c.BindJSON(&newComment)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalComments:"+*newComment.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	key := fmt.Sprintf("PersonalComments:%s-%v", *newComment.CodUsuario, newComment.N_idCurso)
	deleted, err3 := rdb.Del(ctx, key).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Se llama el insert
	result, err := db.Exec(
		"INSERT INTO Comentarios (N_idHorario, T_Comentario) VALUES (?, ?)",
		newComment.N_idHorario,
		newComment.T_comentario,
	)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		log.Printf("LastInsertId error: %v", err)
	}

	// log
	descripcion := "Comentario creado | ID: " +
		strconv.FormatInt(insertedID, 10) +
		" | Horario: " + strconv.Itoa(newComment.N_idHorario)

	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic en insertarLog:", r)
			}
		}()
		insertarLog(newComment.N_idUsuario, "CREAR_COMENTARIO", descripcion)
	}()

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario agregado correctamente",
		"rowsAffected": rowsAffected,
	})

}

// Procedimiento: actualizar comentario TODO //
func updatePersonalComment(c *gin.Context) {

	var newComment edit_ofcComment

	err := c.BindJSON(&newComment)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalComments:"+*newComment.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	key := fmt.Sprintf("PersonalComments:%s-%v", *newComment.CodUsuario, newComment.N_idCurso)
	deleted, err3 := rdb.Del(ctx, key).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegsitro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	result, err := db.Exec(
		"CALL editar_comentario(? , ?)",
		newComment.N_idComentarios,
		newComment.T_comentario,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Log
	descripcion := fmt.Sprintf("Comentario actualizado | ID: %d | Usuario ID: %d",
		newComment.N_idComentarios,
		newComment.N_idUsuario)

	go func(uID int, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Editar): %v", r)
			}
		}()
		insertarLog(uID, acc, desc)
	}(newComment.N_idUsuario, "ACTUALIZAR_COMENTARIO", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario editado correctamente",
		"rowsAffected": rowsAffected,
	})
}

// Eliminar comentario TODO //
func deletePersonalComment(c *gin.Context) {

	var delComment del_ofcComment

	err := c.BindJSON(&delComment)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalComments:"+*delComment.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	key := fmt.Sprintf("PersonalComments:%s-%v", *delComment.CodUsuario, delComment.N_idCurso)
	deleted, err3 := rdb.Del(ctx, key).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegsitro eliminado con éxito")
	} else {
		fmt.Printf("\nNo es encontró registro relacionado")
	}

	result, err := db.Exec(
		"CALL eliminar_comentario(?)",
		delComment.N_idComentarios,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Log
	descripcion := fmt.Sprintf("Comentario eliminado | ID: %d | Usuario ID: %d",
		delComment.N_idComentarios,
		delComment.N_idUsuario)

	go func(uID int, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertarLog(uID, acc, desc)
	}(delComment.N_idUsuario, "ELIMINAR_COMENTARIO", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}
