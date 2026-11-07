package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/router/loopback"
	"webtyp.com/view"
)

func TestFloor_ListAcceptsListArgs(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))

	f1, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})
	if err != nil {
		t.Fatal(err)
	}
	f2, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 2"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 3"})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Limit: 1 -> exactly 1 row
	var list1 roomlayout.FloorList
	var callErr error
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListFloors, &view.ListArgs{Limit: 1}, &list1, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list limit 1 failed: %v", callErr)
	}
	if len(list1) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list1))
	}
	firstID := list1[0].Id
	if firstID != f1.Id {
		t.Fatalf("expected first floor id %s, got %s", f1.Id, firstID)
	}

	// 2. Search exact row -> that row
	var searchList roomlayout.FloorList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListFloors, &view.ListArgs{Search: "Piso 2"}, &searchList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list search failed: %v", callErr)
	}
	if len(searchList) != 1 || searchList[0].Name != "Piso 2" || searchList[0].Id != f2.Id {
		t.Fatalf("expected to find 'Piso 2', got %+v", searchList)
	}

	// 3. After the first row -> first row is not in result
	var afterList roomlayout.FloorList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListFloors, &view.ListArgs{After: firstID, Limit: 10}, &afterList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list after failed: %v", callErr)
	}
	if len(afterList) != 2 {
		t.Fatalf("expected 2 rows after first, got %d", len(afterList))
	}
	for _, f := range afterList {
		if f.Id == firstID {
			t.Fatalf("first row %s should not be in results after first row", firstID)
		}
	}
}

func TestRoom_ListAcceptsListArgs(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))

	fl, _ := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Piso 1"})

	r1, err := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: fl.Id, Code: "R101", Name: "Room 101", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: fl.Id, Code: "R102", Name: "Room 102", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.mod.SaveRoom(roomlayout.Room{TenantId: "tenantA", FloorId: fl.Id, Code: "R103", Name: "Room 103", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Limit: 1 -> exactly 1 row
	var list1 roomlayout.RoomList
	var callErr error
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &view.ListArgs{Limit: 1}, &list1, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list rooms limit 1 failed: %v", callErr)
	}
	if len(list1) != 1 {
		t.Fatalf("expected 1 room, got %d", len(list1))
	}
	firstID := list1[0].Id
	if firstID != r1.Id {
		t.Fatalf("expected first room id %s, got %s", r1.Id, firstID)
	}

	// 2. Search exact code -> that row
	var searchList roomlayout.RoomList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &view.ListArgs{Search: "R102"}, &searchList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list search failed: %v", callErr)
	}
	if len(searchList) != 1 || searchList[0].Code != "R102" || searchList[0].Id != r2.Id {
		t.Fatalf("expected to find 'R102', got %+v", searchList)
	}

	// 3. After the first row -> first row is not in result
	var afterList roomlayout.RoomList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRooms, &view.ListArgs{After: firstID, Limit: 10}, &afterList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list after failed: %v", callErr)
	}
	if len(afterList) != 2 {
		t.Fatalf("expected 2 rooms after first, got %d", len(afterList))
	}
	for _, r := range afterList {
		if r.Id == firstID {
			t.Fatalf("first room %s should not be in results after first room", firstID)
		}
	}
}

func TestEquipment_ListAcceptsListArgs(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))

	eq1, err := env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Ecógrafo"})
	if err != nil {
		t.Fatal(err)
	}
	eq2, err := env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Rayos X"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.mod.SaveEquipment(roomlayout.Equipment{TenantId: "tenantA", Name: "Sillón"})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Limit: 1 -> exactly 1 row
	var list1 roomlayout.EquipmentList
	var callErr error
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &view.ListArgs{Limit: 1}, &list1, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list equipment limit 1 failed: %v", callErr)
	}
	if len(list1) != 1 {
		t.Fatalf("expected 1 equipment, got %d", len(list1))
	}
	firstID := list1[0].Id
	if firstID != eq1.Id {
		t.Fatalf("expected first equipment id %s, got %s", eq1.Id, firstID)
	}

	// 2. Search exact name -> that row
	var searchList roomlayout.EquipmentList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &view.ListArgs{Search: "Rayos X"}, &searchList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list search failed: %v", callErr)
	}
	if len(searchList) != 1 || searchList[0].Name != "Rayos X" || searchList[0].Id != eq2.Id {
		t.Fatalf("expected to find 'Rayos X', got %+v", searchList)
	}

	// 3. After the first row -> first row is not in result
	var afterList roomlayout.EquipmentList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListEquipment, &view.ListArgs{After: firstID, Limit: 10}, &afterList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list after failed: %v", callErr)
	}
	if len(afterList) != 2 {
		t.Fatalf("expected 2 equipment after first, got %d", len(afterList))
	}
	for _, e := range afterList {
		if e.Id == firstID {
			t.Fatalf("first equipment %s should not be in results after first row", firstID)
		}
	}
}

func TestRoomArtifact_ListAcceptsListArgs(t *testing.T) {
	env := newTestEnv("tenantA")
	caller := view.NewRecords().Caller(loopback.WithTenant("tenantA", env.mod))

	a1, err := env.mod.SaveRoomArtifact(roomlayout.RoomArtifact{TenantId: "tenantA", RoomId: "r1", Code: "ART-1", Kind: "pc", Cell: "1,1", Status: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	a2, err := env.mod.SaveRoomArtifact(roomlayout.RoomArtifact{TenantId: "tenantA", RoomId: "r1", Code: "ART-2", Kind: "pc", Cell: "1,2", Status: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.mod.SaveRoomArtifact(roomlayout.RoomArtifact{TenantId: "tenantA", RoomId: "r1", Code: "ART-3", Kind: "chair", Cell: "1,3", Status: "ok"})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Limit: 1 -> exactly 1 row
	var list1 roomlayout.RoomArtifactList
	var callErr error
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomArtifacts, &view.ListArgs{Limit: 1}, &list1, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list artifacts limit 1 failed: %v", callErr)
	}
	if len(list1) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(list1))
	}
	firstID := list1[0].Id
	if firstID != a1.Id {
		t.Fatalf("expected first artifact id %s, got %s", a1.Id, firstID)
	}

	// 2. Search exact code -> that row
	var searchList roomlayout.RoomArtifactList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomArtifacts, &view.ListArgs{Search: "ART-2"}, &searchList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list search failed: %v", callErr)
	}
	if len(searchList) != 1 || searchList[0].Code != "ART-2" || searchList[0].Id != a2.Id {
		t.Fatalf("expected to find 'ART-2', got %+v", searchList)
	}

	// 3. After the first row -> first row is not in result
	var afterList roomlayout.RoomArtifactList
	caller.Call(roomlayout.ModelName+"."+roomlayout.OpListRoomArtifacts, &view.ListArgs{After: firstID, Limit: 10}, &afterList, func(e error) {
		callErr = e
	})
	if callErr != nil {
		t.Fatalf("list after failed: %v", callErr)
	}
	if len(afterList) != 2 {
		t.Fatalf("expected 2 artifacts after first, got %d", len(afterList))
	}
	for _, a := range afterList {
		if a.Id == firstID {
			t.Fatalf("first artifact %s should not be in results after first row", firstID)
		}
	}
}
