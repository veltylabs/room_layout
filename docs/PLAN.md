---
PLAN: "feat: room_layout — espacios, equipamiento, turnos de ocupación y disponibilidad"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `room_layout` v0.1.0

Módulo nuevo y vacío (solo el esqueleto de `gonew`). Este plan lo lleva a su primera versión
publicable: dominio, ops, vistas, demo en navegador, tests y documentación.

Lee **`AGENTS.md`** en la raíz de este repo antes de escribir código: sus reglas son obligatorias.
Las críticas se repiten abajo, en §1.

---

## 0. Contexto — qué problema resuelve

Un establecimiento (el primer consumidor es un consultorio médico, pero el módulo sirve igual para
un spa o un hotel) tiene **espacios de atención** repartidos en **niveles** (pisos). Cada espacio:

- está habilitado solo para ciertas **categorías** de atención (en el consultorio: especialidades);
- tiene cierto **equipamiento** (impresora, electrocardiógrafo, camilla…);
- lo usan distintos **ocupantes** en distintas franjas horarias (**turnos**), y el personal rota:
  si un ocupante falta o cambia, se reasigna el turno, no decenas de citas.

Hoy nadie puede ver de un vistazo qué espacio está libre, quién lo ocupa y hasta qué hora, ni
buscar "un espacio libre con categoría X el martes de 15:00 a 17:00". Eso es lo que entrega este
módulo, más una consulta **pública** (sin sesión) de disponibilidad para un sitio web.

**Vocabulario del dominio — obligatorio en código, nombres y mensajes:**

| Concepto del negocio (consultorio) | Nombre en el dominio | Tabla |
|---|---|---|
| Piso | nivel | `floor` |
| Box | espacio | `room` |
| Infraestructura | equipamiento | `equipment`, `room_equipment` |
| Área / especialidad | categoría (id blando de otro módulo) | `room_category` |
| Profesional / doctor | ocupante (id blando de otro módulo, o externo) | — |
| Turno de box | turno | `room_shift` |
| "Hoy este turno no aplica" | cancelación de una ocurrencia | `room_shift_cancellation` |

Nunca escribas "box", "doctor", "especialidad" ni "paciente" en el código de este repo. Esas
palabras viven solo en la app que lo compone.

---

## 1. Reglas no negociables (repetidas de `AGENTS.md`)

### 1.1 Imports

- El paquete raíz (`package roomlayout`) importa **solo**: `webtyp.com/model`, `webtyp.com/router`,
  `webtyp.com/view`, `webtyp.com/events`, `webtyp.com/orm`, `webtyp.com/storage`,
  `webtyp.com/input`, `webtyp.com/fmt`, `webtyp.com/time`.
- `migrate/` importa además `webtyp.com/ddl`. `ui/` importa además `webtyp.com/layout/*`,
  `webtyp.com/components/*`, `webtyp.com/dom`, `webtyp.com/html`, `webtyp.com/css`,
  `webtyp.com/svg`, `webtyp.com/widget`, `webtyp.com/form`. `web/` (demo) importa además
  `webtyp.com/storage/mem`, `webtyp.com/router/loopback`, `webtyp.com/events/mock`,
  `webtyp.com/unixid`.
- **Prohibido en cualquier archivo, tests incluidos**: `webtyp.com/sqlite`, `sqlt`, `postgres`,
  `indexdb`, `database/sql`, `webtyp.com/mcp`, `webtyp.com/server`, `net/http`,
  `webtyp.com/json`, `webtyp.com/jsvalue`, `encoding/json`, y **cualquier otro módulo
  `github.com/veltylabs/*`**.

| Prohibido (stdlib) | Usar en su lugar |
|---|---|
| `errors`, `strings`, `strconv`, `fmt` | `webtyp.com/fmt` (`fmt.Err(...)`, `fmt.Errf(...)`, conversión y cadenas) |
| `time` | `webtyp.com/time` |
| `encoding/json` | los modelos generados por `ormc` (`model.Encodable`/`Decodable`); el códec lo elige la app |
| `database/sql` | `webtyp.com/orm` |
| `net/http` | `webtyp.com/router` |
| `map[K]V` | slice de structs con búsqueda lineal, o `[]fmt.KeyValue` |
| `reflect` | nada; `ormc` genera lo necesario |

**Cuál build decide:** el módulo se compila para el navegador (WASM/TinyGo). `gotest` corre ambos
carriles (nativo + navegador) en un solo comando; **los dos deben pasar**. Que
`GOOS=js GOARCH=wasm go build` compile **no** prueba nada: usa la stdlib completa y no es TinyGo.

### 1.2 Forma de un módulo — copia exacta de `patient_directory`

Referencia a imitar archivo por archivo:
<https://github.com/veltylabs/patient_directory> (`model.go`, `module.go`, `ops.go`, `view.go`,
`migrate/migrate.go`, `seed/seed.go`, `ui/*.go`, `web/client.go`, `tests/*`).

- Registro de ops: `func (m *Module) MountOperations(reg router.OperationRegistry)` con
  `reg.Operation(Op, handler).Requires(resource, action).Accepts(&Args{})`. `var _
  router.OperationModule = (*Module)(nil)`. (La plantilla `AGENTS.md` todavía dice
  `MountOps`/`Op`: está desactualizada; manda `patient_directory`.)
- El esquema **no** se crea en `New`. Se crea en `migrate.Migrate(conn ddl.Execer, c ddl.Compiler)`
  con `ddl.New(conn, c).Sync(&X{}, ...)`.
- `model_orm.go` lo genera `ormc` (instalar con `go install webtyp.com/ormc/cmd/ormc@latest`,
  correr `ormc` en la raíz). **Nunca se edita a mano.**
