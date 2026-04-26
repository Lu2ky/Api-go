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

//	------------------------ NOTIFICACIONES Y CORREO  ------------------------ //

func GetNotificaciones(c *gin.Context) {

	id_user := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "Notifications:"+id_user).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var notiArray []Notificacion
		err := json.Unmarshal([]byte(val), &notiArray)

		if err == nil {
			c.JSON(200, notiArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

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

	// Devuelve la consulta de la base relacional
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

	// Borrar registro de notificaciones de redis
	deleted, err2 := rdb.Del(ctx, "Notifications:"+*notiNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	if notiNewValue.CodUsuario != nil {
		rdb.Del(ctx, "Notifications:"+*notiNewValue.CodUsuario)
	}

	fmt.Printf("%#v\n", notiNewValue)

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("INSERT INTO Notificaciones (T_nombre, T_descripcion, Dt_fechaEmision, N_idToDoList)  VALUES (?, ?, ?, ?)",
		notiNewValue.T_nombre,
		notiNewValue.T_descripcion,
		notiNewValue.Dt_fechaEmision,
		notiNewValue.N_idToDoList,
	)

	if err != nil {
		log.Printf("Database error: %v", err)

		c.JSON(500, gin.H{
			"error":           "Error interno en la base de datos",
			"mensaje_mysql":   err.Error(),
			"datos_recibidos": notiNewValue,
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Reminder not found"})
		return
	}
	insertedID, _ := result.LastInsertId()

	descripcion := "Se creó notificación | ID: " +
		strconv.FormatInt(insertedID, 10) +
		" | Usuario ID: " + strconv.Itoa(notiNewValue.N_idUsuario) +
		" | Nombre: " + notiNewValue.T_nombre

	insertarLog(
		notiNewValue.N_idUsuario,
		"CREAR_NOTIFICACION",
		descripcion,
	)

	c.JSON(200, gin.H{
		"message": "Notificación creada correctamente",
		"id":      insertedID,
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

	// Borrar registro de notificaciones de redis
	deleted, err2 := rdb.Del(ctx, "Notifications:"+*idsNotifications.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Llamado al procedimiento
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

	// Log
	var userId int
	err4 := db.QueryRow("CALL get_id_tabla(?)", *idsNotifications.CodUsuario).Scan(&userId)
	if err4 != nil {
		log.Printf("Error obteniendo ID: %v", err)
	}

	descripcion := fmt.Sprintf("Se eliminaron los recordatorios | IDs: %s | Usuario ID: %d",
		idsNotifications.Ids, userId)

	go func(uID string, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertLogCod(uID, acc, desc)
	}(*idsNotifications.CodUsuario, "ELIMINAR_NOTIFICACIONES", descripcion)

	c.JSON(200, gin.H{
		"message": "Notificaciones eliminadas correctamente",
	})

}

func muteNotification(c *gin.Context) {

	var notiNewValue MuteNotification

	// Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&notiNewValue)

	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}
	if !AuthorityCheck(*notiNewValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de datos de usuario de redis
	deleted, err2 := rdb.Del(ctx, "UserInfo:"+*notiNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Aquí se hace el llamado al Procedimiento
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

	var correo string
	var antelacion string

	if notiNewValue.P_correo != nil {
		correo = *notiNewValue.P_correo
	} else {
		correo = "Sin cambios"
	}

	if notiNewValue.P_antelacionNotis != nil {
		antelacion = *notiNewValue.P_antelacionNotis
	} else {
		antelacion = "Sin cambios"
	}

	descripcion := "\nConfiguración de notificaciones actualizada | Usuario ID: " +
		strconv.Itoa(notiNewValue.P_idUsuario) +
		" | Correo: " + correo +
		" | Antelación: " + antelacion

	fmt.Println(descripcion)

	insertarLog(
		notiNewValue.P_idUsuario,
		"CONFIGURAR_NOTIFICACIONES",
		descripcion,
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

	insertedID, _ := result.LastInsertId()

	descripcion := "Correo creado | ID: " +
		strconv.FormatInt(insertedID, 10) +
		" | Usuario ID: " + strconv.Itoa(correoNewValue.N_idUsuario) +
		" | Asunto: " + correoNewValue.T_asunto

	insertarLog(
		correoNewValue.N_idUsuario,
		"CREAR_CORREO",
		descripcion,
	)

	c.JSON(200, gin.H{
		"message": "Correo creado correctamente",
	})
}
