package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
)

// gridPos ubica un elemento en la cuadrícula de un piso (fila y columna desde 0). Con números
// («Planta» y «Espacios») la fila 0 y la columna 0 son los números, y las celdas van desde 1.
func gridPos(row, col int) string {
	return "grid-row: " + fmt.Sprint(row+1) + "; grid-column: " + fmt.Sprint(col+1) + "; "
}

// roomOutline es el contorno de un espacio, con un borde de outlineWidth px.
const roomOutline = "z-index: 1; pointer-events: none; display: flex; align-items: center; justify-content: center; text-align: center; padding: 2px; overflow: hidden; border: 2px solid var(--color-primary); font-size: 11px; font-weight: 700; line-height: 1.1; color: var(--color-on-surface);"

const outlineWidth = 2

// gridLayout es cómo se dibuja la cuadrícula de un piso: numbered («Planta» y «Espacios») lleva la
// fila y la columna de números y 2 px entre celdas; plain («Equipamiento» y «Operación») ni números
// ni separación.
type gridLayout struct{ offset, gap int }

var (
	numbered = gridLayout{offset: 1, gap: 2}
	plain    = gridLayout{offset: 0, gap: 0}
)

// wall hace salir el contorno hacia afuera lo justo para que dos espacios que se tocan monten sus
// bordes y compartan una sola pared, en vez de dos paredes juntas.
func (g gridLayout) wall() string {
	return " margin: -" + fmt.Sprint((outlineWidth+g.gap)/2) + "px;"
}

// Dónde va el nombre dentro del contorno de un espacio.
const (
	nameNone   = ""
	nameCenter = " "
	nameTop    = " align-items: flex-start;" // arriba: deja ver los equipos del espacio
)

// roomOutlines dibuja sobre la planta el contorno de cada espacio guardado de un piso, con su
// nombre donde diga name (nameNone: sin nombre). No recibe toques: pasan a la celda de abajo.
// except es un espacio que no se dibuja (el que se está editando).
func (r *RootView) roomOutlines(floorID, except string, g gridLayout, name string) []*dom.Element {
	var out []*dom.Element
	for _, rm := range r.Rooms {
		if rm.FloorID != floorID || rm.ID == except {
			continue
		}
		cells := make([]roomlayout.Cell, 0, len(rm.Cells))
		for _, c := range rm.Cells {
			if cell, ok := roomlayout.ParseCell(c); ok {
				cells = append(cells, cell)
			}
		}
		rects := roomlayout.Rects(cells)
		biggest := 0
		for i, rc := range rects {
			if rc.W*rc.H > rects[biggest].W*rects[biggest].H {
				biggest = i
			}
		}
		for i, rc := range rects {
			// la clave incluye el nombre y el lugar: si cambian, el recuadro se rehace entero
			el := html.Div().Attr("aria-hidden", "true").Key(fmt.Sprintf("outline-%s-%d-%s-%d,%d,%d,%d-%d", rm.ID, i, rm.Name, rc.Col, rc.Row, rc.W, rc.H, g.offset))
			if name != nameNone && i == biggest {
				el.Text(rm.Name)
			}
			place := fmt.Sprintf("grid-row: %d / span %d; grid-column: %d / span %d; ", rc.Row+g.offset, rc.H, rc.Col+g.offset, rc.W)
			el.Attr("style", place+roomOutline+g.wall()+name)
			out = append(out, el)
		}
	}
	return out
}

// Sin números ni líneas («Equipamiento» y «Operación»): las celdas se tocan y solo se ve el plano.
const (
	planGrid    = "display: grid; grid-template-columns: repeat(%d, 28px); grid-auto-rows: 28px; gap: 0; padding: 10px; background: var(--color-background); border: 1px solid var(--color-outline); border-radius: 12px; overflow-x: auto;"
	planOff     = "pointer-events: none;"
	planFree    = "background: var(--color-surface);"
	planBox     = "background: var(--color-selection);"
	planShared  = "background: var(--color-surface-sunken);"
	planStore   = "background: var(--color-accent-wash);"
	planContent = " display: flex; align-items: center; justify-content: center;"
)

// planCell es el fondo de una celda sin líneas, según lo que haya en ella.
func planCell(fl *FloorData, room *RoomData, coord string) string {
	if !CellInList(fl.HabitableCells, coord) {
		return planOff
	}
	if room == nil {
		return planFree
	}
	switch room.RoomType {
	case RoomTypeShared:
		return planShared
	case RoomTypeStore:
		return planStore
	}
	return planBox
}

// roomPlace ubica en una cuadrícula sin números el rectángulo más grande de un espacio.
func roomPlace(cells []string) (string, bool) {
	area := make([]roomlayout.Cell, 0, len(cells))
	for _, c := range cells {
		if cell, ok := roomlayout.ParseCell(c); ok {
			area = append(area, cell)
		}
	}
	rects := roomlayout.Rects(area)
	if len(rects) == 0 {
		return "", false
	}
	b := rects[0]
	for _, rc := range rects {
		if rc.W*rc.H > b.W*b.H {
			b = rc
		}
	}
	return fmt.Sprintf("grid-row: %d / span %d; grid-column: %d / span %d; ", b.Row, b.H, b.Col, b.W), true
}

// roomButton es el espacio que se toca en «Operación»: transparente sobre su bloque, con el nombre al centro.
const roomButton = "z-index: 2; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px; padding: 2px; overflow: hidden; background: transparent; border: none; cursor: pointer; font-size: 11px; font-weight: 700; line-height: 1.1; color: var(--color-on-surface);"
