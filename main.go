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
	err := godotenv.Load("../../config/goapiconfig.env") //PARA LOCAL
	//err := godotenv.Load() // Load enviorement variables
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
	router.GET("/GetAcademicPeriods", getAcademicPeriods)
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

	// Token
	router.POST("/generateToken", generateToken)

	router.Run("0.0.0.0:8080") // The port number for expone the API
	//router.Run(":8080")

}
func method(c *gin.Context) {}

// c *gin.Context essential for method in GET/POST actions
