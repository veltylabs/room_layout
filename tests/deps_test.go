package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

func TestDependencyValidation(t *testing.T) {
	db := orm.New(mem.New())
	ids := &simpleIDs{}
	cats := newFakeCategories()
	occs := newFakeOccupants()

	base := roomlayout.Deps{
		IDs:        ids,
		Categories: cats,
		Occupants:  occs,
		TenantID:   "tenant1",
		Timezone:   "America/Santiago",
	}

	// 1. Missing IDs
	d := base
	d.IDs = nil
	_, err := roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.IDs no supplied_1" {
		t.Fatalf("expected Deps.IDs error, got %v", err)
	}

	// 2. Missing Categories
	d = base
	d.Categories = nil
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Categories req" {
		t.Fatalf("expected Deps.Categories error, got %v", err)
	}

	// 3. Missing Occupants
	d = base
	d.Occupants = nil
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Occupants req" {
		t.Fatalf("expected Deps.Occupants error, got %v", err)
	}

	// 4. Missing TenantID
	d = base
	d.TenantID = ""
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.TenantID no supplied_3" {
		t.Fatalf("expected Deps.TenantID error, got %v", err)
	}

	// 5. Missing Timezone
	d = base
	d.Timezone = ""
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Timezone no supplied_4" {
		t.Fatalf("expected Deps.Timezone error, got %v", err)
	}
}
