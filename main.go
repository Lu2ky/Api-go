package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {
	//err := godotenv.Load("../../config/goapiconfig.env") //PARA LOCAL
	err := godotenv.Load() // Load enviorement variables
	if err != nil {
		log.Fatal(".env file (error corrupted/not found)")
	}
	cfg := mysql.NewConfig()          //Create the cfg for MySQL
	cfg.User = os.Getenv("DB_USER")   //User
	cfg.Passwd = os.Getenv("DB_PASS") //Pass
	cfg.Net = "tcp"
	cfg.Addr = os.Getenv("DB_ADDR") + ":" + os.Getenv("DB_ADDR_PORT")
	cfg.DBName = os.Getenv("DB_NAME")
	var err2 error
	db, err2 = sql.Open("mysql", cfg.FormatDSN())
	if err2 != nil {
		log.Fatal("Error connecting to database:", err2)
	}
	defer db.Close()
	router := gin.Default()
	router.Use(apiKeyAuth())
	/*
		Aqui están los métodos que provee la API, cuando se quiere obtener una consulta nueva de la BD, se tiene que
		especificar en esta sección. Todo debe tener los mismos nombres, en la URL y en el método de la consulta.
	*/
	//	Actividades oficiales
	router.GET("/GetOfficialScheduleByUserId/:id", getOfficialScheduleByUserId)
	router.POST("/GetActivityTimesData", getActivitiesTimesData)
	//	Comentarios de las actividades oficiales
	router.GET("/GetPersonalComments/:id", getPersonalCommentsByUserId)
	router.GET("/GetPersonalCourseComments/:id/:idCourse", getPersonalCommentsByUserIdAndCourseId)
	router.POST("/addPersonalComment", addPersonalComment)
	router.POST("/updatePersonalComment", updatePersonalComment)
	router.POST("/deletePersonalComment", deletePersonalComment)

	//	Actividades personales
	router.GET("/GetPersonalScheduleByUserId/:id", getPersonalScheduleByUserId)
	router.POST("/addPersonalActivity", addPersonalActivity)
	router.POST("/updatePersonalScheduleByIdCourse", updatePersonalScheduleByIdCourse)

	//	router.POST("/updateNameOfPersonalScheduleByIdCourse", updateNameOfPersonalScheduleByIdCourse)
	//	router.POST("/updateDescriptionOfPersonalScheduleByIdCourse", updateDescriptionOfPersonalScheduleByIdCourse)
	//	router.POST("/updateStartHourOfPersonalScheduleByIdCourse", updateStartHourOfPersonalScheduleByIdCourse)
	//	router.POST("/updateEndHourOfPersonalScheduleByIdCourse", updateEndHourOfPersonalScheduleByIdCourse)

	router.POST("/deleteOrRecoveryPersonalScheduleByIdCourse", deleteOrRecoveryPersonalScheduleByIdCourse)

	router.GET("/GetTiposCurso", GetTiposCurso)
	//	Etiquetas
	router.GET("/GetTagsByUserId/:id", GetTagsByUserId)
	router.GET("/GetTagsByUserIdAndReminderId/:id/:reminderId", GetTagsByUserIdAndReminderId)
	router.POST("/deleteTag", deleteTag)

	//	Recordatorios
	router.GET("/GetReminders/:id", GetRemindersByUserId)
	router.GET("/GetRemindersTags/:id", GetRemindersTagsByUserId)
	router.POST("/addReminder", addReminder)
	router.POST("/updateReminder", updateReminderById)
	router.POST("/deleteOrRecoverReminder", deleteOrRecoverReminder)

	//	Notificaciones y correos
	router.GET("/GetNotifications/:id", GetNotificaciones)
	router.POST("/addNotification", addNotificacion)
	router.POST("/muteNotification", muteNotification)
	router.POST("/addCorreo", addCorreo)

	// Importar horario
	router.POST("/importSchedule", importSchedule)

	//	Configuracion de usuario
	router.GET("/GetUserInfo/:id", GetUserInfo)
	//router.GET("/GetUserInfo/:id", GetUserInfo)

	//	LDAP
	router.POST("/auth", auth)
	router.POST("/addauthuser", createUser)
	router.POST("/addadmin", createAdmin)
	router.POST("/changepassword", changeusrpasswd)

	router.Run("0.0.0.0:8080") // The port number for expone the API
	//router.Run(":8080")

}
func method(c *gin.Context) {}

