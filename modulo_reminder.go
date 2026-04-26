package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"github.com/gin-gonic/gin"
)

//	--------------- Recordatorios ----------------------------------------

// Obtener la lista de los recordatorios y sus etiquetas
func GetRemindersTagsByUserId(c *gin.Context) {

	//	Id del usuario
	id_User := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "Reminder&Tags:"+id_User).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var remindersArray []RemindersTag

		err := json.Unmarshal([]byte(val), &remindersArray)

		if err == nil {
			c.JSON(200, remindersArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

	//	Consulta
	rows, err := db.Query(
		`
		SELECT * FROM RecordatoriosCompletos WHERE N_idUsuario=(SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario= ?)
		`,
		id_User,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var remindersArray []RemindersTag

	//	Escanear y guardar la información de la consulta
	for rows.Next() {
		var reminder RemindersTag
		err := rows.Scan(
			&reminder.N_idToDoList,
			&reminder.N_idUsuario,
			&reminder.N_idRecordatorio,
			&reminder.T_nombre,
			&reminder.T_descripcion,
			&reminder.Dt_fechaVencimiento,
			&reminder.B_isDeleted,
			&reminder.T_Prioridad,
			&reminder.B_estado,
			&reminder.N_idEtiqueta,
			&reminder.T_tag_nombre,
			&reminder.B_tag_isDeleted,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		remindersArray = append(remindersArray, reminder)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, remindersArray)
}

// Obtener la lista de los recordatorios
func GetRemindersByUserId(c *gin.Context) {

	/*
		type Reminders struct{
			N_idToDoList		int			`json:"N_idToDoList"`
			N_idUsuario			int			`json:"N_idUsuario"`
			N_idRecordatorio	int			`json:"N_idRecordatorio"`
			T_nombre			string		`json:"T_nombre"`
			T_descripción		string		`json:"T_descripción"`
			Dt_fechaVencimiento	string		`json:"Dt_fechaVencimiento"`
			B_isDeleted			*bool		`json:"B_isDeleted"`
			T_Prioridad			string		`json:"T_Prioridad"`
		}
	*/

	//	Id del usuario
	id_User := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "Reminders:"+id_User).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var remindersArray []Reminders

		err := json.Unmarshal([]byte(val), &remindersArray)

		if err == nil {
			c.JSON(200, remindersArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

	//	Consulta
	rows, err := db.Query(
		`
		SELECT * FROM RecordatoriosUsuarios 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?)
		`,
		id_User,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var remindersArray []Reminders

	//	Escanear y guardar la información de la consulta
	for rows.Next() {
		var reminder Reminders
		err := rows.Scan(
			&reminder.N_idToDoList,
			&reminder.N_idUsuario,
			&reminder.N_idRecordatorio,
			&reminder.T_nombre,
			&reminder.T_descripcion,
			&reminder.Dt_fechaVencimiento,
			&reminder.B_isDeleted,
			&reminder.T_Prioridad,
			&reminder.B_estado,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		remindersArray = append(remindersArray, reminder)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, remindersArray)
}

// Procedimiento crear recordatorio
func addReminder(c *gin.Context) {
	var reminderNewValue ReminderNewValue

	// Se asignan los valores del JSON a la estructura reminderNewValue
	err := c.BindJSON(&reminderNewValue)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	if !AuthorityCheck(*reminderNewValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "Reminders:"+*reminderNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	deleted, err3 := rdb.Del(ctx, "Reminder&Tags:"+*reminderNewValue.CodUsuario).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Borrar registro de etiquetas de usuario de redis
	deleted, err4 := rdb.Del(ctx, "TagsByUser:"+*reminderNewValue.CodUsuario).Result()

	if err4 != nil {
		fmt.Printf("\nError de conexión: %v", err4)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Variables para salida
	var toDoId int64
	var reminderId int64

	// Aquí se hace el llamado al Procedimiento
	err5 := db.QueryRow("SELECT crear_recordatorio_5tags(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		reminderNewValue.P_usuario,
		reminderNewValue.P_nombre,
		reminderNewValue.P_descripcion,
		reminderNewValue.P_fecha,
		reminderNewValue.P_prioridad,
		reminderNewValue.P_tag1,
		reminderNewValue.P_tag2,
		reminderNewValue.P_tag3,
		reminderNewValue.P_tag4,
		reminderNewValue.P_tag5,
	).Scan(&toDoId)

	if err5 != nil {
		log.Printf("Error ejecutando o leyendo resultado: %v", err5)
		c.JSON(500, gin.H{"error": "Error al crear"})
		return
	}

	// Consulta el id toDo del recordatorio
	err6 := db.QueryRow("SELECT N_idRecordatorio FROM ToDoList WHERE N_idToDoList = ?",
		toDoId,
	).Scan(&reminderId)

	if err6 != nil {
		log.Printf("Error ejecutando o leyendo resultado: %v", err5)
		c.JSON(500, gin.H{"error": "Error al consultar el id"})
		return
	}

	// Log

	log.Printf("ID del ToDo creado: %d", reminderId)
	descripcion := "Se creó recordatorio ID: " + strconv.FormatInt(reminderId, 10) +
		" | Usuario: " + strconv.Itoa(reminderNewValue.P_usuario) +
		" | Nombre: " + reminderNewValue.P_nombre

	insertarLog(reminderNewValue.P_usuario, "CREAR_RECORDATORIO", descripcion)

	// Salida
	c.JSON(200, gin.H{
		"message":    "Recordatorio creado correctamente",
		"toDoId":     toDoId,
		"reminderId": reminderId,
	})
}

// Procedimiento: Actualizar recordatorio
func updateReminderById(c *gin.Context) {

	var reminderNewValue EditReminder

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&reminderNewValue)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}
	if !AuthorityCheck(*reminderNewValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "Reminders:"+*reminderNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	deleted, err3 := rdb.Del(ctx, "Reminder&Tags:"+*reminderNewValue.CodUsuario).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Borrar registro de etiquetas de usuario de redis
	deleted, err4 := rdb.Del(ctx, "TagsByUser:"+*reminderNewValue.CodUsuario).Result()

	if err4 != nil {
		fmt.Printf("\nError de conexión: %v", err4)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("CALL editar_recordatorio_5tags(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		reminderNewValue.P_idToDo,
		reminderNewValue.P_nombre,
		reminderNewValue.P_descripcion,
		reminderNewValue.P_fecha,
		reminderNewValue.P_prioridad,
		reminderNewValue.P_estado,
		reminderNewValue.P_tag1,
		reminderNewValue.P_tag2,
		reminderNewValue.P_tag3,
		reminderNewValue.P_tag4,
		reminderNewValue.P_tag5,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Recordatorio no encontrado"})
		return
	}

	// Consulta el id toDo del recordatorio
	var toDoId = reminderNewValue.P_idToDo
	var reminderId int64

	err5 := db.QueryRow("SELECT N_idRecordatorio FROM ToDoList WHERE N_idToDoList = ?",
		toDoId,
	).Scan(&reminderId)

	if err5 != nil {
		log.Printf("Error ejecutando o leyendo resultado: %v", err5)
		c.JSON(500, gin.H{"error": "Error al consultar el id"})
		return
	}

	// Log
	descripcion := fmt.Sprintf("Se actualizó recordatorio | ID_TO_DO: %d | Usuario ID: %d",
		reminderNewValue.P_idToDo, reminderNewValue.P_usuario)

	go func(uID int, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertarLog(uID, acc, desc)
	}(reminderNewValue.P_usuario, "UPDATE_RECORDATORIO", descripcion)

	// Salida
	c.JSON(200, gin.H{
		"message":    "Recordatorio actualizado correctamente",
		"reminderId": reminderId,
	})
}

// Procedimiento: Eliminar recordatorio

func deleteOrRecoverReminder(c *gin.Context) {

	var delReminder DelReminder

	err := c.BindJSON(&delReminder)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	if delReminder.P_usuario == 0 {
		c.JSON(400, gin.H{"error": "usuario requerido"})
		return
	}
	if !AuthorityCheck(*delReminder.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "Reminders:"+*delReminder.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	deleted, err3 := rdb.Del(ctx, "Reminder&Tags:"+*delReminder.CodUsuario).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegsitro eliminado con éxito")
	} else {
		fmt.Printf("\nNo es encontró registro relacionado")
	}

	// Llamado al procedimiento
	result, err := db.Exec("CALL eliminar_recordatorio(?)", delReminder.N_idRecordatorio)
	if err != nil {
		log.Printf(" Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	descripcion := "Se eliminó recordatorio ID: " +
		strconv.Itoa(delReminder.N_idRecordatorio) +
		" | Usuario: " + strconv.Itoa(delReminder.P_usuario)

	insertarLog(delReminder.P_usuario, "ELIMINAR_RECORDATORIO", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Recordatorio alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}

func deleteMultipleReminder(c *gin.Context) {

	var delReminder MultiDelReminder

	err := c.BindJSON(&delReminder)

	fmt.Printf("%v", delReminder)

	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}
	if !AuthorityCheck(*delReminder.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de recordatorios de usuario de redis
	deleted, err2 := rdb.Del(ctx, "Reminders:"+*delReminder.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	deleted, err3 := rdb.Del(ctx, "Reminder&Tags:"+*delReminder.CodUsuario).Result()

	if err3 != nil {
		fmt.Printf("\nError de conexión: %v", err3)

	} else if deleted > 0 {
		fmt.Printf("\nRegsitro eliminado con éxito")
	} else {
		fmt.Printf("\nNo es encontró registro relacionado")
	}

	// Llamado al procedimiento
	result, err := db.Exec("CALL eliminar_recordatorios_multiple(?)", delReminder.N_idRecordatorios)

	rowsAffected, _ := result.RowsAffected()

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Log
	descripcion := fmt.Sprintf("Se eliminaron los recordatorios | IDs: %s | Usuario ID: %d",
		delReminder.N_idRecordatorios, delReminder.P_usuario)

	go func(uID int, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertarLog(uID, acc, desc)
	}(delReminder.P_usuario, "ELIMINAR_MULTIPLES_RECORDATORIOS", descripcion)

	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}
