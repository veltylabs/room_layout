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

	zone, _ := time.LoadZone("America/Santiago")
	base := roomlayout.Deps{
		IDs:        ids,
		Categories: cats,
		Occupants:  occs,
		TenantID:   "tenant1",
		Zone:       zone,
	}

	// 1. Missing IDs
	d := base
	d.IDs = nil
	_, err := roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.IDs is required" {
		t.Fatalf("expected Deps.IDs error, got %v", err)
	}

	// 2. Missing Categories
	d = base
	d.Categories = nil
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Categories is required" {
		t.Fatalf("expected Deps.Categories error, got %v", err)
	}

	// 3. Missing Occupants
	d = base
	d.Occupants = nil
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Occupants is required" {
		t.Fatalf("expected Deps.Occupants error, got %v", err)
	}

	// 4. Missing TenantID
	d = base
	d.TenantID = ""
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.TenantID is required" {
		t.Fatalf("expected Deps.TenantID error, got %v", err)
	}

	// 5. Missing Zone
	d = base
	d.Zone = time.Zone{}
	_, err = roomlayout.New(db, d)
	if err == nil || err.Error() != "room_layout: Deps.Zone is required" {
		t.Fatalf("expected Deps.Zone error, got %v", err)
	}
}
