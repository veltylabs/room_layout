package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/model"
	"webtyp.com/router"
)

type mockRoute struct {
	name       string
	resource   model.Resource
	action     model.Action
	isPublic   bool
	argsSchema model.Fielder
}

func (r *mockRoute) Requires(res model.Resource, act model.Action) router.Route {
	r.resource = res
	r.action = act
	return r
}

func (r *mockRoute) Authenticated() router.Route {
	return r
}

func (r *mockRoute) Public() router.Route {
	r.isPublic = true
	return r
}

func (r *mockRoute) Accepts(args model.Fielder) router.Route {
	r.argsSchema = args
	return r
}

type mockRegistry struct {
	routes []*mockRoute
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{}
}

func (m *mockRegistry) findRoute(name string) *mockRoute {
	for _, r := range m.routes {
		if r.name == name {
			return r
		}
	}
	return nil
}

func (m *mockRegistry) Operation(name string, h router.HandlerFunc) router.Route {
	r := &mockRoute{name: name}
	m.routes = append(m.routes, r)
	return r
}

func TestMountOperationsAndPublicCheck(t *testing.T) {
	env := newTestEnv("tenantA")
	reg := newMockRegistry()

	env.mod.MountOperations(reg)

	expectedOps := []string{
		roomlayout.OpListFloors,
		roomlayout.OpSaveFloor,
		roomlayout.OpDeleteFloor,
		roomlayout.OpListRooms,
		roomlayout.OpGetRoom,
		roomlayout.OpSaveRoom,
		roomlayout.OpDeactivateRoom,
		roomlayout.OpListEquipment,
		roomlayout.OpSaveEquipment,
		roomlayout.OpDeleteEquipment,
		roomlayout.OpListRoomEquipment,
		roomlayout.OpSetRoomEquipment,
		roomlayout.OpListCategories,
		roomlayout.OpListRoomCategories,
		roomlayout.OpSetRoomCategories,
		roomlayout.OpListOccupants,
		roomlayout.OpListRoomShifts,
		roomlayout.OpSaveRoomShift,
		roomlayout.OpDeleteRoomShift,
		roomlayout.OpCancelShiftOccurrence,
		roomlayout.OpMoveShiftOccurrence,
		roomlayout.OpListRoomDay,
		roomlayout.OpListBoard,
		roomlayout.OpFindFreeRooms,
		roomlayout.OpListPublicAvailability,
	}

	for _, op := range expectedOps {
		r := reg.findRoute(op)
		if r == nil {
			t.Fatalf("expected operation '%s' to be mounted", op)
		}
		if op == roomlayout.OpListPublicAvailability {
			if !r.isPublic {
				t.Fatalf("expected opListPublicAvailability to be public")
			}
		} else {
			if r.isPublic {
				t.Fatalf("expected op '%s' NOT to be public", op)
			}
		}
	}
}

func TestEventPublishingOnMutations(t *testing.T) {
	env := newTestEnv("tenantA")

	f, err := env.mod.SaveFloor(roomlayout.Floor{TenantId: "tenantA", Name: "Floor 1"})
	if err != nil {
		t.Fatalf("SaveFloor failed: %v", err)
	}

	if len(env.publisher.events) != 1 || env.publisher.events[0].Topic != roomlayout.TopicFloorSaved {
		t.Fatalf("expected event TopicFloorSaved, got %+v", env.publisher.events)
	}

	r, err := env.mod.SaveRoom(roomlayout.Room{
		TenantId: "tenantA",
		FloorId:  f.Id,
		Code:     "R101",
		Name:     "Room 101",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("SaveRoom failed: %v", err)
	}

	if len(env.publisher.events) != 2 || env.publisher.events[1].Topic != roomlayout.TopicRoomSaved {
		t.Fatalf("expected event TopicRoomSaved, got %+v", env.publisher.events)
	}

	_ = env.mod.DeactivateRoom("tenantA", r.Id)
	if len(env.publisher.events) != 3 || env.publisher.events[2].Topic != roomlayout.TopicRoomDeactivated {
		t.Fatalf("expected event TopicRoomDeactivated, got %+v", env.publisher.events)
	}
}
