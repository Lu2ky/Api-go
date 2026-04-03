package main
import (
	"log"
)


func insertarLog(usuarioID int, accion string, descripcion string) {
	var query string
	var args []interface{}

	if usuarioID == 0 {
		
		query = `
		INSERT INTO Logs (T_accion, T_Descripcion, Dt_fecha)
		VALUES (?, ?, NOW())
		`
		args = []interface{}{accion, descripcion}
	} else {

		query = `
		INSERT INTO Logs (N_idUsuario, T_accion, T_Descripcion, Dt_fecha)
		VALUES (?, ?, ?, NOW())
		`
		args = []interface{}{usuarioID, accion, descripcion}
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Println("Error al insertar log:", err)
	}
}
