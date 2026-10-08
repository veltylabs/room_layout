package ui

import (
	"webtyp.com/fmt"
	"webtyp.com/widget"
)

const (
	NameStagePlanta     = widget.Name("stage-planta")
	NameStageEspacios   = widget.Name("stage-espacios")
	NameStageArtefactos = widget.Name("stage-artefactos")
	NameStageOperacion  = widget.Name("stage-operacion")
)

func (s *Stage1Planta) WidgetName() widget.Name { return NameStagePlanta }
func (s *Stage1Planta) WidgetKind() widget.Kind { return widget.Tabs }

func (s *Stage2Espacios) WidgetName() widget.Name { return NameStageEspacios }
func (s *Stage2Espacios) WidgetKind() widget.Kind { return widget.Tabs }

func (s *Stage3Artefactos) WidgetName() widget.Name { return NameStageArtefactos }
func (s *Stage3Artefactos) WidgetKind() widget.Kind { return widget.Tabs }

func (s *Stage4Operacion) WidgetName() widget.Name { return NameStageOperacion }
func (s *Stage4Operacion) WidgetKind() widget.Kind { return widget.Tabs }

// Constantes de tipos de espacio.
const (
	RoomTypeBox    = "box"
	RoomTypeShared = "shared"
	RoomTypeStore  = "store"
)

// Constantes de tipos de artefacto.
const (
	ArtifactAgenda  = "agenda"
	ArtifactPC      = "pc"
	ArtifactPrinter = "printer"
	ArtifactGastro  = "gastro"
	ArtifactChair   = "chair"
	ArtifactECG     = "ecg"
	ArtifactMonitor = "monitor"
	ArtifactSupply  = "supply"
)

// Constantes de estado de artefacto.
const (
	StatusOk   = "ok"
	StatusWarn = "warn"
	StatusCrit = "crit"
)

// FloorData modela la planta en memoria (slices estrictamente tipados, sin mapas).
type FloorData struct {
	ID             string
	Name           string
	Position       int
	HabitableCells []string // Coordenadas "A1", "B2", etc.
}

// RoomData modela un espacio dentro de una planta.
type RoomData struct {
	ID       string
	FloorID  string
	Code     string
	Name     string
	RoomType string // "box", "shared", "store"
	Cells    []string
}

// ArtifactData modela un artefacto o equipamiento posicionado en una celda.
type ArtifactData struct {
	ID              string
	RoomID          string
	FloorID         string
	Kind            string // "agenda", "pc", "printer", "gastro", "chair", "ecg", "monitor", "supply"
	Code            string // e.g. "PC-01"
	Cell            string // e.g. "B2"
	Status          string // "ok", "warn", "crit"
	Reason          string
	NextMaintenance string
	Items           string
}

// ShiftData modela la asignación de turno en un box.
type ShiftData struct {
	ID            string
	RoomID        string
	OccupantID    string
	OccupantLabel string
	DayOfWeek     int // 1=Lunes .. 6=Sábado
	StartHour     string
	EndHour       string
}

// DoctorData modela un profesional para la vista de operación.
type DoctorData struct {
	ID              string
	Name            string
	Specialty       string
	Initials        string
	WeeklyOccupancy int // 0..100
}

// ColName convierte un índice de columna 0-based a letra ("A", "B", ...).
func ColName(c int) string {
	res := ""
	for c >= 0 {
		rem := c % 26
		res = string(rune('A'+rem)) + res
		c = c/26 - 1
	}
	return res
}

// FormatCell compone fila (0-based) y col (0-based) en coordenada como "A1", "B2".
func FormatCell(r, c int) string {
	return ColName(c) + fmt.Sprint(r+1)
}

// ParseCell descompone "B2" en fila=1, col=1.
func ParseCell(cell string) (int, int, bool) {
	if len(cell) < 2 {
		return 0, 0, false
	}
	col := 0
	idx := 0
	for idx < len(cell) && cell[idx] >= 'A' && cell[idx] <= 'Z' {
		col = col*26 + int(cell[idx]-'A') + 1
		idx++
	}
	if idx == 0 || idx >= len(cell) {
		return 0, 0, false
	}
	col-- // 0-based

	row := 0
	for idx < len(cell) {
		if cell[idx] < '0' || cell[idx] > '9' {
			return 0, 0, false
		}
		row = row*10 + int(cell[idx]-'0')
		idx++
	}
	if row < 1 {
		return 0, 0, false
	}
	return row - 1, col, true
}

// ParsePositiveInt parses a positive integer from string without stdlib strconv.
func ParsePositiveInt(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= '0' && b <= '9' {
			n = n*10 + int(b-'0')
		}
	}
	return n
}

// CellInList verifica existencia por búsqueda lineal (sin usar mapa).
func CellInList(cells []string, target string) bool {
	for _, c := range cells {
		if c == target {
			return true
		}
	}
	return false
}

// AddCell añade la celda si no está presente.
func AddCell(cells []string, cell string) []string {
	if CellInList(cells, cell) {
		return cells
	}
	return append(cells, cell)
}

// RemoveCell remueve la celda si está presente.
func RemoveCell(cells []string, cell string) []string {
	out := make([]string, 0, len(cells))
	for _, c := range cells {
		if c != cell {
			out = append(out, c)
		}
	}
	return out
}

// CellsOverlap comprueba si hay intersección entre dos conjuntos de celdas.
func CellsOverlap(a, b []string) bool {
	for _, ca := range a {
		if CellInList(b, ca) {
			return true
		}
	}
	return false
}

// BoundingBox calcula el rectángulo que encierra las celdas dadas.
func BoundingBox(cells []string) (minCol, minRow, maxCol, maxRow int) {
	if len(cells) == 0 {
		return 0, 0, 0, 0
	}
	minCol, minRow = 9999, 9999
	maxCol, maxRow = -1, -1
	for _, c := range cells {
		r, col, ok := ParseCell(c)
		if !ok {
			continue
		}
		if col < minCol {
			minCol = col
		}
		if col > maxCol {
			maxCol = col
		}
		if r < minRow {
			minRow = r
		}
		if r > maxRow {
			maxRow = r
		}
	}
	return
}

// BoundingBoxLabel devuelve texto descriptivo como "B2:D4 · 6 celdas".
func BoundingBoxLabel(cells []string) string {
	if len(cells) == 0 {
		return "0 celdas"
	}
	minCol, minRow, maxCol, maxRow := BoundingBox(cells)
	start := FormatCell(minRow, minCol)
	end := FormatCell(maxRow, maxCol)
	if start == end {
		return start + " · 1 celda"
	}
	return start + ":" + end + " · " + fmt.Sprint(len(cells)) + " celdas"
}

// FindRoomByCell busca el espacio que contiene una celda en una planta determinada.
func FindRoomByCell(rooms []RoomData, floorID, cell string) *RoomData {
	for i := range rooms {
		if rooms[i].FloorID == floorID && CellInList(rooms[i].Cells, cell) {
			return &rooms[i]
		}
	}
	return nil
}

// FindArtifactByCell busca un artefacto posicionado en una celda.
func FindArtifactByCell(artifacts []ArtifactData, floorID, cell string) *ArtifactData {
	for i := range artifacts {
		if artifacts[i].FloorID == floorID && artifacts[i].Cell == cell {
			return &artifacts[i]
		}
	}
	return nil
}
