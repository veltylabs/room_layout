package roomlayout

import "webtyp.com/fmt"

// Cell es una celda de la cuadrícula de un piso. Cada piso numera desde 1: la celda de arriba a
// la izquierda de cualquier piso es {Col: 1, Row: 1}.
type Cell struct{ Col, Row int }

const (
	coordSep = ","
	rectSep  = ";"
)

// String escribe la celda como "columna,fila"; es el formato de RoomArtifact.Cell.
//
// Ejemplo: Cell{Col: 7, Row: 1}.String() → "7,1"
func (c Cell) String() string { return fmt.Sprint(c.Col) + coordSep + fmt.Sprint(c.Row) }

// ParseCell lee una celda escrita con Cell.String; ok es false si el texto no es una celda.
func ParseCell(s string) (c Cell, ok bool) {
	n, ok := positives(s, 2)
	if !ok {
		return Cell{}, false
	}
	return Cell{Col: n[0], Row: n[1]}, true
}

// Rect es un rectángulo de celdas: desde {Col, Row} hacia la derecha W columnas y hacia abajo H filas.
type Rect struct{ Col, Row, W, H int }

// Rects divide un conjunto de celdas en rectángulos, fila por fila, cada uno lo más ancho y alto
// posible. Un box rectangular es un solo rectángulo; uno en L, dos.
//
// Ejemplo: Rects([]Cell{{1, 1}, {2, 1}, {1, 2}, {2, 2}}) → []Rect{{1, 1, 2, 2}}
func Rects(cells []Cell) []Rect {
	cols, rows := 0, 0
	for _, c := range cells {
		cols, rows = max(cols, c.Col), max(rows, c.Row)
	}
	in := make([]bool, cols*rows)
	for _, c := range cells {
		in[(c.Row-1)*cols+c.Col-1] = true
	}
	free := func(col, row int) bool { return in[(row-1)*cols+col-1] }
	var out []Rect
	for row := 1; row <= rows; row++ {
		for col := 1; col <= cols; col++ {
			if !free(col, row) {
				continue
			}
			w := 1
			for col+w <= cols && free(col+w, row) {
				w++
			}
			h := 1
			for row+h <= rows && rowFree(free, col, row+h, w) {
				h++
			}
			for y := row; y < row+h; y++ {
				for x := col; x < col+w; x++ {
					in[(y-1)*cols+x-1] = false
				}
			}
			out = append(out, Rect{Col: col, Row: row, W: w, H: h})
		}
	}
	return out
}

// EncodeArea escribe un conjunto de celdas como rectángulos "columna,fila,ancho,alto" separados
// por ";"; es el formato de Floor.HabitableCells y Room.Cells. Un box rectangular es un solo
// rectángulo, sin importar su tamaño.
//
// Ejemplo: EncodeArea([]Cell{{1, 1}, {2, 1}, {1, 2}, {2, 2}}) → "1,1,2,2"
func EncodeArea(cells []Cell) string {
	out := ""
	for _, r := range Rects(cells) {
		if out != "" {
			out += rectSep
		}
		out += fmt.Convert([]string{fmt.Sprint(r.Col), fmt.Sprint(r.Row), fmt.Sprint(r.W), fmt.Sprint(r.H)}).Join(coordSep).String()
	}
	return out
}

// DecodeArea lee el texto de EncodeArea y devuelve sus celdas fila por fila. Texto vacío → sin
// celdas; texto mal formado → ErrInvalidArea.
func DecodeArea(s string) ([]Cell, error) {
	if s == "" {
		return nil, nil
	}
	var cells []Cell
	for _, rect := range fmt.Convert(s).Split(rectSep) {
		n, ok := positives(rect, 4)
		if !ok {
			return nil, ErrInvalidArea
		}
		for row := n[1]; row < n[1]+n[3]; row++ {
			for col := n[0]; col < n[0]+n[2]; col++ {
				cells = append(cells, Cell{Col: col, Row: row})
			}
		}
	}
	return cells, nil
}

func rowFree(free func(col, row int) bool, col, row, w int) bool {
	for x := col; x < col+w; x++ {
		if !free(x, row) {
			return false
		}
	}
	return true
}

// positives lee exactamente n enteros ≥ 1 separados por coma.
func positives(s string, n int) ([]int, bool) {
	parts := fmt.Convert(s).Split(coordSep)
	if len(parts) != n {
		return nil, false
	}
	out := make([]int, n)
	for i, p := range parts {
		v, err := fmt.Convert(p).Int()
		if err != nil || v < 1 {
			return nil, false
		}
		out[i] = v
	}
	return out, true
}
