package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// -------------------------- IMPORTAR HORARIO ----------------------------------

func importSchedule(c *gin.Context) {
	var newScheduleValue ImportSchedule

	err := c.BindJSON(&newScheduleValue)
	if err != nil {
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

	// Borrar registro de horario oficial de redis
	deleted, err2 := rdb.Del(ctx, "OfficialSchedule:"+newScheduleValue.CodUsuario).Result()

	if err2 != nil {
		fmt.Printf("\nError de conexión: %v", err2)

	} else if deleted > 0 {
		fmt.Printf("\nRegistro eliminado con éxito")
	} else {
		fmt.Printf("\nNo se encontró registro relacionado")
	}

	result, err := db.Exec("CALL importarHorario(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
		newScheduleValue.Nombre,
		newScheduleValue.Semestre,
		newScheduleValue.Programa,
		newScheduleValue.CodUsuario,
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
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "No se encuentra el archivo a importar"})
		return

	}
	descripcion := "Se importó horario del usuario: " + newScheduleValue.CodUsuario +
		" | Curso: " + newScheduleValue.NombreCurso +
		" | NRC: " + newScheduleValue.Nrc

	var userID int
	err = db.QueryRow(
		"SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?",
		newScheduleValue.CodUsuario,
	).Scan(&userID)

	if err != nil {
		log.Println("Error obteniendo usuario para log:", err)
		userID = 0
	}
	insertarLog(userID, "IMPORTAR_HORARIO", descripcion)
	c.JSON(200, gin.H{
		"message": "Horario importado correctamente",
	})

}
