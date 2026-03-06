package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

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

func getActivitiesTimesData(c *gin.Context) {
	var checkActTime CheckActivitiesTimesData

	err := c.BindJSON(&checkActTime)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}

	rows, err := db.Query(
		`
		SELECT * FROM HorarioCompleto 
		WHERE N_idUsuario = (SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?) AND N_dia = ?
		`, checkActTime.T_codUsuario, checkActTime.N_dia)

	if err != nil {
		log.Printf("Database error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	defer rows.Close()

	//	Aquí se van a almacenar los resultados de la consulta.
	//	Se utiiza el OfficialSchedule para tener una estructura a la hora de guardar la información de la consulta.

	var actTimeArr []ActivitiesTimesData

	for rows.Next() {
		var actTime ActivitiesTimesData
		err := rows.Scan(
			&actTime.N_iduser,
			&actTime.N_idcourse,
			&actTime.N_dia,
			&actTime.StartHour,
			&actTime.EndHour,
			&actTime.FechaInicio,
			&actTime.FechaFinal,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Error en procesamiento de datos"})
			return
		}

		//	Y aquí se agrega el objeto ofcschedule al arreglo ofcschedules.
		actTimeArr = append(actTimeArr, actTime)
	}

	//	Se verifica si hubo errores mientras se hizo la iteración usando rows.Err().
	//	Si Next() retorna False, entonces para revisar cuál fue el error se usa rows.Err()
	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Error leyendo resultados"})
		return
	}

	//	Se retorna con código 200 (OK status) el arreglo formando anteriormente en formato JSON.
	c.JSON(200, actTimeArr)
}
