//go:build !wasm

package ui

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

const (
	PartHeader       = widget.Part("header")
	PartContent      = widget.Part("content")
	PartSlider       = widget.Part("slider")
	PartFloor        = widget.Part("floor")
	PartAside        = widget.Part("aside")
	PartPanel        = widget.Part("panel")
	PartFooter       = widget.Part("footer")
	PartStatusFree   = widget.Part("status-free")
	PartStatusBusy   = widget.Part("status-busy")
	PartStatusClosed = widget.Part("status-closed")
)

func (r *RootView) RenderCSS() *css.Stylesheet {
	return style.For(r).
		Root(
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(PartHeader,
			style.Row(style.Space4),
			style.Pad(style.Space3),
			style.DividerBelow(),
		).
		Part(PartContent,
			style.Fill(),
			style.Pad(style.Space4),
		).
		Part(PartFooter,
			style.Row(style.Space4),
			style.Pad(style.Space3),
		).
		Stylesheet()
}

func (s *Stage1Planta) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Fill(),
			style.Stack(style.Space4),
		).
		Part(PartSlider,
			style.ScrollRow(style.Space4),
			style.Pad(style.Space4),
		).
		Part(PartFloor,
			style.As(style.Panel),
			style.Pad(style.Space4),
			style.Round(style.RadiusLg),
		).
		Stylesheet()
}

func (s *Stage2Espacios) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Fill(),
			style.Stack(style.Space4),
		).
		Part(PartAside,
			style.Stack(style.Space4),
			style.Pad(style.Space3),
		).
		Part(PartPanel,
			style.As(style.Panel),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
		).
		Stylesheet()
}

func (s *Stage3Artefactos) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Fill(),
			style.Stack(style.Space4),
		).
		Part(PartAside,
			style.Stack(style.Space4),
			style.Pad(style.Space3),
		).
		Part(PartPanel,
			style.As(style.Panel),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
		).
		Stylesheet()
}

func (s *Stage4Operacion) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Fill(),
			style.Stack(style.Space4),
		).
		Part(PartAside,
			style.Stack(style.Space4),
			style.Pad(style.Space3),
		).
		Part(PartPanel,
			style.As(style.Panel),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
		).
		Stylesheet()
}

func RootCSS() *css.Stylesheet {
	return css.NewStylesheet()
}
