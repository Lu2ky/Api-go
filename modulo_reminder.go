package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

//	--------------- Recordatorios ----------------------------------------
//
// Obtener la lista de los recordatorios
func GetRemindersTagsByUserId(c *gin.Context) {

	//	Id del usuario
	id_User := c.Param("id")

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

	/*
		type ReminderNewValue struct {
			P_usuario     int            `json:"P_usuario"`
			P_nombre      string         `json:"P_nombre"`
			P_descripcion string         `json:"P_descripcion"`
			P_fecha       string         `json:"P_fecha"`
			P_prioridad   int            `json:"P_prioridad"`
			P_tag1        sql.NullString `json:"P_tag1"`
			P_tag2        sql.NullString `json:"P_tag2"`
			P_tag3        sql.NullString `json:"P_tag3"`
			P_tag4        sql.NullString `json:"P_tag4"`
			P_tag5        sql.NullString `json:"P_tag5"`
		}
	*/

	// Iniciar transacción para garantizar la misma conexión
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error al iniciar transacción: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer tx.Rollback()

	// Aquí se hace el llamado al Procedimiento
	rows, err := tx.Query("SELECT crear_recordatorio_5tags(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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
	)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var newID int64

	// Navegar por todos los result sets hasta encontrar el que tiene el ID
	for {
		if rows.Next() {
			err = rows.Scan(&newID)
			if err != nil {
				log.Printf("Error al leer resultado: %v", err)
			}
		}
		if !rows.NextResultSet() {
			break
		}
	} // <-- el for cierra aquí

	if err = rows.Err(); err != nil {
		log.Printf("Error en rows: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Confirmar la transacción
	if err = tx.Commit(); err != nil {
		log.Printf("Error al confirmar transacción: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("ID del ToDo creado: %d", newID)
	descripcion := "Se creó recordatorio ID: " + strconv.FormatInt(newID, 10) +
		" | Usuario: " + strconv.Itoa(reminderNewValue.P_usuario) +
		" | Nombre: " + reminderNewValue.P_nombre

	insertarLog(reminderNewValue.P_usuario, "CREAR_RECORDATORIO", descripcion)
	c.JSON(200, gin.H{
		"message":    "Recordatorio creado correctamente",
		"InsertedId": newID,
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

	/*
		type EditReminder struct {
			P_idToDo		int				`json:"P_idToDo"`
			P_nombre		sql.NullString			`json:"P_nombre"`
			P_descripcion	sql.NullString			`json:"P_descripcion"`
			P_fecha			sql.NullString			`json:"P_fecha"`
			P_prioridad		sql.NullInt64 	`json:"P_prioridad"`
			P_tag1			string	`json:"P_tag1"`
			P_tag2			string	`json:"P_tag2"`
			P_tag3			string	`json:"P_tag3"`
			P_tag4			string	`json:"P_tag4"`
			P_tag5			string	`json:"P_tag5"`
		}
	*/

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
		c.JSON(404, gin.H{"error": "Personal schedule not found"})
		return
	}
descripcion := "Se actualizó recordatorio ID: " +
	strconv.Itoa(reminderNewValue.P_idToDo) +
	" | Usuario: " + strconv.Itoa(reminderNewValue.P_usuario)

	insertarLog(reminderNewValue.P_usuario, "UPDATE_RECORDATORIO", descripcion)

	c.JSON(200, gin.H{
		"message": "Recordatorio creado correctamente",
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

	result, err := db.Exec("CALL eliminar_recordatorio(?)", delReminder.N_idRecordatorio)
	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	userID := delReminder.P_usuario

	descripcion := "Se eliminó/recuperó recordatorio ID: " +
		strconv.Itoa(delReminder.N_idRecordatorio) +
		" | Usuario: " + strconv.Itoa(userID)

	insertarLog(userID, "DELETE_RECORDATORIO", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
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

	result, err := db.Exec("CALL eliminar_recordatorios_multiple(?)", delReminder.N_idRecordatorios)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	userID := delReminder.P_usuario

	descripcion := "Se eliminaron los recordatorios de ID: " +
		delReminder.N_idRecordatorios +
		" | Usuario: " + strconv.Itoa(userID)

	insertarLog(userID, "DELETE_RECORDATORIO", descripcion)

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}
