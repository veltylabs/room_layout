// Root test: mapErrorStatus and domainError are unexported; tests/ can only reach the public API.
package roomlayout

import (
	"testing"

	"webtyp.com/fmt"
)

// Every sentinel keeps its HTTP status and the exact text fmt.Err gave it.
func TestMapErrorStatusAndSentinelTexts(t *testing.T) {
	cases := []struct {
		err  domainError
		code int
	}{
		{ErrNotFound, 404},
		{ErrCodeAlreadyExists, 409}, {ErrFloorInUse, 409}, {ErrCategoryInUse, 409},
		{ErrRoomOverlap, 409}, {ErrOccupantOverlap, 409},
		{ErrTenantRequired, 400}, {ErrInvalidRange, 400}, {ErrInvalidDate, 400},
		{ErrInvalidWeekday, 400}, {ErrUnknownCategory, 400}, {ErrUnknownOccupant, 400},
		{ErrOccupantRequired, 400}, {ErrCategoryNotAllowed, 400}, {ErrOutsideBounds, 400},
		{ErrNotWeekly, 400},
	}
	for _, c := range cases {
		if got := mapErrorStatus(c.err); got != c.code {
			t.Errorf("mapErrorStatus(%q) = %d, want %d", string(c.err), got, c.code)
		}
		if want := fmt.Err(string(c.err)).Error(); c.err.Error() != want {
			t.Errorf("text changed: %q, fmt.Err gave %q", c.err.Error(), want)
		}
	}
	if got := mapErrorStatus(nil); got != 200 {
		t.Errorf("nil → %d, want 200", got)
	}
	if got := mapErrorStatus(&ValidationError{}); got != 400 {
		t.Errorf("*ValidationError → %d, want 400", got)
	}
	if got := mapErrorStatus(fmt.Err("other")); got != 500 {
		t.Errorf("unknown error → %d, want 500", got)
	}
}
