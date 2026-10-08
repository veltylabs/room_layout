//go:build !wasm

package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/ui"
)

func TestUICSS(t *testing.T) {
	sheet := ui.RootCSS()
	if sheet == nil {
		t.Fatalf("RootCSS returned nil")
	}
}
