package main

import (
	//"encoding/json"
	//"net/http"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

//Pruebita
/* Saving the session of MySQL, this is global for the access in all methods */

type User struct {
	Id       int    `json:"T_idUsuario"`
	Name     string `json:"T_nombre"`
	Programa string `json:"T_programa"`
}
type OfficialSchedule struct {
	N_iduser               int             `json:"N_iduser"`
	N_idcourse             int             `json:"N_idcourse"`
	Nrc                    string          `json:"Nrc"`
	Course                 string          `json:"Course"`
	Tag                    string          `json:"Tag"`
	Teacher                string          `json:"Teacher"`
	Day                    int             `json:"Day"`
	StartHour              string          `json:"StartHour"`
	EndHour                string          `json:"EndHour"`
	Classroom              string          `json:"Classroom"`
	Credits                sql.NullFloat64 `json:"Credits"`
	Standardofcalification string          `json:"Standardofcalification"`
	Campus                 string          `json:"Campus"`
	AcademicPeriod         string          `json:"AcademicPeriod"`
}
type PersonalSchedule struct {
	N_iduser    int            `json:"N_iduser"`
	N_idcourse  int            `json:"N_idcourse"`
	Activity    string         `json:"Activity"`
	Tag         string         `json:"Tag"`
	Description sql.NullString `json:"Description"`
	Dt_Start    sql.NullString `json:"Dt_Start"`
	Dt_End      sql.NullString `json:"Dt_End"`
	Day         int            `json:"Day"`
	StartHour   string         `json:"StartHour"`
	EndHour     string         `json:"EndHour"`
	IsDeleted   *sql.NullBool  `json:"IsDeleted"`
}
type Tags struct {
	T_name string `json:"T_name"`
}
type PersonalScheduleNewValue struct {
	NewActivityValue   string `json:"NewActivityValue" binding:"required"`
	IdPersonalSchedule int    `json:"IdPersonalSchedule" binding:"required"`
}
type forDeleteOrRecoveryPersonalSchedule struct {
	IsDeleted          *bool `json:"IsDeleted" binding:"required"`
	IdPersonalSchedule int   `json:"IdPersonalSchedule" binding:"required"`
}
type NewPersonalActivity struct {
	/*Activity          string `json:"Activity"`
	Description       string `json:"Description"`
	IdTag             int    `json:"IdTag"`
	Day               int    `json:"Day"`
	StartHour         string `json:"StartHour"`
	EndHour           string `json:"EndHour"`
	N_iduser          int    `json:"N_iduser"`
	Id_AcademicPeriod int    `json:"Id_AcademicPeriod"`*/
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
type ofcComments struct {
	N_idHorario  int           `json:"N_idHorario"`
	N_idUsuario  int           `json:"N_idUsuario"`
	N_idCurso    int           `json:"N_idCurso"`
	Curso        string        `json:"Curso"`
	T_comentario string        `json:"T_comentario"`
	B_isDeleted  *sql.NullBool `json:"B_isDeleted"`
}
type new_ofcComments struct {
	N_idHorario  int    `json:"N_idHorario"`
	N_idUsuario  int    `json:"N_idUsuario"`
	N_idCurso    int    `json:"N_idCurso"`
	Curso        string `json:"Curso"`
	T_comentario string `json:"T_comentario"`
}
type Reminders struct{
	N_idUsuario				int			`json:"N_idUsuario"`
	N_idRecordatorio		int			`json:"N_idRecordatorio"`
	T_nombre				string		`json:"T_nombre"`
	T_descripción			string		`json:"T_descripción"`
	Dt_fechaVencimiento		string		`json:"Dt_fechaVencimiento"`
	B_isDeleted				*bool		`json:"B_isDeleted"`
	T_Prioridad				string		`json:"T_Prioridad"`
	Etiqueta				string		`json:"Etiqueta"`
}

func apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		validAPIKey := os.Getenv("API_KEY")
		if validAPIKey == "" {
			log.Fatal("API_KEY no configurada .env")
		}
		if apiKey == "" {
			c.JSON(401, gin.H{"error": "API Key necesaria para uso"})
			c.Abort()
			return
		}
		if apiKey != validAPIKey {
			c.JSON(403, gin.H{"error": "API Key invalida"})
			c.Abort()
			return
		}

		c.Next()
	}
}
func main() {
	err := godotenv.Load("../../config/goapiconfig.env") // Load enviorement variables
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
	//Métodos para obtener información
	router.GET("/GetOfficialScheduleByUserId/:id", getOfficialScheduleByUserId)
	router.GET("/GetPersonalScheduleByUserId/:id", getPersonalScheduleByUserId)
	router.GET("/GetPersonalComments/:id", getPersonalCommentsByUserIdAndCourseId)
	router.GET("/GetTags", getTags)
	router.GET("/GetReminders/:id", GetRemindersByUserId)
	//Métodos para hacer modificaciones a la BD
	router.POST("/updateNameOfPersonalScheduleByIdCourse", updateNameOfPersonalScheduleByIdCourse)
	router.POST("/updateDescriptionOfPersonalScheduleByIdCourse", updateDescriptionOfPersonalScheduleByIdCourse)
	router.POST("/updateStartHourOfPersonalScheduleByIdCourse", updateStartHourOfPersonalScheduleByIdCourse)
	router.POST("/updateEndHourOfPersonalScheduleByIdCourse", updateEndHourOfPersonalScheduleByIdCourse)
	router.POST("/deleteOrRecoveryPersonalScheduleByIdCourse", deleteOrRecoveryPersonalScheduleByIdCourse)
	router.POST("/addPersonalActivity", addPersonalActivity)
	router.POST("/addPersonalComment", addPersonalComment)

	router.Run("0.0.0.0:3913") // The port number for expone the API
}
func method(c *gin.Context) {}

