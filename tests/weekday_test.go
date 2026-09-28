package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestWeekdayOptions(t *testing.T) {
	opts := roomlayout.WeekdayOptions()
	if len(opts) != 8 {
		t.Fatalf("len(WeekdayOptions) = %d, want 8", len(opts))
	}
	if opts[0].Key != "" {
		t.Errorf("opts[0].Key = %q, want empty string", opts[0].Key)
	}

	expectedKeys := []string{
		roomlayout.WeekdayMonday,
		roomlayout.WeekdayTuesday,
		roomlayout.WeekdayWednesday,
		roomlayout.WeekdayThursday,
		roomlayout.WeekdayFriday,
		roomlayout.WeekdaySaturday,
		roomlayout.WeekdaySunday,
	}

	for i, key := range expectedKeys {
		if got := opts[i+1].Key; got != key {
			t.Errorf("opts[%d].Key = %q, want %q", i+1, got, key)
		}
	}
}