- Handler: `ctx.Decode(&args)` (error → 400) → `args.Validate(action)` (error → 400) → método de
  servicio → `ctx.Encode` / estado. Estados: 400 validación · 404 no encontrado · 409 conflicto ·
  500 solo error interno real. Un error real de BD **nunca** se convierte en 404.
- Todo `UPDATE`/`DELETE` lleva la condición de `tenant_id`. Todas las lecturas filtran por
  `tenant_id`.
- Todo `string` con valores cerrados (estados, etc.) es una **constante exportada**; nunca un
  literal en la lógica.
- Tests en `tests/` (paquete `tests`, sin `go.mod` propio), sobre `orm.New(mem.New())`, ops sobre
  `webtyp.com/router/mock`. Correr con `gotest`, nunca `go test`.
- Sin `TODO`, sin código comentado, sin stubs.

---

## 2. Design gate

**1. Prior art.**
- *Google Workspace — calendar resources*: edificios → pisos → salas, cada una con *features* y
  capacidad; se reserva **por evento**.
- *Microsoft Places / Exchange room mailboxes*: edificios, pisos y listas de salas; también reserva
  **por reunión**.
- *Odoo Planning*: **turnos** (*shifts*) asignados a recursos y roles, recurrentes, pensados para
  personal que rota.

Este módulo toma la jerarquía nivel → espacio → equipamiento de los dos primeros y el **turno
recurrente** de Odoo. No reserva por cita porque el problema real es la **rotación por turno**: si
un ocupante falta, se mueve un turno, no N citas. Por la misma razón `appointment_booking` sigue
sin saber nada de espacios.

**2. Prueba del nombre para novatos.** `room`, `floor`, `equipment`, `shift`, `occupant`,
`category`, `CancelShiftOccurrence`, `MoveShiftOccurrence`, `FindFreeRooms`,
`ListPublicAvailability`: todos se leen como frase sin documentación. Descartado `amenity`
(palabra de hotel) y `feature` (en software significa otra cosa).

**3. Balance de complejidad.**
```
Conceptos que aprende el desarrollador   +6 (floor, room, equipment, category ref, occupant ref, shift)
Archivos que toca para "mostrar espacios" +1 línea en modules/browser.go y +1 en modules/server.go de la app
Líneas en el punto de llamada             +~15 (New + dos adaptadores de lectura en la app)
Formas de hacer lo mismo                  0 (no existe ningún concepto de espacio en el ecosistema)
```

**4. Dónde va.** Repo propio `veltylabs/room_layout`: la ocupación de espacios es otra
responsabilidad que la agenda de citas (`appointment_booking`) y que el personal
(`staff_manager`). Categorías y ocupantes se leen mediante dos interfaces estrechas **declaradas
aquí** (`CategoryReader`, `OccupantReader`); la app las conecta. No importa ningún otro módulo.

**5. Qué borra.** El esqueleto de `gonew`: `room_layout.go` (`type RoomLayout struct{}` y
`func New() *RoomLayout`). Lo demás es capacidad nueva.

---

## 3. Modelo de datos — `model.go`

Todos los `Definition` exactamente así. Widgets `input.*` solo en campos que un usuario edita en
un formulario; `model.*` en ids, `tenant_id`, marcas de tiempo y modelos de solo salida.

```go
package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

// ModelName es la identidad del módulo: prefijo de cada op en el cable.
const ModelName = "room_layout"

// Recursos RBAC.
const (
	ResourceFloor     = "floor"
	ResourceRoom      = "room"
	ResourceEquipment = "equipment"
	ResourceShift     = "room_shift"
)

var FloorModel = model.Definition{
	Name: "floor",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 60}},
		{Name: "position", Type: input.Number(), OmitEmpty: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

var RoomModel = model.Definition{
	Name: "room",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "floor_id", Type: input.Select(), Ref: &FloorModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "code", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 20}},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 120}},
		{Name: "notes", Type: input.Textarea(), OmitEmpty: true},
		{Name: "is_active", Type: input.Checkbox(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

var EquipmentModel = model.Definition{
	Name: "equipment",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 80}},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

// RoomEquipmentModel: qué equipamiento tiene un espacio. PK compuesta.
var RoomEquipmentModel = model.Definition{
	Name: "room_equipment",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "equipment_id", Type: model.Text(), Ref: &EquipmentModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
	},
}

// RoomCategoryModel: para qué categorías está habilitado un espacio.
// category_id es un id BLANDO de otro módulo (sin FK).
var RoomCategoryModel = model.Definition{
	Name: "room_category",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},
	},
}

// RoomShiftModel: un turno de ocupación. specific_date == 0 → SEMANAL (aplica cada
// day_of_week); specific_date > 0 → FECHADO (medianoche UTC en segundos, aplica solo ese día).
// start_min/end_min: minutos locales desde medianoche, 0..1440, start < end.
// occupant_id vacío = ocupante externo (occupant_label obligatorio, texto libre).
// occupant_id no vacío = occupant_label es la COPIA del nombre al guardar (snapshot).
var RoomShiftModel = model.Definition{
	Name: "room_shift",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text(), Ref: &RoomModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "category_id", Type: model.Text(), NotNull: true},
		{Name: "occupant_id", Type: model.Text(), OmitEmpty: true},
		{Name: "occupant_label", Type: model.Text(), NotNull: true},
		{Name: "day_of_week", Type: model.Int()},
		{Name: "specific_date", Type: model.Int()},
		{Name: "start_min", Type: model.Int(), NotNull: true},
		{Name: "end_min", Type: model.Int(), NotNull: true},
		{Name: "is_active", Type: model.Bool(), NotNull: true},
		{Name: "updated_at", Type: model.Int(), OmitEmpty: true},
	},
}

// RoomShiftCancellationModel: una ocurrencia de un turno SEMANAL que no aplica en una fecha.
var RoomShiftCancellationModel = model.Definition{
	Name: "room_shift_cancellation",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "shift_id", Type: model.Text(), Ref: &RoomShiftModel, DB: &model.FieldDB{PK: true, RefColumn: "id"}, NotNull: true},
		{Name: "specific_date", Type: model.Int(), DB: &model.FieldDB{PK: true}, NotNull: true},
	},
}
```

