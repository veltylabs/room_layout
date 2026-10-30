package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
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
	HabitableCells []string // Celdas "columna,fila" de FormatCell
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

// FormatCell compone fila y columna (0-based, como los índices de la cuadrícula) en la celda
// guardada: FormatCell(0, 6) → "7,1". Cada piso numera desde 1.
func FormatCell(r, c int) string {
	return roomlayout.Cell{Col: c + 1, Row: r + 1}.String()
}

// ParseCell descompone "7,1" en fila=0, col=6 (0-based).
func ParseCell(cell string) (int, int, bool) {
	c, ok := roomlayout.ParseCell(cell)
	return c.Row - 1, c.Col - 1, ok
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
