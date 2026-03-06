package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

//	--------------- Actividades personales ----------------------------------------

func getPersonalScheduleByUserId(c *gin.Context) {
	id := c.Param("id")
	var rows *sql.Rows
	rows, err := db.Query(`
		SELECT ao.*
		FROM ActividadesPersonales ao
		JOIN Usuarios u ON ao.N_idUsuario = u.N_idUsuario
		WHERE u.T_codUsuario = ?
	`, id)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var perschedules []PersonalSchedule
	for rows.Next() {
		var perschedule PersonalSchedule
		err := rows.Scan(&perschedule.N_iduser,
			&perschedule.N_idcourse,
			&perschedule.Activity, &perschedule.Tag,
			&perschedule.Description,
			&perschedule.Dt_Start,
			&perschedule.Dt_End,
			&perschedule.Day,
			&perschedule.StartHour,
			&perschedule.EndHour,
			&perschedule.IsDeleted)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"Error": "Error en procesamiento de datos"})
			return
		}
		perschedules = append(perschedules, perschedule)

	}
	c.JSON(200, perschedules)
}

//	Aquí está explicado un método POST, en este caso, Actualizar el nombre de una actividad personal.

// Procedimiento: Actualizar actividad personal // TODO
func updatePersonalScheduleByIdCourse(c *gin.Context) {
	//	Aquí se instancia la estructura definida en la parte superior.
	var personalNewValue EditPersonalActivity

	/*
		BindJSON() se encarga de tomar el body request de la petición y lo convierte en una estructura de GO
		Aquí es importante que el JSON del body tenga los mismos campos ya definidos, en este caso, en PersonalScheduleNewValue
		También retorna un error en caso de haber uno.

		Se usa como argumento &newValue para darle la dirección de memoria de la estructura GO y así almacenar la info.
	*/

	//	Se asignan los valores el JSON a la estructura reminderNewValue
	err := c.BindJSON(&personalNewValue)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	/*
		type EditPersonalActivity struct {
			P_idCurso     int    `json:"P_idCurso"`
			P_nombreCurso string `json:"P_nombreCurso"`
			P_descripcion string `json:"P_descripcion"`
			P_fechaInicio string `json:"P_fechaInicio"`
			P_fechaFin    string `json:"P_fechaFin"`
			P_dia         int    `json:"P_dia"`
			P_horaInicio  string `json:"P_horaInicio"`
			P_horaFin     string `json:"P_horaFin"`
		}
	*/

	/*
		El método Query() se utilizaba cuando la consulta era un SELECT.
		En este caso, un UPDATE, se utiliza Exec(), y retorna:
			sql.Result, error

		Los signos de pregunta (?) indican los parámetros que se envían a la consulta.
		en el segundo argumento, los parámetros deben estar en el mismo orden que son solicitados en la consulta.
	*/

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("CALL editar_actividad_personal(?, ?, ?, ?, ?, ?, ?, ?)",
		personalNewValue.P_idCurso,
		personalNewValue.P_nombreCurso,
		personalNewValue.P_descripcion,
		personalNewValue.P_fechaInicio,
		personalNewValue.P_fechaFin,
		personalNewValue.P_dia,
		personalNewValue.P_horaInicio,
		personalNewValue.P_horaFin,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	//	rowsAffected contiene la cantidad de filas que fueron modificadas
	//	Se utiliza un guión al piso (_) para ignorar el error, porque result.RowsAffected retorna int64, error

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Personal schedule not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Actividad actualizada correctamente",
	})
}

