package main

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

// ESTO ES PARA LAS COLISIONES
type CheckActivitiesTimesData struct {
	T_idUsuario int `json:"idUsuario"`
	N_dia       int `json:"dia"`
}

type ActivitiesTimesData struct {
	N_iduser    int     `json:"iduser"`
	N_idcourse  int     `json:"idcourse"`
	N_dia       int     `json:"dia"`
	StartHour   *string `json:"StartHour"`
	EndHour     *string `json:"EndHour"`
	FechaInicio *string `json:"FechaInicio"`
	FechaFinal  *string `json:"FechaFinal"`
	IsDeleted   *bool   `json:"IsDeleted"`
}

type AcademicPeriod struct {
	N_idPeriodoAcademico int    `json:"idPeriodoAcademico"`
	T_nombre             string `json:"nombre"`
	Dt_fechaInicio       string `json:"fechaInicio"`
	Dt_fechaFinal        string `json:"fechaFinal "`
}
type NewAcademicPeriod struct {
	N_idUsuario    int    `json:"idUsuario"`
	T_nombre       string `json:"nombre"`
	Dt_fechaInicio string `json:"fechaInicio"`
	Dt_fechaFinal  string `json:"fechaFinal"`
}

type OfficialSchedule struct {
	N_idHorario            int             `json:"N_idHorario"`
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
	N_idPeriodoAcademico   int             `json:"IdPeriodoAcademico"`
	Periodo_academico      string          `json:"PeriodoAcademico"`
	FechaInicio            string          `json:"FechaInicio"`
	FechaFinal             string          `json:"FechaFinal"`
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
	N_idUsuario      int           `json:"N_idUsuario"`
	N_idRecordatorio int           `json:"N_idRecordatorio"`
	N_idEtiqueta     int           `json:"N_idEtiqueta"`
	T_nombre         string        `json:"T_nombre"`
	B_isDeleted      *sql.NullBool `json:"B_isDeleted"`
}
type DelTag struct {
	N_idEtiqueta int `json:"N_idEtiqueta"`
	P_usuario    int `json:"P_usuario"`
}
type PersonalScheduleNewValue struct {
	NewActivityValue   string `json:"NewActivityValue" binding:"required"`
	IdPersonalSchedule int    `json:"IdPersonalSchedule" binding:"required"`
}
type forDeleteOrRecoveryPersonalSchedule struct {
	IdPersonalSchedule int `json:"IdPersonalSchedule" binding:"required"`
}
type NewPersonalActivity struct {
	P_usuario     int    `json:"P_usuario"`
	P_nombreCurso string `json:"P_nombreCurso"`
	P_descripcion string `json:"P_descripcion"`
	P_fechaInicio string `json:"P_fechaInicio"`
	P_fechaFin    string `json:"P_fechaFin"`
	P_dia         int    `json:"P_dia"`
	P_horaInicio  string `json:"P_horaInicio"`
	P_horaFin     string `json:"P_horaFin"`
	P_periodo     int    `json:"P_periodo"`
}
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
type ofcComments struct {
	N_idHorario     int           `json:"N_idHorario"`
	N_idUsuario     int           `json:"N_idUsuario"`
	N_idCurso       int           `json:"N_idCurso"`
	Curso           string        `json:"Curso"`
	N_idComentarios int           `json:"N_idComentarios"`
	T_comentario    string        `json:"T_comentario"`
	B_isDeleted     *sql.NullBool `json:"B_isDeleted"`
}
type new_ofcComments struct {
	N_idHorario  int    `json:"N_idHorario"`
	N_idUsuario  int    `json:"N_idUsuario"`
	T_comentario string `json:"T_comentario"`
}
type edit_ofcComment struct {
	N_idComentarios int    `json:"N_idComentarios"`
	N_idUsuario     int    `json:"N_idUsuario"`
	T_comentario    string `json:"T_comentario"`
}
type del_ofcComment struct {
	N_idComentarios int `json:"N_idComentarios" binding:"required"`
	N_idUsuario     int `json:"N_idUsuario"`
}
type Reminders struct {
	N_idToDoList        int            `json:"N_idToDoList"`
	N_idUsuario         int            `json:"N_idUsuario"`
	N_idRecordatorio    int            `json:"N_idRecordatorio"`
	T_nombre            string         `json:"T_nombre"`
	T_descripcion       sql.NullString `json:"T_descripcion"`
	Dt_fechaVencimiento sql.NullString `json:"Dt_fechaVencimiento"`
	B_isDeleted         *bool          `json:"B_isDeleted"`
	T_Prioridad         string         `json:"T_Prioridad"`
	B_estado            *bool          `json:"B_estado"`
}
type RemindersTag struct {
	N_idToDoList        int            `json:"N_idToDoList"`
	N_idUsuario         int            `json:"N_idUsuario"`
	N_idRecordatorio    int            `json:"N_idRecordatorio"`
	T_nombre            string         `json:"T_nombre"`
	T_descripcion       sql.NullString `json:"T_descripcion"`
	Dt_fechaVencimiento sql.NullString `json:"Dt_fechaVencimiento"`
	B_isDeleted         *bool          `json:"B_isDeleted"`
	T_Prioridad         string         `json:"T_Prioridad"`
	B_estado            *bool          `json:"B_estado"`
	N_idEtiqueta        *int           `json:"N_idEtiqueta"`
	T_tag_nombre        *string        `json:"T_tag_nombre"`
	B_tag_isDeleted     *bool          `json:"B_tag_isDeleted"`
}
type ReminderNewValue struct {
	P_usuario     int     `json:"P_usuario"`
	P_nombre      string  `json:"P_nombre"`
	P_descripcion string  `json:"P_descripcion"`
	P_fecha       string  `json:"P_fecha"`
	P_prioridad   int     `json:"P_prioridad"`
	P_estado      *bool   `json:"P_estado"`
	P_tag1        *string `json:"P_tag1"`
	P_tag2        *string `json:"P_tag2"`
	P_tag3        *string `json:"P_tag3"`
	P_tag4        *string `json:"P_tag4"`
	P_tag5        *string `json:"P_tag5"`
}
type EditReminder struct {
	P_usuario     int    `json:"P_usuario"`
	P_idToDo      int     `json:"P_idToDo"`
	P_nombre      *string `json:"P_nombre"`
	P_descripcion *string `json:"P_descripcion"`
	P_fecha       *string `json:"P_fecha"`
	P_prioridad   *int    `json:"P_prioridad"`
	P_estado      *bool   `json:"P_estado"`
	P_tag1        *string `json:"P_tag1"`
	P_tag2        *string `json:"P_tag2"`
	P_tag3        *string `json:"P_tag3"`
	P_tag4        *string `json:"P_tag4"`
	P_tag5        *string `json:"P_tag5"`
}
type DelReminder struct {
	N_idRecordatorio int `json:"N_idRecordatorio"`
	P_usuario        int `json:"P_usuario"`
}
type MultiDelReminder struct {
	N_idRecordatorios string `json:"N_idRecordatorios"`
	P_usuario         int    `json:"P_usuario"`
}
type TipoCurso struct {
	N_idTipoCurso int    `json:"N_idTipoCurso"`
	T_nombre      string `json:"T_nombre"`
	B_isDeleted   *bool  `json:"B_isDeleted"`
}
type Notificacion struct {
	N_idNotificacion int    `json:"idNotificacion"`
	N_idUsuario      int    `json:"idUsuario"`
	N_idRecordatorio int    `json:"idRecordatorio"`
	T_nombre         string `json:"nombre"`
	T_descripcion    string `json:"descripcion"`
	Dt_fechaEmision  string `json:"fechaEmision"`
	B_estado         string `json:"estado"`
}
type NewNotificacion struct {
	T_nombre        string `json:"nombre"`
	T_descripcion   string `json:"descripcion"`
	Dt_fechaEmision string `json:"fechaEmision"`
	N_idToDoList    int    `json:"idToDoList"`
	N_idUsuario     int     `json:"N_idUsuario"`
}
type MuteNotification struct {
	P_idUsuario       int     `json:"idUsuario"`
	P_correo          *string `json:"correo"`
	P_antelacionNotis *string `json:"antelacionNotis"`
	N_idUsuario     int     `json:"N_idUsuario"`
}
type NewCorreo struct {
	T_asunto        string `json:"asunto"`
	T_contenido     string `json:"contenido"`
	Dt_fechaEmision string `json:"fechaEmision"`
	N_idToDoList    int    `json:"idToDoList"`
	N_idUsuario     int     `json:"N_idUsuario"`
}
type UserData struct {
	N_idUsuario        int     `json:"idUsuario"`
	T_nombre           *string `json:"nombre"`
	T_correo           *string `json:"correo"`
	N_semestreActual   *int    `json:"semestreActual"`
	T_programa         *string `json:"programa"`
	TM_antelacionNotis *string `json:"antelacionNotis"`
	N_celular          *string `json:"celular"`
}

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

type Token struct {
	UserId string `json:"userId"`
	Token  string `json:"token"`
}

type DeleteNotification struct {
	Ids string `json:"ids"`
	N_idUsuario     int     `json:"N_idUsuario"`
}

type Palette struct {
	UserId  string `json:"userId"`
	Palette string `json:"palette"`
}

type Onboarding struct {
	UserId string `json:"userId"`
	Status string `json:"status"`
}
