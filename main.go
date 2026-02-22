package main

import (
	//"encoding/json"
	//"net/http"
	"crypto/tls"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"unicode/utf16"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var db *sql.DB

//Pruebita
/* Saving the session of MySQL, this is global for the access in all methods */
type User struct {
	Username string
	Roles    []string
}
type UserAuth struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

const clockSkewTolerance = 10 * time.Second

type JWTManager struct {
	Secret []byte
	TTL    time.Duration
	Issuer string
}

type Claims struct {
	UserID string   `json:"sub"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
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
	Activity          string `json:"Activity"`
	Description       string `json:"Description"`
	IdTag             int    `json:"IdTag"`
	Day               int    `json:"Day"`
	StartHour         string `json:"StartHour"`
	EndHour           string `json:"EndHour"`
	N_iduser          int    `json:"N_iduser"`
	Id_AcademicPeriod int    `json:"Id_AcademicPeriod"`
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
	//"../../config/goapiconfig.env"
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
	router.GET("/GetOfficialScheduleByUserId/:id", getOfficialScheduleByUserId)
	router.GET("/GetPersonalScheduleByUserId/:id", getPersonalScheduleByUserId)
	router.GET("/GetPersonalComments/:id", getPersonalCommentsByUserIdAndCourseId)
	router.GET("/GetTags", getTags)
	router.POST("/updateNameOfPersonalScheduleByIdCourse", updateNameOfPersonalScheduleByIdCourse)
	router.POST("/updateDescriptionOfPersonalScheduleByIdCourse", updateDescriptionOfPersonalScheduleByIdCourse)
	router.POST("/updateStartHourOfPersonalScheduleByIdCourse", updateStartHourOfPersonalScheduleByIdCourse)
	router.POST("/updateEndHourOfPersonalScheduleByIdCourse", updateEndHourOfPersonalScheduleByIdCourse)
	router.POST("/deleteOrRecoveryPersonalScheduleByIdCourse", deleteOrRecoveryPersonalScheduleByIdCourse)
	router.POST("/addPersonalActivity", addPersonalActivity)
	router.POST("/addPersonalComment", addPersonalComment)
	router.POST("/auth", auth)
	router.POST("/addauthuser", createUser)

	router.Run("0.0.0.0:3913") // The port number for expone the API

}
func method(c *gin.Context) {}

// c *gin.Context essential for method in GET/POST actions

/* This function is a basic get for get the users from database */

func getOfficialScheduleByUserId(c *gin.Context) {
	id := c.Param("id")

	rows, err := db.Query(`SELECT ao.* FROM ActividadesOficiales ao JOIN Usuarios u ON ao.N_idUsuario = u.N_idUsuario WHERE u.T_codUsuario = ?`, id)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var ofcschedules []OfficialSchedule

	for rows.Next() {
		var ofcschedule OfficialSchedule
		err := rows.Scan(
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
		ofcschedules = append(ofcschedules, ofcschedule)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

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
func updateNameOfPersonalScheduleByIdCourse(c *gin.Context) {
	var newValue PersonalScheduleNewValue
	err := c.BindJSON(&newValue)
	if err != nil {
		c.JSON(400, gin.H{"Palurdo": "formato invalido de json"})
		return
	}
	result, err := db.Exec("UPDATE ActividadesPersonales SET Actividad = ? WHERE N_idCurso= ? ", newValue.NewActivityValue, newValue.IdPersonalSchedule)
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
	err := c.BindJSON(&newPerActivity)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

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

}
func auth(c *gin.Context) {
	var User UserAuth
	err := c.BindJSON(&User)
	token, userU, err2 := ConnectLDAP(User.User, User.Pass, JWTManager{
		Secret: []byte(os.Getenv("JWT_SECRET")),
		TTL:    24 * time.Hour,
		Issuer: "horario_estudiantes",
	})
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	if err2 != nil {
		log.Printf("ldap error: %v", err2)
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
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
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
	l, err := ldap.DialURL("ldap://127.0.0.1:389")
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
