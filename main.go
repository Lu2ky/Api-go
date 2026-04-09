package main

import (
	"database/sql"
	"log"
	"os"

	"context"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB
var ctx = context.Background()
var rdb *redis.Client

func init() {
	err := godotenv.Load("../../config/goapiconfig.env") //PARA LOCAL
	//err := godotenv.Load() // Load enviorement variables

	if err != nil {
		log.Println("No se pudo cargar el archivo .env, usando variables de sistema")
	}

	// Leer las variables
	addr := os.Getenv("DB_ADDR_REDIS") + ":" + os.Getenv("DB_ADDR_PORT_REDIS")
	pass := os.Getenv("DB_PASS_REDIS")

	// Inicializar el cliente
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})
}

func main() {
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

	registerLegacyRoutes(router)
	v1 := router.Group("/api/v1")
	registerV1Routes(v1)

	router.Run("0.0.0.0:8080") // The port number for expone the API

	//router.Run(":8080")
}

func registerLegacyRoutes(router gin.IRoutes) {
	// Actividades oficiales
	router.GET("/GetOfficialScheduleByUserId/:id", getOfficialScheduleByUserId)
	router.POST("/GetActivityTimesData", getActivitiesTimesData)
	router.GET("/GetAcademicPeriods", getAcademicPeriods)

	// Comentarios
	router.GET("/GetPersonalComments/:id", getPersonalCommentsByUserId)
	router.GET("/GetPersonalCourseComments/:id/:idCourse", getPersonalCommentsByUserIdAndCourseId)
	router.POST("/addPersonalComment", addPersonalComment)
	router.POST("/updatePersonalComment", updatePersonalComment)
	router.POST("/deletePersonalComment", deletePersonalComment)

	// Actividades personales
	router.GET("/GetPersonalScheduleByUserId/:id", getPersonalScheduleByUserId)
	router.POST("/addPersonalActivity", addPersonalActivity)
	router.POST("/updatePersonalScheduleByIdCourse", updatePersonalScheduleByIdCourse)
	router.POST("/deleteOrRecoveryPersonalScheduleByIdCourse", deleteOrRecoveryPersonalScheduleByIdCourse)
	router.GET("/GetTiposCurso", GetTiposCurso)

	// Etiquetas
	router.GET("/GetTagsByUserId/:id", GetTagsByUserId)
	router.GET("/GetTagsByUserIdAndReminderId/:id/:reminderId", GetTagsByUserIdAndReminderId)
	router.POST("/deleteTag", deleteTag)

	// Recordatorios
	router.GET("/GetReminders/:id", GetRemindersByUserId)
	router.GET("/GetRemindersTags/:id", GetRemindersTagsByUserId)
	router.POST("/addReminder", addReminder)
	router.POST("/updateReminder", updateReminderById)
	router.POST("/deleteOrRecoverReminder", deleteOrRecoverReminder)

	// Notificaciones y correos
	router.GET("/GetNotifications/:id", GetNotificaciones)
	router.POST("/addNotification", addNotificacion)
	router.POST("/muteNotification", muteNotification)
	router.POST("/addCorreo", addCorreo)

	// Importar horario
	router.POST("/importSchedule", importSchedule)

	// Configuración de usuario
	router.GET("/GetUserInfo/:id", GetUserInfo)

	// LDAP
	router.POST("/auth", Auth)
	router.POST("/addauthuser", createUser)
	router.POST("/addadmin", createAdmin)
	router.POST("/changepassword", changeusrpasswd)

	// Token
	router.POST("/receiveTokenData", receiveTokenData)
	router.POST("/getToken", getToken)

}

func registerV1Routes(router gin.IRouter) {

	autho := JWTManager{Secret: []byte(os.Getenv("JWT_SECRET"))}
	protected := router.Group("/")
	protected.Use(autho.AuthMiddleware())
	{
		// Official schedules
		protected.GET("/course-types", GetTiposCurso)
		protected.GET("/schedules/official/users/:id", getOfficialScheduleByUserId)
		protected.POST("/schedules/activities/times", getActivitiesTimesData)

		// Schedule import
		protected.POST("/schedules/import", importSchedule)

		//	Academic periods
		protected.GET("/academic-periods", getAcademicPeriods)
		protected.POST("/academic-periods/insert", addAcademicPeriod)
		protected.POST("/academic-periods/update", updateAcademicPeriod)
		protected.POST("/academic-periods/delete", deleteAcademicPeriod)

		// Personal comments
		protected.GET("/comments/personal/users/:id", getPersonalCommentsByUserId)
		protected.GET("/comments/personal/users/:id/courses/:idCourse", getPersonalCommentsByUserIdAndCourseId)
		protected.POST("/comments/personal", addPersonalComment)
		protected.POST("/comments/personal/update", updatePersonalComment)
		protected.POST("/comments/personal/delete", deletePersonalComment)

		// Personal schedules
		protected.GET("/schedules/personal/users/:id", getPersonalScheduleByUserId)
		protected.POST("/schedules/personal", addPersonalActivity)
		protected.POST("/schedules/personal/update", updatePersonalScheduleByIdCourse)
		protected.POST("/schedules/personal/delete-or-recover", deleteOrRecoveryPersonalScheduleByIdCourse)

		// Tags
		protected.GET("/tags/users/:id", GetTagsByUserId)
		protected.GET("/tags/users/:id/reminders/:reminderId", GetTagsByUserIdAndReminderId)
		protected.POST("/tags/delete", deleteTag)

		// Reminders
		protected.GET("/reminders/users/:id", GetRemindersByUserId)
		protected.GET("/reminders/users/:id/tags", GetRemindersTagsByUserId)
		protected.POST("/reminders", addReminder)
		protected.POST("/reminders/update", updateReminderById)
		protected.POST("/reminders/delete-or-recover", deleteOrRecoverReminder)
		protected.POST("/reminders/delete/multiple", deleteMultipleReminder)

		// Notifications and emails
		protected.GET("/notifications/users/:id", GetNotificaciones)
		protected.POST("/notifications", addNotificacion)

		// User configuration
		protected.GET("/users/:id", GetUserInfo)
		protected.POST("/notifications/mute", muteNotification)

		// Paleta de colores
		protected.POST("/palette", receivePaletteData)
		protected.POST("/palette/get", getPalette)
	}

	// Notifications and emails
	router.POST("/notifications/delete", deleteNotifications)
	router.POST("/emails", addCorreo)

	// Registro de incorporación
	router.POST("/onboarding", receiveOnboardingStatus)
	router.POST("/onboarding/get", getOnboardingStatus)

	// LDAP/auth
	router.POST("/auth/login", Auth)
	router.POST("/auth/users", createUser)
	router.POST("/auth/admins", createAdmin)
	router.POST("/auth/change-password", changeusrpasswd)

	// Tokens
	router.POST("/tokens", receiveTokenData)
	router.POST("/tokens/get", getToken)

}

func method(c *gin.Context) {}

// c *gin.Context essential for method in GET/POST actions
