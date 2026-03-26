package main

import (
	"log"
	"github.com/gin-gonic/gin"
)

// --------- COMENTARIOS -----------------------
func getPersonalCommentsByUserIdAndCourseId(c *gin.Context) {

	id_User := c.Param("id")
	id_course := c.Param("idCourse")
	rows, err := db.Query(`SELECT * FROM ComentariosOficiales 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?)
		AND N_idCurso = ?`, id_User, id_course)

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

	c.JSON(200, ofcCommentsArray)
}

func getPersonalCommentsByUserId(c *gin.Context) {
	id_User := c.Param("id")
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
	//insertedID, _ := result.LastInsertId()
	

	 //descripcion := "El id del comentario ingresado fue: " + strconv.FormatInt(insertedID, 10)

	 //insertarLog(
	 //	newComment.N_idUsuario, 
	 //	"INSERTAR_COMENTARIO",
	 //	descripcion,
	 //)

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
	 //descripcion := "Se actualizó el comentario con id: " + strconv.Itoa(newComment.N_idComentarios)

	//	insertarLog(
	//		newComment.N_idUsuario, 
	//		"ACTUALIZAR_COMENTARIO",
	//		descripcion,
	//	)
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

	result, err := db.Exec(
		"CALL eliminar_comentario(?)",
		delComment.N_idComentarios,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	
	insertarLog(
		delComment.N_idUsuario, 
	 	"ELIMINAR_COMENTARIO",
		descripcion,
	  )

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}
