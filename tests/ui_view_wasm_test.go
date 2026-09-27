//go:build wasm

package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/seed"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

func TestUIBrowserWASM(t *testing.T) {
	db := orm.New(mem.New())
	ids, _ := unixid.NewUnixID()
	pub := &mock.Broker{}

	cats := newFakeCategories()
	occs := newFakeOccupants()

	deps := roomlayout.Deps{
		IDs:        ids,
		Categories: cats,
		Occupants:  occs,
		TenantID:   "tenantA",
		Timezone:   "America/Santiago",
		Publisher:  pub,
	}

	mod, err := roomlayout.New(db, deps)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	_, err = seed.Load(mod, "tenantA", seed.LoadOptions{
		CategoryIDs: []string{"cat1", "cat2"},
		OccupantIDs: []string{"occ1", "occ2"},
	})
	if err != nil {
		t.Fatalf("seed.Load failed: %v", err)
	}

	loop := loopback.WithTenant("tenantA", mod)

	module, err := ui.Browser(loop, ids, "tenantA")
	if err != nil {
		t.Fatalf("ui.Browser failed: %v", err)
	}
	if module == nil {
		t.Fatalf("expected non-nil UIModule")
	}
}
