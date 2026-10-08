---
PLAN: "fix: detect sentinel errors without == between interfaces (no reflection in wasm)"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 8526907848916096680
---

# Plan — `room_layout`: errores centinela sin `==` entre interfaces

> Master: `webtyp/docs/NO_REFLECTION_MASTER_PLAN.md` (ola 3). Doctrina: skill `api-design`.

## 1. El problema

En TinyGo, `==`, `!=` y `switch` entre valores de interfaz compilan a `runtime.interfaceEqual`, que
llama a `reflectValueEqual(reflectlite.ValueOf(x), reflectlite.ValueOf(y))`. `error` es una interfaz:
cada `err == ErrX` mete `internal/reflectlite` (~9 KB) en el binario wasm. La regla del dueño es que
el código que compila a wasm no use reflexión nunca. `errors.Is`/`errors.As` tampoco sirven: también
usan reflectlite.

## 2. La corrección — dos patrones, ninguno más

**A. Centinelas de otros paquetes** — usar su función de consulta:

| Antes | Después |
|---|---|
| `err == orm.ErrNotFound` | `orm.IsNotFound(err)` |
| `err != orm.ErrNotFound` | `!orm.IsNotFound(err)` |
| `err == storage.ErrNoRows` | `storage.IsNoRows(err)` |

**B. Centinelas propios de este paquete** — un tipo string no exportado; se afirma una vez y se
compara el valor concreto (comparación de strings, sin reflexión):

```go
// domainError is the concrete type of this package's sentinel errors. Code
// compares them by asserting this type and comparing the value: == between two
// error values compiles, under TinyGo, to runtime.interfaceEqual, which pulls
// internal/reflectlite into the wasm binary.
type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrNotFound domainError = "<texto actual>"
	// … uno por centinela, con su texto actual
)
```

- `<texto actual>`: el string exacto que devuelve hoy el centinela (`fmt.Err("a", "b")` une las
  palabras con un espacio: `"a b"`). Un test fija cada texto: los mensajes no cambian.
- Uso, por ejemplo al traducir errores a códigos:

```go
if e, ok := err.(domainError); ok {
	switch e {
	case ErrFloorInUse, ErrRoomOverlap:
		return conflict
	case ErrNotFound:
		return notFound
	}
}
if orm.IsNotFound(err) {
	return notFound
}
```

- Un `switch err { case ErrA: … }` pasa a `if e, ok := err.(domainError); ok { switch e { … } }`.
- Si un centinela propio se envuelve antes de compararlo (`fmt.Errf("…%v", ErrX)`), la comparación
  con `==` ya no funcionaba: dejarlo igual y anotarlo en el PR, no inventar otra detección.

## 3. Sitios a cambiar (inventario del 2026-10-08)

### Código de producción

- `ops.go:46` — `if err == ErrNotFound {`
- `ops.go:49` — `if err == ErrCodeAlreadyExists || err == ErrFloorInUse || err == ErrCategoryInUse || err == ErrRoomOverlap || err == ErrOccupantOverlap {`
- `ops.go:52` — `if _, ok := err.(*ValidationError); ok || err == ErrTenantRequired || err == ErrInvalidRange || err == ErrInvalidDate || err == ErrInvalidWe`

### Centinelas propios de este repo (patrón B)

- `model.go:57` — `ErrNotFound           = fmt.Err("room_layout: not found")`
- `model.go:58` — `ErrTenantRequired     = fmt.Err("room_layout: tenant_id is required")`
- `model.go:59` — `ErrCodeAlreadyExists  = fmt.Err("room_layout: room code already exists in this tenant")`
- `model.go:60` — `ErrFloorInUse         = fmt.Err("room_layout: floor still has rooms")`
- `model.go:61` — `ErrInvalidRange       = fmt.Err("room_layout: start must be before end, both within 00:00 and 24:00")`
- `model.go:62` — `ErrInvalidDate        = fmt.Err("room_layout: invalid date, expected YYYY-MM-DD")`
- `model.go:63` — `ErrInvalidWeekday     = fmt.Err("room_layout: day_of_week must be 0..6")`
- `model.go:64` — `ErrUnknownCategory    = fmt.Err("room_layout: unknown category")`
- `model.go:65` — `ErrUnknownOccupant    = fmt.Err("room_layout: unknown occupant")`
- `model.go:66` — `ErrOccupantRequired   = fmt.Err("room_layout: occupant_id or occupant_label is required")`
- `model.go:67` — `ErrCategoryNotAllowed = fmt.Err("room_layout: room is not enabled for this category")`
- `model.go:68` — `ErrCategoryInUse      = fmt.Err("room_layout: category still has active shifts in this room")`
- `model.go:69` — `ErrRoomOverlap        = fmt.Err("room_layout: room already has a shift in that time range")`
- `model.go:70` — `ErrOccupantOverlap    = fmt.Err("room_layout: occupant already has a shift in that time range")`
- `model.go:71` — `ErrOutsideBounds      = fmt.Err("room_layout: shift is outside the establishment's opening hours for that date")`
- `model.go:72` — `ErrNotWeekly          = fmt.Err("room_layout: only a weekly shift can have one occurrence cancelled")`

### Tests (se migran igual: un solo camino también en los tests)

- `tests/board_test.go:115` — `if err != roomlayout.ErrCategoryNotAllowed {`
- `tests/room_test.go:36` — `if err != roomlayout.ErrCodeAlreadyExists {`
- `tests/overlap_test.go:58` — `if err != roomlayout.ErrRoomOverlap {`
- `tests/overlap_test.go:133` — `if err != roomlayout.ErrOccupantOverlap {`
- `tests/category_test.go:46` — `if err != roomlayout.ErrCategoryInUse {`
- `tests/tenant_test.go:24` — `if err != roomlayout.ErrNotFound {`
- `tests/floor_test.go:22` — `if err != roomlayout.ErrFloorInUse {`
- `tests/shift_test.go:32` — `if err != roomlayout.ErrInvalidRange {`
- `tests/shift_test.go:45` — `if err != roomlayout.ErrUnknownCategory {`
- `tests/shift_test.go:58` — `if err != roomlayout.ErrCategoryNotAllowed {`

Si encuentras otro `==`/`!=`/`switch` entre valores de interfaz con operandos no nil que no esté en la
lista, se migra igual. `x == nil` y `x != nil` están bien.

## 4. Tests

- Todos los tests existentes siguen verdes sin cambiar su intención.
- Un test que fija el `Error()` de cada centinela propio convertido (patrón B) contra su texto anterior.
- Si el paquete traduce errores a códigos/respuestas (por ejemplo en `ops.go`), un test por rama
  cambiada: el mismo error produce el mismo código que antes.
- `gotest` verde (vet, race, tests, wasm).

## 5. Criterios de aceptación

- `grep -rnE '(==|!=) *[A-Za-z_.]*Err[A-Za-z]*' --include=*.go . | grep -v '_temp/'` → vacío.
- `grep -rn 'switch err {' --include=*.go .` → vacío.
- `grep -rn 'errors.Is\|errors.As' --include=*.go .` → vacío.
- Ningún símbolo exportado nuevo: `git diff | grep '^+func [A-Z]'`.
- `gotest` verde.

## 6. Restricciones

Las de `AGENTS.md`, más: nada de `reflect`, `unsafe`, `errors.Is`/`errors.As`, ni `==`/`!=`/`switch`
entre valores de interfaz con operandos no nil. No tocar otros repos.
