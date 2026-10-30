package migrate

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/ddl"
	"webtyp.com/fmt"
)

// errNoCompiler: sin compilador de SQL no hay cómo crear las tablas. Es un error de programación,
// dato y no texto a traducir: por eso es variable (langc traduce las constantes de fmt.Err).
var errNoCompiler = "room_layout/migrate: the connection does not compile SQL (ddl.Compiler); pass one"

// Migrate crea (o completa) todas las tablas del módulo. La app lo llama desde su cmd/migrate.
//
// Ejemplo: err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, c ddl.Compiler) error {
	if c == nil {
		if compiler, ok := conn.(ddl.Compiler); ok {
			c = compiler
		}
	}
	if c == nil {
		return fmt.Err(errNoCompiler)
	}
	return ddl.New(conn, c).Sync(
		&roomlayout.Floor{},
		&roomlayout.Room{},
		&roomlayout.RoomArtifact{},
		&roomlayout.Equipment{},
		&roomlayout.RoomEquipment{},
		&roomlayout.RoomCategory{},
		&roomlayout.RoomShift{},
		&roomlayout.RoomShiftCancellation{},
	)
}
