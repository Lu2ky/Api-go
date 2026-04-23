package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//	--------------- Actividades personales ----------------------------------------

func getPersonalScheduleByUserId(c *gin.Context) {
	id := c.Param("id")
	var rows *sql.Rows

	//	Consulta a redis
	val, err2 := rdb.Get(c.Request.Context(), "PersonalSchedule:"+id).Result()

	if err2 == nil {
		fmt.Printf("\n Si existe registro")
		var perschedules []PersonalSchedule

		err := json.Unmarshal([]byte(val), &perschedules)

		if err == nil {
			c.JSON(200, perschedules)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

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
		err := rows.Scan(
			&perschedule.N_iduser,
			&perschedule.N_idcourse,
			&perschedule.Activity,
			//&perschedule.Tag,
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

	// Convertir a formato apto para redis
	data, err := json.Marshal(perschedules)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err3 := rdb.Set(ctx, "PersonalSchedule:"+id, data, 48*time.Hour).Err()

	if err3 != nil {
		log.Printf("Error al guardar en Redis: %v", err3)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
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
	if !AuthorityCheck(*personalNewValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
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

	// Borrar registro de datos de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalSchedule:"+*personalNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

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

	// Log
	var userId int
	err4 := db.QueryRow("CALL get_id_tabla(?)", *personalNewValue.CodUsuario).Scan(&userId)
	if err4 != nil {
		log.Printf("Error obteniendo ID: %v", err)
	}

	descripcion := fmt.Sprintf("Se actualizó actividad personal | ID: %d | Usuario ID: %d",
		personalNewValue.P_idCurso, userId)

	go func(uID string, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertLogCod(uID, acc, desc)
	}(*personalNewValue.CodUsuario, "ACTUALIZAR_ACTIVIDAD_PERSONAL", descripcion)

	c.JSON(200, gin.H{
		"message": "Actividad actualizada correctamente",
	})
}

func deleteOrRecoveryPersonalScheduleByIdCourse(c *gin.Context) {
	var deleteValue forDeleteOrRecoveryPersonalSchedule

	err := c.BindJSON(&deleteValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	if !AuthorityCheck(*deleteValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
		return
	}

	// Borrar registro de datos de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalSchedule:"+*deleteValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	// Aquí se hace la acutalización
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

	// Log
	descripcion := fmt.Sprintf("Se eliminó actividad personal | ID: %d | Usuario ID: %d",
		deleteValue.IdPersonalSchedule, deleteValue.N_idUsuario)

	go func(uID string, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertLogCod(uID, acc, desc)
	}(*deleteValue.CodUsuario, "ELIMINAR_ACTIVIDAD_PERSONAL", descripcion)

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
	if !AuthorityCheck(*personalNewValue.CodUsuario, c) {
		c.AbortWithStatusJSON(401, gin.H{"error": "Autorización requerida"})
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

	// Borrar registro de datos de usuario de redis
	deleted, err2 := rdb.Del(ctx, "PersonalSchedule:"+*personalNewValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	var newActId int

	//	Aquí se hace el llamado al Procedimiento
	err = db.QueryRow("SELECT crear_actividad_personal(?, ?, ?, ?, ?, ?, ?, ?)",
		personalNewValue.P_usuario,
		personalNewValue.P_nombreCurso,
		personalNewValue.P_descripcion,
		personalNewValue.P_fechaInicio,
		personalNewValue.P_fechaFin,
		personalNewValue.P_dia,
		personalNewValue.P_horaInicio,
		personalNewValue.P_horaFin,
		//personalNewValue.P_periodo,
	).Scan(&newActId)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	/*
		rowsAffected, _ := result.RowsAffected()

		if rowsAffected == 0 {
			c.JSON(404, gin.H{"error": "Personal schedule not found"})
			return
		}
	*/
	descripcion := "Se creó actividad personal: " + personalNewValue.P_nombreCurso

	insertarLog(
		personalNewValue.P_usuario,
		"CREAR_ACTIVIDAD_PERSONAL",
		descripcion,
	)
	c.JSON(200, gin.H{
		"message":      "Actividad creada correctamente",
		"new_activity": newActId,
	})
}

// Get tipo cursos QUERDE AQUIIIIIIIIIIIIIIIII ES DIFERENTE ES OTRO GET
func GetTiposCurso(c *gin.Context) {

	/*
		type TipoCurso struct {
			N_idTipoCurso int    `json:"N_idTipoCurso"`
			T_nombre      string `json:"T_nombre"`
			B_isDeleted   int  `json:"B_isDeleted"`
		}
	*/
	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "CourseType").Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var tiposCursoArray []TipoCurso

		err := json.Unmarshal([]byte(val), &tiposCursoArray)

		if err == nil {
			c.JSON(200, tiposCursoArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
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

	// Convertir a formato apto para redis
	data, err := json.Marshal(tiposCursoArray)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err2 := rdb.Set(ctx, "CourseType", data, 48*time.Hour).Err()

	if err2 != nil {
		log.Printf("Error al guardar en Redis: %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, tiposCursoArray)
}
