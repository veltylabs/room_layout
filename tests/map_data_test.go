package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/layout"
	"webtyp.com/msgtype"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

// El mapa abre con lo que hay en el servidor: pisos con sus celdas, boxes y equipos.
func TestMap_LoadsServerData(t *testing.T) {
	env := newTestEnv("tenantA")
	f, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1", Position: 1, Cols: 20, Rows: 14, HabitableCells: "1,1,2,1"})
	if err != nil {
		t.Fatal(err)
	}
	rm, err := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: f.Id, Code: "B-101", Name: "Box 1", RoomType: "box", Cells: "1,1,1,1", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.mod.SaveRoomArtifact(roomlayout.RoomArtifact{TenantId: "tenantA", RoomId: rm.Id, Kind: "pc", Code: "pc11", Cell: "1,1", Status: "ok"}); err != nil {
		t.Fatal(err)
	}

	m := ui.NewMap(view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod)))
	m.Init(nil)

	if len(m.Floors) != 1 || len(m.Floors[0].HabitableCells) != 2 || m.Floors[0].HabitableCells[1] != "2,1" {
		t.Fatalf("floors: %+v", m.Floors)
	}
	if m.Cols != 20 || m.Rows != 14 {
		t.Fatalf("grid: %dx%d, want 20x14", m.Cols, m.Rows)
	}
	if len(m.Rooms) != 1 || len(m.Rooms[0].Cells) != 1 || m.Rooms[0].Cells[0] != "1,1" {
		t.Fatalf("rooms: %+v", m.Rooms)
	}
	if len(m.Artifacts) != 1 || m.Artifacts[0].FloorID != f.Id {
		t.Fatalf("artifacts: %+v", m.Artifacts)
	}
}

// Un piso nuevo queda guardado en el servidor al crearlo, con el id que pone el servidor.
func TestMap_AddFloorIsSaved(t *testing.T) {
	env := newTestEnv("tenantA")
	m := ui.NewMap(view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod)))
	notices := &notices{}
	m.UseNotifier(notices)
	m.Init(nil)

	m.AddFloor("Piso 1")

	stored, err := env.mod.ListFloors("tenantA")
	if err != nil || len(stored) != 1 || stored[0].Name != "Piso 1" {
		t.Fatalf("stored floors: %+v (err %v)", stored, err)
	}
	if len(m.Floors) != 1 || m.Floors[0].ID != stored[0].Id {
		t.Fatalf("map floors: %+v, want the server's id %q", m.Floors, stored[0].Id)
	}
	if len(notices.types) != 1 || notices.types[0] != msgtype.Success {
		t.Fatalf("notices: %v, want one success from the platform", notices.types)
	}
}

// notices hace de plataforma: guarda el tipo de cada aviso.
type notices struct{ types []msgtype.Type }

func (n *notices) Notify(t msgtype.Type, _ string, _ layout.Duration) { n.types = append(n.types, t) }

// Celdas guardadas en un formato que no es el de roomlayout.EncodeArea se avisan como error.
func TestMap_BadCellsAreAnError(t *testing.T) {
	env := newTestEnv("tenantA")
	if _, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1", HabitableCells: "A1,B1"}); err != nil {
		t.Fatal(err)
	}
	m := ui.NewMap(view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod)))
	notices := &notices{}
	m.UseNotifier(notices)
	m.Init(nil)

	if len(notices.types) != 1 || notices.types[0] != msgtype.Error {
		t.Fatalf("notices: %v, want one error", notices.types)
	}
}

// La cuadrícula no se achica por debajo de lo dibujado: no cambia, avisa y no guarda nada.
func TestMap_GridNeverHidesWhatIsDrawn(t *testing.T) {
	env := newTestEnv("tenantA")
	if _, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1", Cols: 22, Rows: 38, HabitableCells: "1,1,22,38"}); err != nil {
		t.Fatal(err)
	}
	m := ui.NewMap(view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod)))
	notices := &notices{}
	m.UseNotifier(notices)
	m.Init(nil)

	m.ResizeGrid(19, 27)

	stored, _ := env.mod.ListFloors("tenantA")
	if m.Cols != 22 || m.Rows != 38 || stored[0].Cols != 22 || stored[0].Rows != 38 {
		t.Fatalf("grid %dx%d, stored %dx%d; want 22x38 kept", m.Cols, m.Rows, stored[0].Cols, stored[0].Rows)
	}
	if len(notices.types) != 1 || notices.types[0] != msgtype.Warning {
		t.Fatalf("notices: %v, want one warning", notices.types)
	}
}
