package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

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
