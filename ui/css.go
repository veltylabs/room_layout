//go:build !wasm

package ui

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

const (
	PartStatusFree   = widget.Part("status-free")
	PartStatusBusy   = widget.Part("status-busy")
	PartStatusClosed = widget.Part("status-closed")
)

// RootCSS returns the package-level stylesheet for room_layout ui.
func RootCSS() *css.Stylesheet {
	return css.NewStylesheet()
}