**Proyección de formulario del turno** (lo que edita un usuario; el servicio la convierte a
`RoomShift`). Igual que `ReservationForm` en `appointment_booking`:

```go
// RoomShiftFormModel. date vacío = turno semanal (usa day_of_week).
// date "YYYY-MM-DD" = turno fechado (day_of_week se ignora).
// start/end "HH:MM".
var RoomShiftFormModel = model.Definition{
	Name: "room_shift_form",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "room_id", Type: input.Select(), NotNull: true},
		{Name: "category_id", Type: input.Select(), NotNull: true},
		{Name: "occupant_id", Type: input.Select(), OmitEmpty: true},
		{Name: "occupant_label", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Maximum: 120}},
		{Name: "day_of_week", Type: weekdayInput(), OmitEmpty: true},
		{Name: "date", Type: input.Date(), OmitEmpty: true},
		{Name: "start", Type: input.Hour(), NotNull: true},
		{Name: "end", Type: input.Hour(), NotNull: true},
	},
}
```

`weekdayInput()` vive en `model.go`: `input.Select(opts...)` con 7 opciones cuyas `Key` son las
constantes `"0"`…`"6"` (`WeekdaySunday = "0"` … `WeekdaySaturday = "6"`, exportadas) y `Value`
"Domingo" … "Sábado". Mismo patrón que `webtyp.com/input/gender.go`.

**Modelos de argumentos** (sin widgets, `model.*`). Todos llevan `tenant_id` (`model.Text()`)
como primer campo, salvo los públicos:

| Definition | Campos (además de `tenant_id`) |
|---|---|
| `list_floors_args` | — |
| `delete_floor_args` | `id` |
| `list_rooms_args` | `active_only` (`model.Bool()`) |
| `get_room_args` | `id` |
| `deactivate_room_args` | `id` |
| `list_equipment_args` | — |
| `delete_equipment_args` | `id` |
| `list_room_equipment_args` | `room_id` |
| `set_room_equipment_args` | `room_id`, `equipment_ids` (`model.StructSlice(&IdRefModel)`) |
| `list_room_categories_args` | `room_id` |
| `set_room_categories_args` | `room_id`, `category_ids` (`model.StructSlice(&IdRefModel)`) |
| `list_options_args` | — (usada por `list_categories` y `list_occupants`) |
| `list_room_shifts_args` | `room_id` |
| `delete_room_shift_args` | `id` |
| `cancel_shift_occurrence_args` | `shift_id`, `date` (`"YYYY-MM-DD"`) |
| `move_shift_occurrence_args` | `shift_id`, `date`, `target_room_id` |
| `list_room_day_args` | `room_id` (vacío = todos), `date` (vacío = hoy en `Deps.Timezone`) |
| `list_board_args` | `category_id` (opcional), `equipment_id` (opcional) |
| `find_free_rooms_args` | `date`, `start`, `end` (`"HH:MM"`), `category_id` (opcional) |
| `list_public_availability_args` | `date`, `category_id` (opcional). **Sin `tenant_id`**: usa `Deps.TenantID` |

`IdRefModel` = `Definition{Name: "id_ref", Fields: {{Name: "id", Type: model.Text(), NotNull: true}}}`.

**Modelos de salida** (solo `model.*`, sin widgets):

| Definition | Campos |
|---|---|
| `option` | `id`, `label` — filas de `list_categories` / `list_occupants` |
| `day_shift` | `shift_id`, `room_id`, `room_code`, `category_id`, `category_label`, `occupant_id`, `occupant_label`, `start` (`"HH:MM"`), `end`, `is_weekly` (`model.Bool()`) |
| `board_room` | `room_id`, `code`, `name`, `floor_name`, `floor_position` (Int), `category_labels` (texto unido con `", "`), `equipment_labels` (ídem), `status`, `current_occupant_label`, `current_until` (`"HH:MM"`), `next_start`, `next_occupant_label`, `today` (`"YYYY-MM-DD"`) |
| `free_room` | `room_id`, `code`, `name`, `floor_name`, `equipment_labels` |
| `public_slot` | `room_code`, `room_name`, `floor_name`, `start`, `end` — **nunca** datos del ocupante |

Constantes de estado del tablero (exportadas; los únicos valores de `board_room.status`):
`BoardStatusFree = "free"`, `BoardStatusBusy = "busy"`, `BoardStatusClosed = "closed"`.

**Errores de dominio** (texto exacto):

