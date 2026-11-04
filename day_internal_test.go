package roomlayout

import (
	"testing"
	"webtyp.com/time"
)

func TestLocalAt(t *testing.T) {
	zone, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatalf("Failed to load zone: %v", err)
	}

	mod := &Module{
		deps: Deps{
			Zone: zone,
		},
	}

	// 2019-07-16 01:30 UTC
	// 2019-07-16 is roughly 1563240000 + 1:30 = 1563240000 + 5400 = 1563245400
	// Let's use DateUTC and add minutes. Or just use a hardcoded value if we know it.
	// Actually, time.ParseDateUTC gives UTC midnight.
	dateSec, _ := time.ParseDateUTC("2019-07-16")
	sec := dateSec + 90*60 // 01:30 UTC

	d, min := mod.localAt(sec)

	expectedD, _ := time.ParseDateUTC("2019-07-15")
	expectedMin := 1290

	if d != expectedD {
		t.Errorf("Expected date %v, got %v", expectedD, d)
	}
	if min != expectedMin {
		t.Errorf("Expected minute %v, got %v", expectedMin, min)
	}
}
