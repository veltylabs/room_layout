package migrate_test

import (
	"testing"

	"github.com/veltylabs/room_layout/migrate"
	"webtyp.com/storage/mem"
)

func TestMigrate(t *testing.T) {
	conn := mem.New()
	if err := migrate.Migrate(conn, nil); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}
