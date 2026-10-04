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

	// Test Render
	rootView, ok := module.View().(*ui.RootView)
	if !ok {
		t.Fatalf("expected module.View() to be *ui.RootView")
	}

	el := rootView.Render()
	if el == nil {
		t.Fatalf("Render() returned nil")
	}

	// Test stage transitions
	for stage := 0; stage <= 3; stage++ {
		rootView.GoToStage(stage)
		if rootView.ActiveStage != stage {
			t.Fatalf("expected ActiveStage %d, got %d", stage, rootView.ActiveStage)
		}
	}
}

func TestCoordinateAndSliceHelpers(t *testing.T) {
	// Test FormatCell and ParseCell
	tests := []struct {
		r   int
		c   int
		str string
	}{
		{0, 0, "A1"},
		{0, 1, "B1"},
		{1, 0, "A2"},
		{1, 2, "C2"},
		{13, 19, "T14"},
		{0, 25, "Z1"},
		{0, 26, "AA1"},
	}

	for _, tt := range tests {
		gotStr := ui.FormatCell(tt.r, tt.c)
		if gotStr != tt.str {
			t.Fatalf("FormatCell(%d, %d): expected %s, got %s", tt.r, tt.c, tt.str, gotStr)
		}
		r, c, ok := ui.ParseCell(tt.str)
		if !ok || r != tt.r || c != tt.c {
			t.Fatalf("ParseCell(%s): expected (%d, %d), got (%d, %d)", tt.str, tt.r, tt.c, r, c)
		}
	}

	// Test Slice Helpers (strictly without maps)
	cells := []string{"B2", "B3", "C2"}
	if !ui.CellInList(cells, "B2") {
		t.Fatalf("expected B2 in cells")
	}
	if ui.CellInList(cells, "A1") {
		t.Fatalf("expected A1 not in cells")
	}

	added := ui.AddCell(cells, "C3")
	if len(added) != 4 || !ui.CellInList(added, "C3") {
		t.Fatalf("expected C3 added")
	}
	// Adding again should not duplicate
	addedAgain := ui.AddCell(added, "C3")
	if len(addedAgain) != 4 {
		t.Fatalf("expected duplicate not added")
	}

	removed := ui.RemoveCell(added, "B2")
	if len(removed) != 3 || ui.CellInList(removed, "B2") {
		t.Fatalf("expected B2 removed")
	}

	// Test Overlap
	setA := []string{"B2", "B3"}
	setB := []string{"B3", "B4"}
	setC := []string{"E1", "E2"}
	if !ui.CellsOverlap(setA, setB) {
		t.Fatalf("expected overlap between setA and setB")
	}
	if ui.CellsOverlap(setA, setC) {
		t.Fatalf("expected no overlap between setA and setC")
	}

	// Test BoundingBox
	minC, minR, maxC, maxR := ui.BoundingBox([]string{"B2", "D4", "C3"})
	if minC != 1 || minR != 1 || maxC != 3 || maxR != 3 {
		t.Fatalf("unexpected bounding box: (%d,%d)-(%d,%d)", minC, minR, maxC, maxR)
	}

	label := ui.BoundingBoxLabel([]string{"B2", "D4", "C3"})
	if label != "B2:D4 · 3 celdas" {
		t.Fatalf("unexpected label: %s", label)
	}
}
