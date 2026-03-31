package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

//	------------------------ NOTIFICACIONES Y CORREO  ------------------------ //

func GetNotificaciones(c *gin.Context) {

	id_user := c.Param("id")

	//	Consulta
	rows, err := db.Query(
		`
		SELECT * FROM campanitaNotis 
		WHERE N_idUsuario= (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?);
		`,
		id_user,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var notiArray []Notificacion

	//	Escanear y guardar la información de la consulta
	for rows.Next() {
		var noti Notificacion
		err := rows.Scan(
			&noti.N_idNotificacion,
			&noti.N_idUsuario,
			&noti.N_idRecordatorio,
			&noti.T_nombre,
			&noti.T_descripcion,
			&noti.Dt_fechaEmision,
			&noti.B_estado,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		notiArray = append(notiArray, noti)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	c.JSON(200, notiArray)
}

func addNotificacion(c *gin.Context) {

	var notiNewValue NewNotificacion

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&notiNewValue)

	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("INSERT INTO Notificaciones (T_nombre, T_descripcion, Dt_fechaEmision, N_idToDoList) VALUES(?, ?, ?, ?)",
		notiNewValue.T_nombre,
		notiNewValue.T_descripcion,
		notiNewValue.Dt_fechaEmision,
		notiNewValue.N_idToDoList,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Reminder not found"})
		return
	}
	//insertedID, _ := result.LastInsertId()

	//descripcion := "Se creó una notificación con id: " + strconv.FormatInt(insertedID, 10)

	// insertarLog(
	//	notiNewValue.N_idUsuario,
	//	"CREAR_NOTIFICACION",
	//	descripcion,
	// )
	c.JSON(200, gin.H{
		"message": "Notificacion creada correctamente",
	})
}

func deleteNotifications(c *gin.Context) {

	var idsNotifications DeleteNotification

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&idsNotifications)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	result, err := db.Exec("CALL leer_noti(?)",
		idsNotifications.Ids,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Notificaciones no halladas"})
		return
	}

	//descripcion := "Se actualizó elimaron los recordatorios con IDs: " + strconv.Itoa(ids.Ids)

	//insertarLog(reminderNewValue.P_idToDo, "UPDATE_RECORDATORIO", descripcion)
	c.JSON(200, gin.H{
		"message": "Notificaciones eliminadas correctamente",
	})

}

func muteNotification(c *gin.Context) {

	var notiNewValue MuteNotification

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&notiNewValue)

	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("CALL configuracion_notificaciones(?, ?, ?);",
		notiNewValue.P_idUsuario,
		notiNewValue.P_correo,
		notiNewValue.P_antelacionNotis,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	insertarLog(
		notiNewValue.P_idUsuario,
		"CONFIGURAR_NOTIFICACIONES",
		"El usuario modificó configuración de notificaciones",
	)
	if rowsAffected == 0 {
		c.JSON(200, gin.H{"message": "No hubo cambios"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Preferencias actualizadas",
	})
}

func addCorreo(c *gin.Context) {
	var correoNewValue NewCorreo

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&correoNewValue)

	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("INSERT INTO Correos (T_asunto, T_contenido, Dt_fechaEmision, N_idToDoList) VALUES (?, ?, ?, ?)",
		correoNewValue.T_asunto,
		correoNewValue.T_contenido,
		correoNewValue.Dt_fechaEmision,
		correoNewValue.N_idToDoList,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Reminder not found"})
		return
	}

	//insertedID, _ := result.LastInsertId()
	//descripcion := "Se creó un correo con id: " + strconv.FormatInt(insertedID, 10)
	//insertarLog(
	//	correoNewValue.N_idUsuario,
	//	"CREAR_CORREO",
	//	descripcion,
	//)
	c.JSON(200, gin.H{
		"message": "Correo creado correctamente",
	})
}
