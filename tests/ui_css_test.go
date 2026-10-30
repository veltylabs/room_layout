//go:build !wasm

package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/ui"
)

// La hoja del mapa es válida: Stylesheet() entra en pánico ante una composición inválida.
func TestUICSS(t *testing.T) {
	if (&ui.RootView{}).RenderCSS() == nil {
		t.Fatal("RenderCSS returned nil")
	}
}
