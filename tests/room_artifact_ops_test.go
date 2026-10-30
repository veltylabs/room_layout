package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/model"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

// El mapa guarda, lista y borra los artefactos por las operaciones del módulo, y el tenant lo pone
// el servidor aunque la petición traiga otro.
func TestRoomArtifactOps_SaveListDelete(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))
	arts := view.NewCallerLister(caller,
		view.Ops{Module: roomlayout.ModelName, List: roomlayout.OpListRoomArtifacts, Save: roomlayout.OpSaveRoomArtifact, Delete: roomlayout.OpDeleteRoomArtifact},
		func() model.ModelSlice { return &roomlayout.RoomArtifactList{} })

	var lastErr error
	done := func(e error) { lastErr = e }
	a := &roomlayout.RoomArtifact{TenantId: "other-tenant", RoomId: "rm-1", Kind: "pc", Code: "pc11", Cell: "12,4", Status: "ok"}
	arts.(view.Saver).Save([]model.Model{a}, done)
	if lastErr != nil {
		t.Fatalf("save artifact: %v", lastErr)
	}

	stored, err := env.mod.ListRoomArtifacts("tenantA", "")
	if err != nil || len(stored) != 1 || stored[0].Cell != "12,4" {
		t.Fatalf("stored under the server's tenant: got %+v (err %v)", stored, err)
	}
	if other, _ := env.mod.ListRoomArtifacts("other-tenant", ""); len(other) != 0 {
		t.Fatalf("tenant sent by the request got %d artifacts, want 0", len(other))
	}

	arts.(view.Deleter).Delete([]string{stored[0].Id}, done)
	if lastErr != nil {
		t.Fatalf("delete artifact: %v", lastErr)
	}
	if left, _ := env.mod.ListRoomArtifacts("tenantA", ""); len(left) != 0 {
		t.Fatalf("after delete: %d artifacts left, want 0", len(left))
	}
}
