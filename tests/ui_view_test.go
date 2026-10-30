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
	// El mapa (pantalla propia) y las tres listas: boxes, pisos y equipamiento.
	if len(grp.Children) != 4 {
		t.Fatalf("len(grp.Children) = %d, want 4", len(grp.Children))
	}
	if grp.Children[0].Component == nil || grp.Children[0].Help == "" {
		t.Errorf("first child must be the map: a Component with Help, got %+v", grp.Children[0].ID)
	}
	for _, child := range grp.Children {
		if err := child.Validate(); err != nil {
			t.Errorf("child %s Validate() error: %v", child.ID, err)
		}
		if child.Presenter != nil && child.Presenter.Description() == "" {
			t.Errorf("child %s Description() is empty", child.ID)
		}
	}
}
