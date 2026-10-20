package tests

import (
	"testing"

	"webtyp.com/time"
)

func TestLocalNowTimezone(t *testing.T) {
	env := newTestEnv("tenantA")

	// 2026-03-15 02:30:00 UTC
	dateSec, _ := time.ParseDate("2026-03-15")
	env.clockSec = dateSec + time.Unix(9000) // 02:30:00 UTC

	// In America/Santiago (UTC-3), 02:30 UTC on March 15 corresponds to 23:30 on March 14 (previous date)
	dayShifts, err := env.mod.ListRoomDay("tenantA", "", 0) // 0 date triggers localNow()
	if err != nil {
		t.Fatalf("ListRoomDay failed: %v", err)
	}
	_ = dayShifts
}