// c *gin.Context essential for method in GET/POST actions

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
			&ofcschedule.AcademicPeriod, //falta
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

func getPersonalCommentsByUserIdAndCourseId(c *gin.Context) {
	id_User := c.Param("id")
	rows, err := db.Query(`SELECT * FROM ComentariosOficiales WHERE N_idUsuario=(SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario=?);`, id_User)
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


//	Aquí está explicado un método POST, en este caso, Actualizar el nombre de una actividad personal.

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
	result, err := db.Exec("UPDATE ActividadesPersonales SET B_isDeleted = ? WHERE N_idCurso=?", deleteValue.IsDeleted, deleteValue.IdPersonalSchedule)
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
func getTags(c *gin.Context) {
	rows, err := db.Query(`SELECT T_nombre FROM Etiquetas;`)
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
			&Tags.T_name,
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
func addPersonalActivity(c *gin.Context) {
	var newPerActivity NewPersonalActivity

	//	Se asignan los valores el JSON a la estructura newPerActivity
	err := c.BindJSON(&newPerActivity)
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
		newPerActivity.p_usuario,
		newPerActivity.p_nombreCurso,
		newPerActivity.p_descripcion,
		newPerActivity.p_fechaInicio,
		newPerActivity.p_fechaFin,
		newPerActivity.p_dia,
		newPerActivity.p_horaInicio,
		newPerActivity.p_horaFin,
		newPerActivity.p_hp_periodooraFin
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
		newPerActivity.Activity,
		newPerActivity.IdTag,
		newPerActivity.Description,
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
		newPerActivity.Day,
		newPerActivity.StartHour,
		newPerActivity.EndHour,
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
		newPerActivity.N_iduser,
		idCurso,
		newPerActivity.Id_AcademicPeriod)
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
func addPersonalComment(c *gin.Context) {
	var newComment new_ofcComments
	err := c.BindJSON(&newComment)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	result, err := db.Exec(
		"INSERT INTO Comentarios (N_idHorario, N_idUsuario, N_idCurso, T_comentario) VALUES (?, ?, ?, ?)",
		newComment.N_idHorario,
		newComment.N_idUsuario,
		newComment.N_idCurso,
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

	//	--------------- Recordatorios ----------------------------------------

	//	Obtener la lista de los recordatorios

	func GetRemindersByUserId(c *gin.Context) {

		/*
			type Reminders struct{
				N_idUsuario			int			`json:"N_idUsuario"`
				N_idRecordatorio	int			`json:"N_idRecordatorio"`
				T_nombre			string		`json:"T_nombre"`
				T_descripción		string		`json:"T_descripción"`
				Dt_fechaVencimiento	string		`json:"Dt_fechaVencimiento"`
				B_isDeleted			*bool		`json:"B_isDeleted"`
				T_Prioridad			string		`json:"T_Prioridad"`
				Etiqueta			string		`json:"Etiqueta"`
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
			id_User
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
				&reminder.N_idUsuario,
				&reminder.N_idRecordatorio,
				&reminder.T_nombre,
				&reminder.T_descripción,
				&reminder.Dt_fechaVencimiento,
				&reminder.B_isDeleted,
				&reminder.T_Prioridad,
				&reminder.Etiqueta
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

		c.JSON(200, reminder)
	}

}

// Last test for today :P -Luky (CI/CD test)
