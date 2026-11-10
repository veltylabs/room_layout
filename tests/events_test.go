package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

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