func updateNameOfPersonalScheduleByIdCourse(c *gin.Context) {
	//	Aquí se instancia la estructura definida en la parte superior.
	var newValue PersonalScheduleNewValue

	/*
		BindJSON() se encarga de tomar el body request de la petición y lo convierte en una estructura de GO
		Aquí es importante que el JSON del body tenga los mismos campos ya definidos, en este caso, en PersonalScheduleNewValue
		También retorna un error en caso de haber uno.

		Se usa como argumento &newValue para darle la dirección de memoria de la estructura GO y así almacenar la info.
	*/
	err := c.BindJSON(&newValue)

	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	/*
		El método Query() se utilizaba cuando la consulta era un SELECT.
		En este caso, un UPDATE, se utiliza Exec, y retorna:
			sql.Result, error

		Los signos de pregunta (?) indican los parámetros que se envían a la consulta.
		en el segundo argumento, los parámetros deben estar en el mismo orden que son solicitados en la consulta.
	*/
	result, err := db.Exec("UPDATE ActividadesPersonales SET Actividad = ? WHERE N_idCurso= ? ", newValue.NewActivityValue, newValue.IdPersonalSchedule)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	//	rowsAffected contiene la cantidad de filas que fueron modificadas
	//	Se utiliza un guión al piso (_) para ignorar el error, porque result.RowsAffected retorna int64, error

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Personal schedule not found"})
		return
	}

	c.JSON(200, gin.H{
		"message":      "Personal schedule updated successfully",
		"rowsAffected": rowsAffected,
	})
}
func updateDescriptionOfPersonalScheduleByIdCourse(c *gin.Context) {
	var newValue PersonalScheduleNewValue
	err := c.BindJSON(&newValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	result, err := db.Exec("UPDATE ActividadesPersonales SET Descripcion = ? WHERE N_idCurso= ? ", newValue.NewActivityValue, newValue.IdPersonalSchedule)
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

	c.JSON(200, gin.H{
		"message":      "Personal schedule updated successfully",
		"rowsAffected": rowsAffected,
	})
}
func updateStartHourOfPersonalScheduleByIdCourse(c *gin.Context) {
	var newValue PersonalScheduleNewValue
	err := c.BindJSON(&newValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	result, err := db.Exec("UPDATE ActividadesPersonales SET Hora_Inicio = ? WHERE N_idCurso= ? ", newValue.NewActivityValue, newValue.IdPersonalSchedule)
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

	c.JSON(200, gin.H{
		"message":      "Personal schedule updated successfully",
		"rowsAffected": rowsAffected,
	})
}
func updateEndHourOfPersonalScheduleByIdCourse(c *gin.Context) {
	var newValue PersonalScheduleNewValue
	err := c.BindJSON(&newValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	result, err := db.Exec("UPDATE ActividadesPersonales SET Hora_Fin = ? WHERE N_idCurso= ? ", newValue.NewActivityValue, newValue.IdPersonalSchedule)
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

	c.JSON(200, gin.H{
		"message":      "Personal schedule updated successfully",
		"rowsAffected": rowsAffected,
	})
}
func deleteOrRecoveryPersonalScheduleByIdCourse(c *gin.Context) {
	var deleteValue forDeleteOrRecoveryPersonalSchedule

	err := c.BindJSON(&deleteValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}

	result, err := db.Exec("CALL eliminar_actividad_personal (?);", deleteValue.IdPersonalSchedule)

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

	c.JSON(200, gin.H{
		"message":      "Personal schedule updated successfully",
		"rowsAffected": rowsAffected,
	})
}

// Procedimiento: Agregar actividad personal
func addPersonalActivity(c *gin.Context) {
	var personalNewValue NewPersonalActivity

	//	Se asignan los valores el JSON a la estructura personalNewValue
	err := c.BindJSON(&personalNewValue)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	/*
		type NewPersonalActivity struct {
			P_usuario		int			`json:"P_usuario"`
			P_nombreCurso	string		`json:"P_nombreCurso"`
			P_descripcion	string		`json:"P_descripcion"`
			P_fechaInicio	string		`json:"P_fechaInicio"`
			P_fechaFin		string		`json:"P_fechaFin"`
			P_dia			int			`json:"P_dia"`
			P_horaInicio	string		`json:"P_horaInicio"`
			P_horaFin		string		`json:"P_horaFin"`
			P_periodo		int			`json:"P_periodo"`
		}
	*/

	//	Aquí se hace el llamado al Procedimiento
	result, err := db.Exec("CALL crear_actividad_personal(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		personalNewValue.P_usuario,
		personalNewValue.P_nombreCurso,
		personalNewValue.P_descripcion,
		personalNewValue.P_fechaInicio,
		personalNewValue.P_fechaFin,
		personalNewValue.P_dia,
		personalNewValue.P_horaInicio,
		personalNewValue.P_horaFin,
		personalNewValue.P_periodo,
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

	/*
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Transaction error: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		result0, err0 := tx.Exec(
			"INSERT INTO Cursos (T_nombre, N_idEtiqueta, T_descripcion) VALUES (?, ?, ?);",
			reminderNewValue.Activity,
			reminderNewValue.IdTag,
			reminderNewValue.Description,
		)
		idCurso, _ := result0.LastInsertId()

		if err0 != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err0)
			c.JSON(500, gin.H{"error": "Error en primer query"})
			return
		}
		result1, err1 := tx.Exec(
			"INSERT INTO dias_clase(N_dia, TM_horaInicio, TM_horaFin) VALUES (?, ?, ?)",
			reminderNewValue.Day,
			reminderNewValue.StartHour,
			reminderNewValue.EndHour,
		)
		nIdDias, _ := result1.LastInsertId()
		if err1 != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err1)
			c.JSON(500, gin.H{"error": "Error en segunda query"})
			return
		}

		_, err = tx.Exec(
			"INSERT INTO Materia_has_dias_clase(N_idCurso, N_idDiasClase) VALUES (?, ?);",
			idCurso,
			nIdDias,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Error en tercer query"})
			return
		}
		_, err = tx.Exec(
			"INSERT INTO horario (N_idUsuario, N_idCurso, N_idPeriodoAcademico) VALUES (?, ?,?);",
			reminderNewValue.N_iduser,
			idCurso,
			reminderNewValue.Id_AcademicPeriod)
		if err != nil {
			tx.Rollback()
			log.Printf("Database error: %v", err)
			c.JSON(500, gin.H{"error": "Error en cuarto query"})
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("Commit error: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
	*/
	c.JSON(200, gin.H{
		"message": "Actividad creada correctamente",
	})
}

// Get tipo cursos
func GetTiposCurso(c *gin.Context) {

	/*
		type TipoCurso struct {
			N_idTipoCurso int    `json:"N_idTipoCurso"`
			T_nombre      string `json:"T_nombre"`
			B_isDeleted   int  `json:"B_isDeleted"`
		}
	*/

	rows, err := db.Query("SELECT * FROM TipoCurso")

	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	defer rows.Close()

	var tiposCursoArray []TipoCurso

	for rows.Next() {

		var tipoCurso TipoCurso

		err := rows.Scan(
			&tipoCurso.N_idTipoCurso,
			&tipoCurso.T_nombre,
			&tipoCurso.B_isDeleted,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en el procesamiento de datos."})
			return
		}

		tiposCursoArray = append(tiposCursoArray, tipoCurso)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	c.JSON(200, tiposCursoArray)
}
