package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

//	------------------------ FUNCIONALIDADES DEL USUARIO ------------------------ //

func GetUserInfo(c *gin.Context) {

	id_user := c.Param("id")

	//	Consulta a redis
	val, err := rdb.Get(c.Request.Context(), "UserInfo:"+id_user).Result()

	if err == nil {
		fmt.Printf("\n Si existe registro")
		var userDataArray []UserData

		err := json.Unmarshal([]byte(val), &userDataArray)

		if err == nil {
			c.JSON(200, userDataArray)
			return

		}

	}

	// Si no existe en redis, se debe crear la consulta
	fmt.Printf("\n>>>>Creando registro")

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

	// Convertir a formato apto para redis
	data, err := json.Marshal(userDataArray)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error al serializar datos"})
		return
	}
	// Guardar datos en redis
	err2 := rdb.Set(ctx, "UserInfo:"+id_user, data, 48*time.Hour).Err()

	if err2 != nil {
		log.Printf("Error al guardar en Redis: %v", err2)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Devuelve la consulta de la base relacional
	c.JSON(200, userDataArray)

}

// Guardar datos del token en redis
func receiveTokenData(c *gin.Context) {
	var data Token

	// Leer json
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El formato del JSON es incorrecto o faltan campos",
		})
		return
	}

	// Guardar en Redis
	err := rdb.Set(ctx, "reset:"+data.UserId, data.Token, 15*time.Minute).Err()

	if err != nil {
		log.Printf("Error al guardar en Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	// Log
	descripcion := fmt.Sprintf("Token guardado en Redis | Usuario ID: %s", data.UserId)

	go func(uID string, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertLogCod(uID, acc, desc)
	}(data.UserId, "GUARDAR_TOKEN", descripcion)

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Token guardado correctamente en Redis",
	})
}

// Obtener token de redis
func getToken(c *gin.Context) {
	var req Token

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON mal formado"})
		return
	}

	val, err := rdb.Get(c.Request.Context(), "reset:"+req.UserId).Result()

	if err != nil {
		fmt.Printf("Error de Redis: %v\n", err)
		c.JSON(401, gin.H{"error": "Sesión no encontrada o expirada"})
		return
	}

	if val != req.Token {
		c.JSON(401, gin.H{"error": "El token no coincide para este usuario"})
		return
	}

	// Log
	userID, err := strconv.Atoi(req.UserId)
	descripcion := fmt.Sprintf("Se validó el token | Usuario ID: %s", req.UserId)

	go func(uID int, acc, desc string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recuperado de pánico en log (Eliminar): %v", r)
			}
		}()
		insertarLog(uID, acc, desc)
	}(userID, "VALIDAR_TOKEN", descripcion)

	c.JSON(200, gin.H{"userId": req.UserId})
}

// Guardar paleta de colores en redis
func receivePaletteData(c *gin.Context) {
	var data Palette

	// Leer json
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El formato del JSON es incorrecto o faltan campos",
		})
		return
	}

	// Guardar en Redis
	err := rdb.Set(ctx, "palette:"+data.UserId, data.Palette, 0).Err()

	if err != nil {
		log.Printf("Error al guardar en Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}
	userID, err := strconv.Atoi(data.UserId)

	descripcion := "Paleta guardada en Redis | Usuario ID: " + data.UserId

	insertarLog(
		userID,
		"GUARDAR_PALETA",
		descripcion,
	)

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Paleta guardado correctamente en Redis",
	})
}

// Obtener paleta de redis
func getPalette(c *gin.Context) {
	var req Palette

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON mal formado"})
		return
	}

	val, err := rdb.Get(c.Request.Context(), "palette:"+req.UserId).Result()

	if err != nil {
		fmt.Printf("Error de Redis: %v\n", err)
		c.JSON(401, gin.H{"error": "Sesión no encontrada o expirada"})
		return
	}

	c.JSON(200, gin.H{
		"userId":  req.UserId,
		"palette": val,
	})
}

// Guardar registro de haber hecho el tutorial en redis
func receiveOnboardingStatus(c *gin.Context) {
	var data Onboarding

	// Leer json
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "El formato del JSON es incorrecto o faltan campos",
		})
		return
	}

	// Guardar en Redis
	err := rdb.Set(ctx, "onboarding:"+data.UserId, data.Status, 0).Err()

	if err != nil {
		log.Printf("Error al guardar en Redis: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno al guardar en caché",
		})
		return
	}

	userID, err := strconv.Atoi(data.UserId)

	descripcion := "Onboarding actualizado en Redis | Usuario ID: " + data.UserId +
		" | Estado: " + data.Status

	insertarLog(
		userID,
		"GUARDAR_ONBOARDING",
		descripcion,
	)

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Registro de tutorial guardado correctamente en Redis",
	})
}

// Obtener registro de tutorial de redis
func getOnboardingStatus(c *gin.Context) {
	var req Onboarding

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON mal formado"})
		return
	}

	val, err := rdb.Get(c.Request.Context(), "onboarding:"+req.UserId).Result()

	if err != nil {
		fmt.Printf("Error de Redis: %v\n", err)
		c.JSON(401, gin.H{"error": "Sesión no encontrada o expirada"})
		return
	}

	c.JSON(200, gin.H{
		"userId": req.UserId,
		"status": val,
	})
}