```go
var (
	ErrNotFound            = fmt.Err("room_layout: not found")
	ErrTenantRequired      = fmt.Err("room_layout: tenant_id is required")
	ErrCodeAlreadyExists   = fmt.Err("room_layout: room code already exists in this tenant")
	ErrFloorInUse          = fmt.Err("room_layout: floor still has rooms")
	ErrInvalidRange        = fmt.Err("room_layout: start must be before end, both within 00:00 and 24:00")
	ErrInvalidDate         = fmt.Err("room_layout: invalid date, expected YYYY-MM-DD")
	ErrInvalidWeekday      = fmt.Err("room_layout: day_of_week must be 0..6")
	ErrUnknownCategory     = fmt.Err("room_layout: unknown category")
	ErrUnknownOccupant     = fmt.Err("room_layout: unknown occupant")
	ErrOccupantRequired    = fmt.Err("room_layout: occupant_id or occupant_label is required")
	ErrCategoryNotAllowed  = fmt.Err("room_layout: room is not enabled for this category")
	ErrCategoryInUse       = fmt.Err("room_layout: category still has active shifts in this room")
	ErrRoomOverlap         = fmt.Err("room_layout: room already has a shift in that time range")
	ErrOccupantOverlap     = fmt.Err("room_layout: occupant already has a shift in that time range")
	ErrOutsideBounds       = fmt.Err("room_layout: shift is outside the establishment's opening hours for that date")
	ErrNotWeekly           = fmt.Err("room_layout: only a weekly shift can have one occurrence cancelled")
)

type ValidationError struct{ Err error } // → 400, igual que patient_directory
```

Mapeo a estado HTTP: `ValidationError`, `ErrInvalid*`, `ErrUnknown*`, `ErrOccupantRequired`,
`ErrCategoryNotAllowed`, `ErrOutsideBounds`, `ErrNotWeekly` → 400 · `ErrNotFound` → 404 ·
`ErrCodeAlreadyExists`, `ErrFloorInUse`, `ErrCategoryInUse`, `ErrRoomOverlap`,
`ErrOccupantOverlap` → 409 · resto → 500.

**Topics de eventos** (`<módulo>.<entidad>.<verbo>`; `tenant_id` va en el payload):
`room_layout.floor.saved`, `room_layout.floor.deleted`, `room_layout.room.saved`,
`room_layout.room.deactivated`, `room_layout.equipment.saved`, `room_layout.equipment.deleted`,
`room_layout.shift.saved`, `room_layout.shift.deleted`, `room_layout.shift.occurrence_cancelled`,
`room_layout.shift.occurrence_moved`. Constantes `TopicFloorSaved`, etc.

---

## 4. Dependencias y constructor — `module.go`

```go
// CategoryReader: las categorías para las que un espacio puede habilitarse.
// Key = id de la categoría, Value = etiqueta visible.
type CategoryReader interface {
	CategoryOptions(tenantID string) ([]fmt.KeyValue, error)
}

// OccupantReader: quiénes pueden ocupar un turno. Key = id, Value = nombre visible.
type OccupantReader interface {
	OccupantOptions(tenantID string) ([]fmt.KeyValue, error)
}

// BoundsReader: horario de apertura del establecimiento para una fecha (medianoche UTC, s).
// Mismo contrato que appointment_booking.BoundsReader; business_calendar lo satisface.
type BoundsReader interface {
	GetDayBounds(date int64) (time.DayBounds, error) // webtyp.com/time
}

type Deps struct {
	IDs        model.IDGenerator // requerido
	Categories CategoryReader    // requerido
	Occupants  OccupantReader    // requerido
	TenantID   string            // requerido — tenant por defecto y el de la consulta pública
	Timezone   string            // requerido — IANA, p. ej. "America/Santiago"
	Bounds     BoundsReader      // opcional — nil = time.Unbounded() cada día
	Publisher  events.Publisher  // opcional — nil no publica
	Clock      func() int64      // opcional — nil = time.Now (nanosegundos UTC); inyectable en tests
}

func New(db *orm.DB, deps Deps) (*Module, error)
```

`New` devuelve error, con el texto exacto, si falta uno requerido:
`room_layout: Deps.IDs is required`, `room_layout: Deps.Categories is required`,
`room_layout: Deps.Occupants is required`, `room_layout: Deps.TenantID is required`,
`room_layout: Deps.Timezone is required`. `New` **no** toca el esquema.

---

## 5. Reglas de dominio — `module.go`, `shift.go`, `day.go`

Métodos de servicio exportados (firma exacta; `tenantID` primero):

```go
// floor.go
SaveFloor(f Floor) (Floor, error)                 // id vacío crea; si no, actualiza
ListFloors(tenantID string) ([]Floor, error)      // orden: position asc, luego name asc
DeleteFloor(tenantID, id string) error            // ErrFloorInUse si algún room lo referencia
// room.go
SaveRoom(r Room) (Room, error)                    // code único por tenant → ErrCodeAlreadyExists
GetRoom(tenantID, id string) (Room, error)
ListRooms(tenantID string, activeOnly bool) ([]Room, error) // orden: code asc
DeactivateRoom(tenantID, id string) error
// equipment.go
SaveEquipment(e Equipment) (Equipment, error)
ListEquipment(tenantID string) ([]Equipment, error) // orden: name asc
DeleteEquipment(tenantID, id string) error          // en una Tx borra también sus room_equipment
ListRoomEquipment(tenantID, roomID string) ([]Equipment, error)
SetRoomEquipment(tenantID, roomID string, equipmentIDs []string) error // reemplaza el conjunto, en una Tx
// category.go
ListRoomCategories(tenantID, roomID string) ([]fmt.KeyValue, error)  // etiquetas vía CategoryReader
SetRoomCategories(tenantID, roomID string, categoryIDs []string) error
// shift.go
SaveShift(tenantID string, f RoomShiftForm) (RoomShift, error)
ListRoomShifts(tenantID, roomID string) ([]RoomShift, error) // activos; semanales primero (dow, start), luego fechados (fecha, start)
DeleteShift(tenantID, id string) error                        // en una Tx borra también sus cancelaciones
CancelShiftOccurrence(tenantID, shiftID string, date int64) error
MoveShiftOccurrence(tenantID, shiftID string, date int64, targetRoomID string) (RoomShift, error)
// day.go
ListRoomDay(tenantID, roomID string, date int64) ([]DayShift, error)
ListBoard(tenantID, categoryID, equipmentID string) ([]BoardRoom, error)
FindFreeRooms(tenantID string, date int64, startMin, endMin int, categoryID string) ([]FreeRoom, error)
ListPublicAvailability(date int64, categoryID string) ([]PublicSlot, error)
```

