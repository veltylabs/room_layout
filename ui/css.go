//go:build !wasm

package ui

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS es la hoja de estilos de las cuatro etapas del mapa (ver docs/ui-reference).
func (r *RootView) RenderCSS() *css.Stylesheet {
	return style.For(r).
		Root(style.Stack(style.Space4), style.Pad(style.Space4)).
		Part(PartStage, style.Stack(style.Space4)).
		Part(PartHeader, style.Row(style.Space4), style.DividerBelow(), style.Pad(style.Space2)).
		Part(PartActions, style.Row(style.Space2), style.PushEnd()).
		Part(PartFooter, style.Row(style.Space3)).
		Part(PartSummary, style.Grow(), style.FontSize(style.TextSm)).
		Part(PartTitleBar, style.Row(style.Space4)).
		Part(PartTitle, style.Stack(style.Space1), style.Grow()).
		Part(PartStepTag, style.FontSize(style.TextXs), style.FontWeight(style.WeightBold)).
		Part(PartHeading, style.Heading()).
		Part(PartButton, style.Button(style.Panel)).
		Part(PartPrimary, style.Button(style.Primary)).
		Part(PartDanger, style.Button(style.Subtle)).
		Part(PartDelete, style.Button(style.Danger)).
		Part(PartNav, style.Row(style.Space2), style.CenterContent()).
		Part(PartFloor, style.Stack(style.Space3), style.As(style.Panel), style.Pad(style.Space4), style.Round(style.RadiusLg)).
		Part(PartFloorHeader, style.Row(style.Space3)).
		Part(PartFloorBadge, style.IconBox(style.IconMd), style.CenterContent(), style.Round(style.RadiusMd), style.As(style.Primary)).
		Part(PartFloorName, style.Grow(), style.FontWeight(style.WeightBold)).
		Part(PartFloorCount, style.FontSize(style.TextSm)).
		Part(PartFloorAdd, style.Button(style.Subtle)).
		Part(PartPanel, style.Stack(style.Space3), style.As(style.Panel), style.Pad(style.Space4), style.Round(style.RadiusLg)).
		Part(PartPanelTitle, style.Heading()).
		Part(PartField, style.Stack(style.Space1), style.FontSize(style.TextSm)).
		Part(PartInput, style.As(style.Field), style.ControlBox(), style.PadInline(style.Space2), style.Round(style.RadiusMd)).
		Part(PartSeg, style.Row(style.Space1)).
		Part(PartSegItem, style.Button(style.Subtle)).
		When(widget.Selected, PartSegItem, style.As(style.AccentInverse)).
		Part(PartColumns, style.Sidebar(style.SideEnd, style.RailWide, style.Space4)).
		Part(PartContent, style.Stack(style.Space3)).
		Part(PartAside, style.Stack(style.Space4)).
		Part(PartSlider, style.Stack(style.Space4)).
		Part(PartLegend, style.Row(style.Space3), style.FontSize(style.TextSm)).
		Part(PartLegendItem, style.Row(style.Space1)).
		Part(PartSwatchBox, style.IconBox(style.IconSm), style.Round(style.RadiusSm), style.As(style.Highlight)).
		Part(PartSwatchShared, style.IconBox(style.IconSm), style.Round(style.RadiusSm), style.As(style.Inset)).
		Part(PartSwatchStore, style.IconBox(style.IconSm), style.Round(style.RadiusSm), style.As(style.AccentWash)).
		Part(PartSwatchFree, style.IconBox(style.IconSm), style.Round(style.RadiusSm), style.As(style.Field)).
		Part(PartList, style.Stack(style.SpaceNone), style.DividerBetween()).
		Part(PartItem, style.Row(style.Space2), style.Pad(style.Space2), style.Round(style.RadiusMd), style.As(style.Panel), style.Interactive(style.Panel)).
		When(widget.Selected, PartItem, style.As(style.AccentWash)).
		Part(PartItemInfo, style.Stack(style.SpaceNone), style.Grow()).
		Part(PartItemMeta, style.FontSize(style.TextXs)).
		Part(PartItemName, style.Grow(), style.FontWeight(style.WeightMedium)).
		Part(PartGridField, style.Row(style.Space1), style.CenterContent()).
		Stylesheet()
}