// c *gin.Context essential for method in GET/POST actions

//	--------------- Actividades oficiales ----------------------------------------

/*
	This function is a basic get for get the users from database


	Aquí está explicado un método el método GET para obtener las actividades oficiales.
*/

func getOfficialScheduleByUserId(c *gin.Context) {
	//	este ID sale de la URL | /GetOfficialScheduleByUserId/:id
	//	Param() se encarga de extraer los parámetros definidos en la ruta.
	id := c.Param("id")

	/*
		db.Query retorna rows y err
		rows = *sql.rows | Es un puntero que tiene información de la consulta.

		* Para iterar sobre los resultados se usa rows.Next()
		* Para leer los valores de cada fila se hace un rows.Scan()
		* Y para cerrar la consulta se hace un rows.Close(), lo cual es necesario para evitar fugas de recursos que causan
		errores como que ya no se pueden hacer conexiones.

		Cada vez que se hace el db.Query hay que hacer esos pasos para sacar la info de la consulta.

		El operador := lo que hace es definir una variable e inferir su tipo automáticamente.
	*/
	rows, err := db.Query(`SELECT ao.* FROM ActividadesOficiales ao JOIN Usuarios u ON ao.N_idUsuario = u.N_idUsuario WHERE u.T_codUsuario = ?`, id)

	//	si err != nil entonces significa que hay un error.
	//	nil es similar a null. Entonces si el error es nulo significa que no hay errores.

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	/*
		defer hace que cuando la función actual termine, entonces se ejecute rows.Close()
		es decir, después de hacer el return.

		Es buena práctica hacer el defer rows.Close inmediatamente después de abrir la consulta.
	*/
	defer rows.Close()

	//	Aquí se van a almacenar los resultados de la consulta.
	//	Se utiiza el OfficialSchedule para tener una estructura a la hora de guardar la información de la consulta.

	var ofcschedules []OfficialSchedule

	for rows.Next() {
		var ofcschedule OfficialSchedule
		err := rows.Scan(
			//	Lo que hace en cada parámetro aquí es asignarle a la dirección de memoria el resultado dado por la base de datos
			//	Es MUY importante que estén en el mismo orden que lo devuelve la consulta, porque sino puede haber errores
			//	Los nombres de cada atributo pueden ser diferentes, pero para no perderse, es mejor usar el mismo nombre.
			&ofcschedule.N_idHorario,
			&ofcschedule.N_iduser,
			&ofcschedule.N_idcourse,
			&ofcschedule.Nrc,
			&ofcschedule.Course,
			&ofcschedule.Tag,
			&ofcschedule.Teacher, //falta
			&ofcschedule.Day,
			&ofcschedule.StartHour,
			&ofcschedule.EndHour,
			&ofcschedule.Classroom,
			&ofcschedule.Credits,
			&ofcschedule.Standardofcalification,
			&ofcschedule.Campus,
			&ofcschedule.FechaInicio,
			&ofcschedule.FechaFinal,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}

		//	Y aquí se agrega el objeto ofcschedule al arreglo ofcschedules.
		ofcschedules = append(ofcschedules, ofcschedule)
	}

	//	Se verifica si hubo errores mientras se hizo la iteración usando rows.Err().
	//	Si Next() retorna False, entonces para revisar cuál fue el error se usa rows.Err()
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	//	Se retorna con código 200 (OK status) el arreglo formando anteriormente en formato JSON.
	c.JSON(200, ofcschedules)
}

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

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}

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

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Etiqueta alterada correctamente",
		"rowsAffected": rowsAffected,
	})
}

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

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{
		"message":      "Comentario alterado correctamente",
		"rowsAffected": rowsAffected,
	})
}

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

	c.JSON(200, gin.H{
		"message": "Notificacion creada correctamente",
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

	if rowsAffected == 0 {
		c.JSON(200, gin.H{"message": "No hubo cambios"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Notificacion MUTEADA correctamente",
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

	c.JSON(200, gin.H{
		"message": "Correo creado correctamente",
	})
}

// -------------------------- IMPORTAR HORARIO ----------------------------------

func importSchedule(c *gin.Context) {

	log.Println("Inicio importacion de horario")
	var newScheduleValue ImportSchedule

	err := c.BindJSON(&newScheduleValue)
	if err != nil {
		log.Printf("Error: formato inválido de JSON: %v", err)
		c.JSON(400, gin.H{"Error": "Formato invalido de json"})
		return

	}

	/*

		type ImportSchedule struct {
			Nombre           string `json:"nombre"`
			Semestre         int    `json:"semestre"`
			Programa         string `json:"programa"`
			CodUSuario       string `json:"codUsuario"`
			Nrc              string `json:"nrc"`
			NombreCurso      string `json:"nombreCurso"`
			Docente          string `json:"docente"`
			Creditos         int    `json:"creditos"`
			ModoCalificar    string `json:"modoCalificar"`
			Campus           string `json:"campus"`
			TipoCurso        string `json:"tipoCurso"`
			Dia              int    `json:"dia"`
			HoraInicio       string `json:"horaInicio"`
			HoraFin          string `json:"horaFin"`
			Salon            string `json:"salon"`
			PeriodoAcademico string `json:"periodoAcademico"`
		}

	*/

	
	log.Printf("Usuario: %s", newScheduleValue.CodUSuario)
	log.Printf("Curso: %s | NRC: %s", newScheduleValue.NombreCurso, newScheduleValue.Nrc)
	log.Printf("Programa: %s | Semestre: %d", newScheduleValue.Programa, newScheduleValue.Semestre)

	result, err := db.Exec("CALL importarHorario(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
		newScheduleValue.Nombre,
		newScheduleValue.Semestre,
		newScheduleValue.Programa,
		newScheduleValue.CodUSuario,
		newScheduleValue.Nrc,
		newScheduleValue.NombreCurso,
		newScheduleValue.Docente,
		newScheduleValue.Creditos,
		newScheduleValue.ModoCalificar,
		newScheduleValue.Campus,
		newScheduleValue.TipoCurso,
		newScheduleValue.Dia,
		newScheduleValue.HoraInicio,
		newScheduleValue.HoraFin,
		newScheduleValue.Salon,
		newScheduleValue.PeriodoAcademico,
	)

	if err != nil {
		log.Printf("Database error al ejecutar importarHorario: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return

	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Registros afectados: %d", rowsAffected)

	if rowsAffected == 0 {
		log.Println("No se encontró información para importar")
		c.JSON(404, gin.H{"error": "No se encuentra el archivo a importar"})
		return

	}

	log.Println("Importación completada correctamente")
	log.Println("Fin importacion de horario")

	c.JSON(200, gin.H{
		"message": "Horario importado correctamente",
	})

}

//	------------------------ FUNCIONALIDADES DEL USUARIO ------------------------ //

func GetUserInfo(c *gin.Context) {

	id_user := c.Param("id")

	//	Consulta
	rows, err := db.Query(
		`
		SELECT u.N_idUsuario, u.T_nombre, u.T_correo, u.N_semestreActual, u.T_programa, u.TM_antelacionNotis
		FROM Usuarios u
		WHERE u.T_codUsuario = ?
		`,
		id_user,
	)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var userDataArray []UserData

	//	Escanear y guardar la información de la consulta
	for rows.Next() {
		var userData UserData
		err := rows.Scan(
			&userData.N_idUsuario,
			&userData.T_nombre,
			&userData.T_correo,
			&userData.N_semestreActual,
			&userData.T_programa,
			&userData.TM_antelacionNotis,
		)

		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}
		userDataArray = append(userDataArray, userData)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	c.JSON(200, userDataArray)

}

//	------------------------ FUNCIONALIDADES DEL LDAP ------------------------ //

func auth(c *gin.Context) {
	var User UserAuth
	err := c.BindJSON(&User)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}
	token, userU, err := ConnectLDAP(User.User, User.Pass, JWTManager{
		Secret: []byte(os.Getenv("JWT_SECRET")),
		TTL:    24 * time.Hour,
		Issuer: "horario_estudiantes",
	})
	if err != nil {
		log.Printf("ldap error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, gin.H{
		"Token":    token,
		"UserAuth": userU,
	})
}

func (j JWTManager) Generate(u *User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: u.Username,
		Roles:  u.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-clockSkewTolerance)),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
			Subject:   u.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j JWTManager) Validate(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("sin token")
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("erro en el metodo de inicio")
		}
		return j.Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("token invalido")
	}

	if j.Issuer != "" && claims.Issuer != j.Issuer {
		return nil, errors.New("que?")
	}

	return claims, nil
}

func ConnectLDAP(user string, pass string, j JWTManager) (string, *User, error) {
	l, err := ldap.DialURL("ldap://" + os.Getenv("LDAP_ADDR") + ":" + os.Getenv("LDAP_PORT"))
	if err != nil {
		return "", nil, err
	}
	defer l.Close()

	l.SetTimeout(5 * time.Second)

	err = l.StartTLS(&tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return "", nil, err
	}

	err = l.Bind(user+"@adhe.local", pass)
	if err != nil {
		return "", nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		"DC=adhe,DC=local",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", user),
		[]string{"memberOf", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return "", nil, err
	}

	if len(sr.Entries) == 0 {
		return "", nil, errors.New("usuario no encontrado en LDAP")
	}

	entry := sr.Entries[0]

	var roles []string
	for _, groupDN := range entry.GetAttributeValues("memberOf") {
		dn, err := ldap.ParseDN(groupDN)
		if err == nil && len(dn.RDNs) > 0 {
			cn := dn.RDNs[0].Attributes[0].Value
			roles = append(roles, cn)
		}
	}

	u := &User{
		Username: user,
		Roles:    roles,
	}

	token, err := j.Generate(u)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}
func createUser(c *gin.Context) {
	var req UserAuth

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}

	err := CreateLDAPUser(
		os.Getenv("ADMIN_LDAP_ADMIN"),
		os.Getenv("ADMIN_LDAP_PASS"),
		req.User,
		req.Pass,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Usuario creado correctamente"})
}
func CreateLDAPUser(adminUser, adminPass, username, password string) error {
	l, err := ldap.DialURL("ldap://" + os.Getenv("LDAP_ADDR") + ":" + os.Getenv("LDAP_PORT"))
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}

	err = l.Bind(adminUser+"@adhe.local", adminPass)
	if err != nil {
		return err
	}

	userDN := fmt.Sprintf("CN=%s,CN=Users,DC=adhe,DC=local", username)

	addReq := ldap.NewAddRequest(userDN, nil)

	addReq.Attribute("objectClass", []string{
		"top",
		"person",
		"organizationalPerson",
		"user",
	})

	addReq.Attribute("cn", []string{username})
	addReq.Attribute("sAMAccountName", []string{username})
	addReq.Attribute("userPrincipalName", []string{username + "@adhe.local"})
	addReq.Attribute("displayName", []string{username})
	addReq.Attribute("userAccountControl", []string{"544"})

	err = l.Add(addReq)
	if err != nil {
		return err
	}

	quotedPwd := fmt.Sprintf("\"%s\"", password)
	utf16Pwd := utf16.Encode([]rune(quotedPwd))

	pwdBytes := make([]byte, len(utf16Pwd)*2)
	for i, v := range utf16Pwd {
		binary.LittleEndian.PutUint16(pwdBytes[i*2:], v)
	}

	modPwd := ldap.NewModifyRequest(userDN, nil)
	modPwd.Replace("unicodePwd", []string{string(pwdBytes)})

	err = l.Modify(modPwd)
	if err != nil {
		return fmt.Errorf("error seteando password: %v", err)
	}
	modEnable := ldap.NewModifyRequest(userDN, nil)
	modEnable.Replace("userAccountControl", []string{"512"})

	err = l.Modify(modEnable)
	if err != nil {
		return fmt.Errorf("error habilitando usuario: %v", err)
	}

	groupDN := "CN=Usuario,CN=Users,DC=adhe,DC=local"

	modGroup := ldap.NewModifyRequest(groupDN, nil)
	modGroup.Add("member", []string{userDN})

	err = l.Modify(modGroup)
	if err != nil {
		return fmt.Errorf("error agregando al grupo Usuario: %v", err)
	}

	return nil
}

// Last test for today :P -Luky (CI/CD test)