`date int64` = medianoche UTC en segundos. Conversión desde `"YYYY-MM-DD"`:
`nano, err := time.ParseDate(s)` → `nano / 1_000_000_000`; error → `ErrInvalidDate`. Desde
`"HH:MM"`: `time.ParseTime(s)` → minutos. Estas dos conversiones viven en **un solo** archivo,
`convert.go` (funciones no exportadas `parseDate`, `parseHour`, `formatHour(min int) string`), y se
reutilizan en todos lados.

### 5.1 `SaveShift` — orden exacto de validación

1. `f.Validate(action)` → `ValidationError`.
2. `room := GetRoom(tenantID, f.RoomId)`; inactivo o inexistente → `ErrNotFound`.
3. `start, end := parseHour(f.Start), parseHour(f.End)`; `0 <= start < end <= 1440`, si no →
   `ErrInvalidRange`.
4. Semanal si `f.Date == ""`: `day_of_week` en 0..6, si no → `ErrInvalidWeekday`;
   `specific_date = 0`. Fechado si no: `specific_date = parseDate(f.Date)`; `day_of_week =
   time.Weekday(specific_date)`.
5. `category_id` debe estar en `Categories.CategoryOptions(tenantID)`, si no →
   `ErrUnknownCategory`; y en `room_category` de ese espacio, si no → `ErrCategoryNotAllowed`.
6. Ocupante: si `occupant_id != ""` debe estar en `Occupants.OccupantOptions(tenantID)` (si no →
   `ErrUnknownOccupant`) y `occupant_label` **se sobrescribe** con su `Value`. Si
   `occupant_id == ""`, `occupant_label` (sin espacios al borde) no puede quedar vacío → si no,
   `ErrOccupantRequired`.
7. Fechado: `bounds := Bounds.GetDayBounds(specific_date)` (o `time.Unbounded()` si `Bounds ==
   nil`); si `!bounds.Open || start < bounds.OpenMin || end > bounds.CloseMin` → `ErrOutsideBounds`.
   Semanal: no se valida contra bounds (varían por fecha); los feriados se resuelven al leer (§5.3).
8. Solapamiento de espacio → `ErrRoomOverlap`; solapamiento de ocupante (solo si
   `occupant_id != ""`, en **cualquier** espacio) → `ErrOccupantOverlap`. Regla en §5.2. Al
   actualizar, el propio turno se excluye.
9. `Create` o `Update` (condición `id` + `tenant_id`), publicar `TopicShiftSaved`.

### 5.2 Solapamiento

Dos rangos `[a1,a2)` y `[b1,b2)` se solapan si `a1 < b2 && b1 < a2` (semiabiertos: 09:00–13:00 y
13:00–15:00 **no** se solapan). Un candidato C choca con un turno activo existente E del mismo
espacio (o del mismo ocupante) cuando se solapan los minutos **y** además:

| C \ E | E semanal | E fechado |
|---|---|---|
| **C semanal (dow d)** | `E.day_of_week == d` | `time.Weekday(E.specific_date) == d` y `E.specific_date >= hoy` |
| **C fechado (fecha D)** | `E.day_of_week == time.Weekday(D)` y E **no** tiene cancelación en D | `E.specific_date == D` |

`hoy` = fecha local actual calculada con `localNow` (§5.4).

### 5.3 Resolución de un día — `shiftsOn(tenantID string, date int64) ([]RoomShift, error)`

Función no exportada, **única** fuente de verdad que usan `ListRoomDay`, `ListBoard`,
`FindFreeRooms`, `ListPublicAvailability` y el chequeo de solapamiento:

1. `bounds` del día (o `Unbounded`). Si `!bounds.Open` → lista vacía (día cerrado).
2. Turnos activos semanales con `day_of_week == time.Weekday(date)` **sin** cancelación en `date`,
   más turnos activos fechados con `specific_date == date`.
3. Descartar los de espacios inactivos.
4. Orden: `room_id`, luego `start_min`.

### 5.4 Hora local — `localNow() (date int64, minute int)`

`nowSec := clock()/1e9`. `D := time.MidnightUTC(nowSec)`. `m := (nowSec -
time.LocalMinutesToUnixUTC(D, 0, tz)) / 60`. Si `m < 0`: `D -= 86400` y recalcular; si
`m >= 1440`: `D += 86400` y recalcular. Test obligatorio con `Clock` fijo y
`Timezone = "America/Santiago"` cerca de medianoche UTC (p. ej. 02:30 UTC, que en Santiago es el
día anterior).

### 5.5 `ListBoard`

Para cada espacio activo (filtrado por `categoryID` en `room_category` y por `equipmentID` en
`room_equipment` cuando no están vacíos), con los turnos de `shiftsOn(hoy)`:
- día cerrado → `status = closed`, sin ocupante;
- hay un turno con `start <= ahora < end` → `busy`, `current_occupant_label`, `current_until`;
- si no → `free`.
- `next_*` = el primer turno con `start > ahora` (vacío si no hay).
- Orden: `floor_position`, `floor_name`, `code`.

### 5.6 `FindFreeRooms`

Espacios activos, habilitados para `categoryID` (si no está vacío), sin ningún turno de
`shiftsOn(date)` que se solape con `[startMin,endMin)` y con `[startMin,endMin)` dentro de los
bounds del día (si no, lista vacía). Orden como el tablero.

