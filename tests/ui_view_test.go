//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

type mockCaller struct{}

func (c *mockCaller) Call(route string, args model.Encodable, res model.Decodable, cb func(error)) {
	if cb != nil {
		cb(nil)
	}
}

func (c *mockCaller) Dispatch(route string, args model.Encodable) {}

var _ router.Caller = (*mockCaller)(nil)

func TestUI_Screens(t *testing.T) {
	fake := view.NewRecords().Caller(&mockCaller{})
	screens := ui.Screens(fake)
	if len(screens) != 1 {
		t.Fatalf("len(screens) = %d, want 1", len(screens))
	}
	grp := screens[0]
	if err := grp.Validate(); err != nil {
		t.Fatalf("grp.Validate() error: %v", err)
	}
	if grp.ID != ui.ID {
		t.Errorf("grp.ID = %q, want %q", grp.ID, ui.ID)
	}
	if len(grp.Children) != 3 {
		t.Fatalf("len(grp.Children) = %d, want 3", len(grp.Children))
	}
	for _, child := range grp.Children {
		if err := child.Validate(); err != nil {
			t.Errorf("child %s Validate() error: %v", child.ID, err)
		}
		if child.Presenter.Description() == "" {
			t.Errorf("child %s Description() is empty", child.ID)
		}
	}
}
