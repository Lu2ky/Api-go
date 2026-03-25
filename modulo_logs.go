package main
import (
	"log"
)


func insertarLog(usuarioID int, accion string, descripcion string) {
	query := `
	INSERT INTO Logs (N_idUsuario, T_accion, T_Descripcion, Dt_fecha)
	VALUES (?, ?, ?, NOW())
	`

	_, err := db.Exec(query, usuarioID, accion, descripcion)
	if err != nil {
		log.Println("Error al insertar log:", err)
		}
}