### 5.7 `ListPublicAvailability`

Tenant = `Deps.TenantID`. Por cada espacio activo habilitado para `categoryID`: los **huecos
libres** del día = `[bounds.OpenMin, bounds.CloseMin)` menos los turnos, fusionando adyacentes,
descartando huecos de menos de 15 minutos (`MinPublicSlotMinutes = 15`, constante exportada).
**Nunca** incluye `occupant_*` ni `category_*`.

### 5.8 `CancelShiftOccurrence` y `MoveShiftOccurrence`

- Cancelar: el turno debe ser semanal (`ErrNotWeekly`) y aplicar ese día de la semana
  (`ErrInvalidDate` si no); inserta `room_shift_cancellation` (idempotente: si ya existe, no hace
  nada y no falla).
- Mover, **dentro de `db.Tx`**: si el turno es semanal → cancelar la ocurrencia en `date` y crear
  un turno **fechado** en `targetRoomID` con la misma categoría, ocupante y minutos; si es fechado →
  actualizar su `room_id`. En ambos casos, validar el espacio destino con las reglas 5–8 de §5.1.
  Si falla, la Tx se revierte (el turno original queda intacto). Publicar
  `TopicShiftOccurrenceMoved`.

### 5.9 Categorías de un espacio

`SetRoomCategories` valida cada id contra `CategoryOptions` (`ErrUnknownCategory`) y rechaza
quitar una categoría que todavía tenga turnos activos en ese espacio (`ErrCategoryInUse`).
Reemplaza el conjunto en una `db.Tx`.

---

## 6. Ops — `ops.go`

| Op | Recurso | Acción | Accepts | Respuesta |
|---|---|---|---|---|
| `list_floors` | floor | `model.Read` | `ListFloorsArgs` | `FloorList` |
| `save_floor` | floor | `model.Create\|model.Update` | `Floor` | `Floor` |
| `delete_floor` | floor | `model.Delete` | `DeleteFloorArgs` | 200 |
| `list_rooms` | room | `model.Read` | `ListRoomsArgs` | `RoomList` |
| `get_room` | room | `model.Read` | `GetRoomArgs` | `Room` |
| `save_room` | room | `model.Create\|model.Update` | `Room` | `Room` |
| `deactivate_room` | room | `model.Update` | `DeactivateRoomArgs` | 200 |
| `list_equipment` | equipment | `model.Read` | `ListEquipmentArgs` | `EquipmentList` |
| `save_equipment` | equipment | `model.Create\|model.Update` | `Equipment` | `Equipment` |
| `delete_equipment` | equipment | `model.Delete` | `DeleteEquipmentArgs` | 200 |
| `list_room_equipment` | room | `model.Read` | `ListRoomEquipmentArgs` | `EquipmentList` |
| `set_room_equipment` | room | `model.Update` | `SetRoomEquipmentArgs` | 200 |
| `list_categories` | room | `model.Read` | `ListOptionsArgs` | `OptionList` |
| `list_room_categories` | room | `model.Read` | `ListRoomCategoriesArgs` | `OptionList` |
| `set_room_categories` | room | `model.Update` | `SetRoomCategoriesArgs` | 200 |
| `list_occupants` | room_shift | `model.Read` | `ListOptionsArgs` | `OptionList` |
| `list_room_shifts` | room_shift | `model.Read` | `ListRoomShiftsArgs` | `RoomShiftFormList` |
| `save_room_shift` | room_shift | `model.Create\|model.Update` | `RoomShiftForm` | `RoomShiftForm` |
| `delete_room_shift` | room_shift | `model.Delete` | `DeleteRoomShiftArgs` | 200 |
| `cancel_shift_occurrence` | room_shift | `model.Update` | `CancelShiftOccurrenceArgs` | 200 |
| `move_shift_occurrence` | room_shift | `model.Create\|model.Update` | `MoveShiftOccurrenceArgs` | `RoomShiftForm` |
| `list_room_day` | room_shift | `model.Read` | `ListRoomDayArgs` | `DayShiftList` |
| `list_board` | room | `model.Read` | `ListBoardArgs` | `BoardRoomList` |
| `find_free_rooms` | room | `model.Read` | `FindFreeRoomsArgs` | `FreeRoomList` |
| `list_public_availability` | — | **`.Public()`** | `ListPublicAvailabilityArgs` | `PublicSlotList` |

- Constantes `OpListFloors = "list_floors"`, etc., una por fila.
- Un `tenant_id` vacío en los args usa `m.tenantID` (igual que `patient_directory.opListPatients`).
- `list_room_shifts` devuelve **formularios** (`RoomShiftForm`, con `date`/`start`/`end` como
  texto) porque su único consumidor es la vista de edición; la conversión `RoomShift →
  RoomShiftForm` es una función no exportada en `convert.go`.
- `list_public_availability` es la **única** op pública. Su test verifica que la ruta registrada
  en `router/mock` tiene acceso público y que la respuesta no contiene la etiqueta del ocupante
  sembrado.

---

## 7. Vistas — `view.go` (paquete raíz: solo `view`+`model`+`router`)

- `func (f *Floor) Item() view.Item` → `{ID, Label: Name}`.
- `func (r *Room) Item() view.Item` → `{ID, Label: Code + " · " + Name}`.
- `func (e *Equipment) Item() view.Item` → `{ID, Label: Name}`.
- `func (s *RoomShiftForm) Item() view.Item` → `Label: occupant_label`, `Description`: si es
  semanal `"<Día> <start>–<end>"` (día por nombre desde las opciones de `weekdayInput`), si no
  `"<date> <start>–<end>"`.
