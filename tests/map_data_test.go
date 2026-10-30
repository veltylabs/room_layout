package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

// El mapa abre con lo que hay en el servidor: pisos con sus celdas, boxes y equipos.
func TestMap_LoadsServerData(t *testing.T) {
	env := newTestEnv("tenantA")
	f, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1", Position: 1, Cols: 20, Rows: 14, HabitableCells: "A1,B1"})
	if err != nil {
		t.Fatal(err)
	}
	rm, err := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: f.Id, Code: "B-101", Name: "Box 1", RoomType: "box", Cells: "A1", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.mod.SaveRoomArtifact(roomlayout.RoomArtifact{TenantId: "tenantA", RoomId: rm.Id, Kind: "pc", Code: "pc11", Cell: "A1", Status: "ok"}); err != nil {
		t.Fatal(err)
	}

	m := ui.NewMap(view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod)))
	m.Init(nil)

	if len(m.Floors) != 1 || len(m.Floors[0].HabitableCells) != 2 || m.Floors[0].HabitableCells[1] != "B1" {
		t.Fatalf("floors: %+v", m.Floors)
	}
	if m.Cols != 20 || m.Rows != 14 {
		t.Fatalf("grid: %dx%d, want 20x14", m.Cols, m.Rows)
	}
	if len(m.Rooms) != 1 || len(m.Rooms[0].Cells) != 1 || m.Rooms[0].Cells[0] != "A1" {
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
	m.Init(nil)

	m.AddFloor("Piso 1")

	stored, err := env.mod.ListFloors("tenantA")
	if err != nil || len(stored) != 1 || stored[0].Name != "Piso 1" {
		t.Fatalf("stored floors: %+v (err %v)", stored, err)
	}
	if len(m.Floors) != 1 || m.Floors[0].ID != stored[0].Id {
		t.Fatalf("map floors: %+v, want the server's id %q", m.Floors, stored[0].Id)
	}
}
