package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/model"
)

// Each error of the module carries its kind, so rpc.WriteError answers 404/409/400
// with its text instead of a 500 that hides it.
func TestErrorKinds(t *testing.T) {
	cases := []struct {
		err  error
		kind model.ErrorKind
	}{
		{roomlayout.ErrNotFound, model.ErrorNotFound},
		{roomlayout.ErrCodeAlreadyExists, model.ErrorConflict},
		{roomlayout.ErrFloorInUse, model.ErrorConflict},
		{roomlayout.ErrCategoryInUse, model.ErrorConflict},
		{roomlayout.ErrRoomOverlap, model.ErrorConflict},
		{roomlayout.ErrOccupantOverlap, model.ErrorConflict},
		{roomlayout.ErrTenantRequired, model.ErrorInvalid},
		{roomlayout.ErrInvalidRange, model.ErrorInvalid},
		{roomlayout.ErrInvalidDate, model.ErrorInvalid},
		{roomlayout.ErrInvalidWeekday, model.ErrorInvalid},
		{roomlayout.ErrUnknownCategory, model.ErrorInvalid},
		{roomlayout.ErrUnknownOccupant, model.ErrorInvalid},
		{roomlayout.ErrOccupantRequired, model.ErrorInvalid},
		{roomlayout.ErrCategoryNotAllowed, model.ErrorInvalid},
		{roomlayout.ErrOutsideBounds, model.ErrorInvalid},
		{roomlayout.ErrNotWeekly, model.ErrorInvalid},
		{roomlayout.ErrCannotRead, model.ErrorInvalid},
	}
	for _, c := range cases {
		if got := model.KindOf(c.err); got != c.kind {
			t.Errorf("%q: kind %v, want %v", c.err.Error(), got, c.kind)
		}
	}
}