- `NewFloorView(caller)`, `NewRoomView(caller)`, `NewEquipmentView(caller)`: `view.NewCallerLister`
  + `view.New`, con `Ops{Module: ModelName, List, Save, Delete}` (`NewRoomView` **sin** Delete: un
  espacio se desactiva, no se borra).
- `NewShiftView(caller router.Caller, tenantID, roomID string) view.Presenter`: lister propio
  acotado al espacio, mismo patrón que `employeeServiceConfigLister` en
  <https://github.com/veltylabs/appointment_booking/blob/main/lister.go> (List llama
  `list_room_shifts` con `RoomId`; Save llama `save_room_shift` con `RoomId` forzado al del
  contexto).

---

## 8. `ui/` — pantalla del módulo

Archivos: `ui/module.go`, `ui/browser.go`, `ui/board.go`, `ui/freesearch.go`, `ui/shifts.go`,
`ui/assign.go`, `ui/catalogs.go`, `ui/css.go` (`//go:build !wasm`), `ui/svg.go` (`//go:build
!wasm`).

```go
// ui/module.go
const ID = "room_layout"
const DefaultLabel = "Espacios"

// ui/browser.go
type Option func(*options)
func WithLabel(label string) Option // la app escribe "Boxes"; un hotel "Habitaciones"
func Browser(caller router.Caller, ids model.IDGenerator, tenantID string, opts ...Option) (platformd.UIModule, error)
```

`Browser` devuelve `platformd.NewUIModule(ID, label, svg.Icon(ID), tabs)` con
`webtyp.com/components/decktabs` y **seis** pestañas, en este orden:

| Pestaña (label) | Archivo | Contenido |
|---|---|---|
| Tablero | `board.go` | Dos `select` arriba (categoría ← `list_categories`, equipamiento ← `list_equipment`; opción vacía "Todas"/"Todo"); al cambiar, recarga `list_board`. Agrupa por `floor_name` (un `H2` por nivel) y dibuja cada espacio con `webtyp.com/components/contentcard`: Header = `code · name`; Body = `Libre` / `Ocupado — <occupant> hasta <current_until>` / `Cerrado`, y debajo `Próximo: <next_start> <next_occupant>` si existe; Footer = `category_labels` y `equipment_labels`. El estado se pinta con **una parte por estado** en la línea de estado (`PartStatusFree`, `PartStatusBusy`, `PartStatusClosed`, declaradas en `css.go`); no se usan `widget.State` para esto, porque no son estados de interacción. Un botón "Ver día" en el Footer despliega la lista de `list_room_day` de ese espacio. |
| Buscar libre | `freesearch.go` | Formulario generado con `webtyp.com/form` sobre `FindFreeRoomsArgs` (fecha, desde, hasta, categoría con `form.SetOptions("category_id", …)`). "Buscar" llama `find_free_rooms` y dibuja una `contentcard` por resultado con botón "Asignar", que despliega: select de ocupante (← `list_occupants`) **o** texto "Ocupante externo", y "Guardar" → `save_room_shift` con `date` fijado (turno fechado). |
| Turnos | `shifts.go` | `webtyp.com/layout/crudview` con `Context` = select de espacio (← `list_rooms` activos) y `Presenter = NewShiftView(caller, tenantID, roomID)`; al cambiar de espacio se reconstruye, igual que `ServiceConfigView` en `appointment_booking/ui/serviceconfigview.go`. Tras construir, `cv.Form.SetOptions("room_id", …)`, `("category_id", …)` (solo las de ese espacio, ← `list_room_categories`) y `("occupant_id", …)`. |
| Espacios | `catalogs.go` | `crudview` sobre `NewRoomView`; `cv.Form.SetOptions("floor_id", …)` desde `list_floors`. |
| Áreas y equipos | `assign.go` | Select de espacio + dos listas de casillas (categorías ← `list_categories`, marcadas según `list_room_categories`; equipamiento ← `list_equipment`, marcadas según `list_room_equipment`) + botón "Guardar" que llama `set_room_categories` y `set_room_equipment`. Un 409 (`ErrCategoryInUse`) se muestra como mensaje en la pestaña, no en consola. |
| Niveles y equipamiento | `catalogs.go` | Dos `crudview` apilados: `NewFloorView` y `NewEquipmentView` (con `ParentID` distintos: `ID+".floors"` y `ID+".equipment"`). |

- Estilos solo con el DSL `webtyp.com/widget/style` en `ui/css.go` (patrón:
  `appointment_booking/ui/css.go`). Cada vista declara `widget.Name`, `WidgetName()`,
  `WidgetKind()`. Nada de CSS en línea ni `style=` en el DOM.
- Los ids de elementos los genera `dom`; nunca se componen a mano.
- `ui/svg.go`: un ícono de planta/cuadrícula de un solo `path`, `currentColor`, igual que
  `patient_directory/ui/svg.go`.

---

## 9. `migrate/`, `seed/`, `web/`

- `migrate/migrate.go`: `Migrate(conn ddl.Execer, c ddl.Compiler) error` →
  `ddl.New(conn, c).Sync(&Floor{}, &Room{}, &Equipment{}, &RoomEquipment{}, &RoomCategory{},
  &RoomShift{}, &RoomShiftCancellation{})` (si `Sync` no acepta varios, una llamada por tabla en ese
  orden). Test en `migrate/migrate_test.go` como el de `patient_directory`.
