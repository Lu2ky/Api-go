package main

import (
	"log"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

//	------------------------ FUNCIONALIDADES DEL USUARIO ------------------------ //

func GetUserInfo(c *gin.Context) {

	id_user := c.Param("id")

	//	Consulta
	rows, err := db.Query(
		`
		SELECT u.N_idUsuario, u.T_nombre, u.T_correo, u.N_semestreActual, u.T_programa, u.TM_antelacionNotis, u.N_celular
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
			&userData.N_celular,
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

// Guardar datos del token en la base de datos
func receiveTokenData(c *gin.Context) {
	var data NewToken

	// Leer json
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El formato del JSON es incorrecto o faltan campos",
		})
		return
	}

	// Guardar en Redis
	err := rdb.Set(ctx, "reset:"+data.Token, data.UserId, 15*time.Minute).Err()

	if err != nil {
		log.Printf("Error al guardar en Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Token guardado correctamente en Redis",
	})
}

// Obtener token de la base de datos
func getToken(c *gin.Context) {
	var req RequestToken

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Datos inválidos"})
		return
	}

	// Hacer la consulta
	val, err := rdb.Get(c.Request.Context(), req.Token).Result()

	if err != nil {
		c.JSON(404, gin.H{"error": "Token no encontrado"})
		return
	}

	// Devolver el token
	c.JSON(200, gin.H{"userId": val})
}
