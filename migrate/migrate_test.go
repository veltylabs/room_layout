package migrate_test

import (
	"testing"

	"github.com/veltylabs/room_layout/migrate"
	"webtyp.com/ddl"
)

// Migrate crea todas las tablas del módulo; una que falte aparece recién en producción
// («relation room_artifact does not exist»), así que se comprueba aquí, tabla por tabla.
func TestMigrate_CreatesEveryTable(t *testing.T) {
	tables, err := ddl.Capture([]ddl.Step{{Name: "room_layout", Run: migrate.Migrate}})
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	want := []string{"floor", "room", "room_artifact", "equipment", "room_equipment", "room_category", "room_shift", "room_shift_cancellation"}
	for _, name := range want {
		found := false
		for _, tb := range tables {
			if tb.Model.ModelName() == name {
				found = true
			}
		}
		if !found {
			t.Errorf("Migrate does not create table %q", name)
		}
	}
}

// Sin compilador de SQL no hay cómo crear las tablas: Migrate lo dice, no termina en silencio.
func TestMigrate_NoCompilerIsAnError(t *testing.T) {
	if err := migrate.Migrate(nil, nil); err == nil {
		t.Fatal("Migrate with no compiler must return an error")
	}
}