- `seed/seed.go`: `Load(m *roomlayout.Module, tenantID string) (Data, error)`, escribiendo **solo**
  a través de los métodos del módulo: 2 niveles ("Primer piso", "Segundo piso"), 5 espacios
  (`B-101`, `B-102`, `B-103` en el primero; `B-201`, `B-202` en el segundo), 3 equipos
  ("Impresora", "Electrocardiógrafo", "Camilla ginecológica"), categorías asignadas y 6 turnos
  semanales (lunes a viernes, mañana y tarde) que dejan al menos un espacio libre cada tarde.
  `seed` no conoce ids reales de categorías ni de ocupantes: `Load` recibe los que va a usar
  mediante `LoadOptions{CategoryIDs []string; OccupantIDs []string}` como tercer argumento.
- `web/client.go` (`//go:build wasm`, `package main`): demo en el navegador como
  `patient_directory/web/client.go`, con `storage/mem`, `router/loopback.WithTenant`,
  `events/mock`, `unixid`. Las interfaces de lectura se satisfacen con dos structs locales del
  demo que devuelven listas fijas (3 categorías: "Medicina general", "Cardiología", "Ginecología";
  4 ocupantes), `Timezone: "America/Santiago"`, `Bounds: nil`.

---

## 10. Tests — `tests/`

Un archivo por tema: `setup_test.go` (fábrica con `mem`, lectores falsos, `Clock` fijo),
`floor_test.go`, `room_test.go`, `equipment_test.go`, `category_test.go`, `shift_test.go`,
`overlap_test.go`, `day_test.go`, `board_test.go`, `free_test.go`, `public_test.go`,
`localnow_test.go`, `ops_test.go`, `tenant_test.go`, `seed_test.go`, `widgets_test.go`,
`ui_view_wasm_test.go` (`//go:build wasm`).

Casos obligatorios (cada uno afirma resultados, no solo "sin error"):

1. `New` falla con el texto exacto por cada dependencia requerida ausente.
2. `SaveRoom` con `code` repetido en el mismo tenant → `ErrCodeAlreadyExists`; en otro tenant → ok.
3. `DeleteFloor` con espacios → `ErrFloorInUse`.
4. `SaveShift`: cada error de §5.1 con su sentinela exacta; ocupante con id → `occupant_label`
   queda con el nombre del lector aunque el cliente mande otro.
5. Solapamiento: las 4 celdas de la tabla de §5.2, más el caso límite 09:00–13:00 / 13:00–15:00
   (no choca), más "fechado en la fecha de una ocurrencia cancelada" (no choca).
6. `ErrOccupantOverlap` entre dos espacios distintos.
7. `shiftsOn`: semanal + fechado + cancelado + espacio inactivo + día cerrado por `Bounds`.
8. `ListBoard` con `Clock` fijo: un espacio `busy` con `current_until` correcto, uno `free` con
   `next_start`, y `closed` cuando `Bounds` dice cerrado.
9. `MoveShiftOccurrence`: semanal → cancelación + turno fechado nuevo; destino sin la categoría →
   error **y** el turno original sigue aplicando ese día (la Tx se revirtió).
10. `ListPublicAvailability`: huecos fusionados, huecos < 15 min descartados, sin datos del
    ocupante.
11. `localNow` a las 02:30 UTC con `America/Santiago` devuelve el día anterior.
12. Aislamiento de tenant: el tenant B no lee, no modifica ni borra filas del tenant A por ningún
    método ni op.
13. Cada op vía `router/mock`: ruteo correcto y `.Requires` declarado; `list_public_availability`
    pública.
14. `events/mock` recibe el topic y el payload tipado correctos en cada mutación.
15. Widgets: `form.New(id, &RoomShiftForm{}, ids)` produce exactamente los inputs `room_id`,
    `category_id`, `occupant_id`, `occupant_label`, `day_of_week`, `date`, `start`, `end` (y lo
    mismo para `Room`, `Floor`, `Equipment`).
16. `ui_view_wasm_test.go`: `ui.Browser` sobre `loopback` + seed renderiza las seis pestañas y el
    tablero muestra una tarjeta por espacio activo.

---

## 11. Documentación

- `docs/ARCHITECTURE.md` (español): alcance, vocabulario (tabla de §0), entidades, reglas §5,
  tabla de ops, ejemplo de raíz de composición (con los dos adaptadores de lectura).
- `docs/diagrams/database.md`: ERD mermaid de las 7 tablas.
- `README.md`: inicio rápido, tabla de ops, archivos clave.
- `AGENTS.md`: **no** tocar lo que está sobre "Domain-specific notes".

---

## 12. Etapas

| # | Etapa | Archivos | Criterio de aceptación |
|---|---|---|---|
| E1 | Modelo + esquema | borrar `room_layout.go`; `model.go`, `model_orm.go` (ormc), `convert.go`, `migrate/` | `test ! -f room_layout.go`; `grep -rn "RoomLayout" --include=*.go .` vacío; `gotest ./migrate/...` verde |
| E2 | Dominio | `module.go`, `floor.go`, `room.go`, `equipment.go`, `category.go`, `shift.go`, `day.go` | tests 1–12 verdes |
| E3 | Ops | `ops.go` | tests 13–14 verdes; `var _ router.OperationModule = (*Module)(nil)` presente |
| E4 | Vistas + seed | `view.go`, `seed/seed.go` | test 15 y `seed_test.go` verdes |
| E5 | UI + demo | `ui/*`, `web/client.go` | test 16 verde; `webtyp` en la raíz abre el demo sin errores de consola |
| E6 | Docs + cierre | `docs/*`, `README.md` | `grep -rn "TODO\|FIXME\|map\[" --include=*.go .` vacío; `grep -rn '"errors"\|"strings"\|"strconv"\|"encoding/json"\|"net/http"' --include=*.go .` vacío; `gotest` verde en ambos carriles |
