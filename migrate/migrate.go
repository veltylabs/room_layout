package migrate

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/ddl"
)

func Migrate(conn ddl.Execer, c ddl.Compiler) error {
	if c == nil {
		if compiler, ok := conn.(ddl.Compiler); ok {
			c = compiler
		}
	}
	if c == nil {
		return nil
	}
	return ddl.New(conn, c).Sync(
		&roomlayout.Floor{},
		&roomlayout.Room{},
		&roomlayout.Equipment{},
		&roomlayout.RoomEquipment{},
		&roomlayout.RoomCategory{},
		&roomlayout.RoomShift{},
		&roomlayout.RoomShiftCancellation{},
	)
}
