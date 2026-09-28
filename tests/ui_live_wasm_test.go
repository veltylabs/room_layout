//go:build wasm

package tests

import (
	"syscall/js"
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"github.com/veltylabs/room_layout/seed"
	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/dom"
	"webtyp.com/events/mock"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

// mountLiveUI renders ui.Browser into a live DOM over the real module, the
// seed and a loopback caller: the stack the demo and the app use. The older
// TestUIBrowserWASM only checked that Browser returned non-nil, which let a
// screen that paints nothing pass.
func mountLiveUI(t *testing.T, rootID string) js.Value {
	t.Helper()
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", rootID)
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	ids, err := unixid.NewUnixID()
	if err != nil {
		t.Fatalf("unixid: %v", err)
	}
	mod, err := roomlayout.New(orm.New(mem.New()), roomlayout.Deps{
		IDs:        ids,
		Categories: newFakeCategories(),
		Occupants:  newFakeOccupants(),
		TenantID:   "tenantA",
		Timezone:   "America/Santiago",
		Publisher:  &mock.Broker{},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := seed.Load(mod, "tenantA", seed.LoadOptions{
		CategoryIDs: []string{"cat1", "cat2"},
		OccupantIDs: []string{"occ1", "occ2"},
	}); err != nil {
		t.Fatalf("seed.Load: %v", err)
	}

	m, err := ui.Browser(loopback.WithTenant("tenantA", mod), ids, "tenantA")
	if err != nil {
		t.Fatalf("ui.Browser: %v", err)
	}
	if err := dom.Render(rootID, m.View()); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}
	return root
}

func countAll(root js.Value, selector string) int {
	return root.Call("querySelectorAll", selector).Get("length").Int()
}

func optionValues(t *testing.T, root js.Value, selector string) []string {
	t.Helper()
	sel := root.Call("querySelector", selector)
	if sel.IsNull() || sel.IsUndefined() {
		t.Fatalf("%s not rendered", selector)
	}
	opts := sel.Get("options")
	out := make([]string, opts.Get("length").Int())
	for i := range out {
		out[i] = opts.Call("item", i).Get("value").String()
	}
	return out
}

// Espacios: the room form's floor select lists the seeded floors. Without
// them no room can be created from the screen.
func TestUILive_RoomForm_FloorSelectListsFloors(t *testing.T) {
	root := mountLiveUI(t, "rl-live-rooms")
	if got := optionValues(t, root, "select[name='floor_id']"); len(got) != 2 {
		t.Fatalf("floor_id options = %v, want the 2 seeded floors", got)
	}
}

// Turnos: the shift form offers every room, every category and, for the
// occupant, "external" (empty value, RL-5) followed by the known occupants.
func TestUILive_ShiftForm_SelectsHaveChoices(t *testing.T) {
	root := mountLiveUI(t, "rl-live-shifts")
	if got := optionValues(t, root, "select[name='room_id']"); len(got) != 5 {
		t.Errorf("room_id options = %v, want the 5 seeded rooms", got)
	}
	if got := optionValues(t, root, "select[name='category_id']"); len(got) != 3 {
		t.Errorf("category_id options = %v, want the 3 categories", got)
	}
	occ := optionValues(t, root, "select[name='occupant_id']")
	if len(occ) != 3 || occ[0] != "" {
		t.Errorf("occupant_id options = %v, want [\"\" occ1 occ2]", occ)
	}
}

// Tablero: one card per active room, and a reload (filter change) replaces
// the cards instead of appending a second copy.
func TestUILive_Board_OneCardPerRoom_AfterReload(t *testing.T) {
	root := mountLiveUI(t, "rl-live-board")
	if n := countAll(root, "[data-board-room]"); n != 5 {
		t.Fatalf("board cards = %d, want 5", n)
	}
	filter := root.Call("querySelector", "select[data-board-filter='category']")
	if filter.IsNull() || filter.IsUndefined() {
		t.Fatal("board category filter not rendered")
	}
	filter.Set("value", "")
	filter.Call("dispatchEvent", js.Global().Get("Event").New("change", map[string]any{"bubbles": true}))
	if n := countAll(root, "[data-board-room]"); n != 5 {
		t.Fatalf("after reload board cards = %d, want 5 (reload must replace, not append)", n)
	}
}

// Áreas y equipos: the selected room shows one checkbox per category and per
// equipment, so the administrator can enable them.
func TestUILive_Assign_ShowsCheckboxes(t *testing.T) {
	root := mountLiveUI(t, "rl-live-assign")
	if n := countAll(root, "input[type='checkbox'][data-category-id]"); n != 3 {
		t.Errorf("category checkboxes = %d, want 3", n)
	}
	if n := countAll(root, "input[type='checkbox'][data-equipment-id]"); n != 3 {
		t.Errorf("equipment checkboxes = %d, want 3", n)
	}
}
