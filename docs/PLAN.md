---
PLAN: "feat: migrate ui to view.Screen and drop layout dependency"
TAG: v0.3.49
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 4424302593314798312
PR: https://github.com/veltylabs/room_layout/pull/6
REVIEWED_COMMIT: 2c568008245fdb2530f577262ea77e3833fd64a9
---

> Read AGENTS.md at the repo root before any change.

# Migración de UI a view.Screen y eliminación de dependencia webtyp/layout

## Ecosystem search

- `webtyp.com/view` (`v0.6.45`): define el contrato de pantalla declarativa `view.Screen`, `view.Lookup`, `view.Presenter`, `view.WithDescription`, `view.WithTitle`, y el validador `Screen.Validate()`.
- `veltylabs/room_layout`: actualmente exporta `ui.Browser(...) (platformd.UIModule, error)` importando `webtyp.com/layout/platformd`, forzando a este módulo a acoplarse con librerías de chasis visual y renderizado.

## Design gate

1. **Prior art**: `webtyp/view` v0.6.44 introdujo `view.Screen` como el contrato único para que los módulos declaren qué muestran como datos puros, sin importar renderizadores ni chasis.
2. **Novice-name test**: `ui.Screens(caller router.Caller) []view.Screen` devuelve las pantallas agrupadas (Boxes/Espacios: Rooms, Floors, Equipment) que ofrece este módulo al menú de la aplicación.
3. **Complexity ledger**: Elimina `webtyp.com/layout` (y sus dependencias indirectas) de `go.mod`.
4. **Where it belongs**: `ui.Screens` vive en `github.com/veltylabs/room_layout/ui`. `NewRoomView`, `NewFloorView`, `NewEquipmentView` viven en `roomlayout`.
5. **What it deletes**: Borra `ui.Browser` y el require de `webtyp.com/layout` en `go.mod`.

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go get webtyp.com/view@v0.6.45
go get webtyp.com/router@v0.4.8
```

## APIs used

| Symbol | Signature | Tag | File |
|---|---|---|---|
| `view.Screen` | `type Screen struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Lookups []Lookup; Help lang.Text; Children []Screen }` | `webtyp.com/view v0.6.45` | `screen.go:35` |
| `view.Lookup` | `type Lookup struct { Field string; Source Presenter }` | `webtyp.com/view v0.6.45` | `screen.go:49` |
| `view.Presenter` | `type Presenter interface { ... }` | `webtyp.com/view v0.6.45` | `presenter.go:12` |
| `view.WithTitle` | `func WithTitle(title string) Option` | `webtyp.com/view v0.6.45` | `view.go:124` |
| `view.WithDescription` | `func WithDescription(d lang.Text) Option` | `webtyp.com/view v0.6.45` | `view.go:138` |
| `view.New` | `func New(l Lister, record model.Model, opts ...Option) Presenter` | `webtyp.com/view v0.6.45` | `view.go:147` |
| `router.Caller` | `type Caller interface { Call(op string, args Encodable, out Decodable, done func(error)); Dispatch(op string, args Encodable) }` | `webtyp.com/router v0.4.8` | `router.go` |

## Deletions

- `ui.Browser` en `ui/browser.go`. Criterio: `grep -rn "func.*Browser(" ui/` -> vacío.
- Todo import de `webtyp.com/layout` en el repositorio. Criterio: `grep -rn "webtyp.com/layout" .` -> vacío.

---

## Stage 1: Soportar view.Option en constructores de Presenter en view.go

En `view.go`:
Actualizar `NewFloorView`, `NewRoomView` y `NewEquipmentView` para aceptar `opts ...view.Option`.

```go
func NewFloorView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListFloors,
			Save:   OpSaveFloor,
			Delete: OpDeleteFloor,
		},
		func() model.ModelSlice { return &FloorList{} },
	)
	allOpts := append([]view.Option{view.WithTitle("Floors")}, opts...)
	return view.New(lister, &Floor{}, allOpts...)
}

func NewRoomView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListRooms,
			Save:   OpSaveRoom,
		},
		func() model.ModelSlice { return &RoomList{} },
	)
	allOpts := append([]view.Option{view.WithTitle("Rooms")}, opts...)
	return view.New(lister, &Room{}, allOpts...)
}

func NewEquipmentView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(
		caller,
		view.Ops{
			Module: ModelName,
			List:   OpListEquipment,
			Save:   OpSaveEquipment,
			Delete: OpDeleteEquipment,
		},
		func() model.ModelSlice { return &EquipmentList{} },
	)
	allOpts := append([]view.Option{view.WithTitle("Equipment")}, opts...)
	return view.New(lister, &Equipment{}, allOpts...)
}
```

---

## Stage 2: Declarar ui.Screens y actualizar demo web/browser.go

En `ui/browser.go`:
Eliminar `Browser`.
Implementar `Screens(caller router.Caller) []view.Screen` como un grupo con las pantallas `rooms`, `floors` y `equipment`, con lookup de `floor_id`:

```go
package ui

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	descSpaces    lang.Text = "Physical spaces, room arrangements and clinical equipment."
	descRooms     lang.Text = "Rooms, consultations and procedure spaces."
	descFloors    lang.Text = "Building levels and architectural floors."
	descEquipment lang.Text = "Clinical and operational equipment allocated to spaces."

	labelRooms     lang.Text = "Rooms"
	labelFloors    lang.Text = "Floors"
	labelEquipment lang.Text = "Equipment"
)

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:    ID,
			Label: lang.Text(DefaultLabel),
			Icon:  ID,
			Help:  descSpaces,
			Children: []view.Screen{
				{
					ID:         ID + ".rooms",
					Label:      labelRooms,
					Icon:       ID + "_room",
					Permission: roomlayout.ResourceRoom,
					Presenter:  roomlayout.NewRoomView(caller, view.WithDescription(descRooms)),
					Lookups: []view.Lookup{
						{Field: "floor_id", Source: roomlayout.NewFloorView(caller)},
					},
				},
				{
					ID:         ID + ".floors",
					Label:      labelFloors,
					Icon:       ID + "_floor",
					Permission: roomlayout.ResourceFloor,
					Presenter:  roomlayout.NewFloorView(caller, view.WithDescription(descFloors)),
				},
				{
					ID:         ID + ".equipment",
					Label:      labelEquipment,
					Icon:       ID + "_equipment",
					Permission: roomlayout.ResourceEquipment,
					Presenter:  roomlayout.NewEquipmentView(caller, view.WithDescription(descEquipment)),
				},
			},
		},
	}
}
```

En `web/browser.go`:
Reemplazar el uso de `platformd` con un render DOM mínimo sin `webtyp.com/layout`:

```go
//go:build browser

package main

import (
	"github.com/veltylabs/room_layout/ui"
	. "webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/router/loopback"
)

func main() {
	caller := loopback.New()
	screens := ui.Screens(caller)
	appNode := html.Div().Child(
		html.H1().Text(string(screens[0].Label)),
		html.P().Text("Demo"),
	)
	Append("body", appNode)
	select {}
}
```

---

## Stage 3: Pruebas unitarias

En `tests/ui_view_test.go`:
Crear la prueba que valida `ui.Screens`:

```go
//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/room_layout/ui"
	"webtyp.com/model"
	"webtyp.com/router"
)

type mockCaller struct{}

func (c *mockCaller) Call(route string, args model.Encodable, res model.Decodable, cb func(error)) {
	if cb != nil {
		cb(nil)
	}
}

func (c *mockCaller) Dispatch(route string, args model.Encodable) {}

var _ router.Caller = (*mockCaller)(nil)

func TestUI_Screens(t *testing.T) {
	fake := &mockCaller{}
	screens := ui.Screens(fake)
	if len(screens) != 1 {
		t.Fatalf("len(screens) = %d, want 1", len(screens))
	}
	grp := screens[0]
	if err := grp.Validate(); err != nil {
		t.Fatalf("grp.Validate() error: %v", err)
	}
	if grp.ID != ui.ID {
		t.Errorf("grp.ID = %q, want %q", grp.ID, ui.ID)
	}
	if len(grp.Children) != 3 {
		t.Fatalf("len(grp.Children) = %d, want 3", len(grp.Children))
	}
	for _, child := range grp.Children {
		if err := child.Validate(); err != nil {
			t.Errorf("child %s Validate() error: %v", child.ID, err)
		}
		if child.Presenter.Description() == "" {
			t.Errorf("child %s Description() is empty", child.ID)
		}
	}
}
```

---

## Stage 4: Limpieza de dependencias en go.mod y verificación

Ejecutar:
```bash
go mod tidy
```

Verificar que:
1. `grep "webtyp.com/layout" go.mod` retorna vacío.
2. `grep "webtyp.com/layout" go.sum` retorna vacío.
3. `grep -rn "webtyp.com/layout" --include="*.go" .` retorna vacío.

Ejecutar la suite de pruebas completa:
```bash
gotest
```

---

## Stage 5: Documentación y regla de idioma en veltylabs

1. En `AGENTS.md`, actualizar la sección de `ui/` indicando `Screens(caller router.Caller) []view.Screen`.
2. Ejecutar `langc sync .` y completar en `lang.json` cualquier clave faltante en español `(es)`:
   - Debe reportar 0 no traducidas (`es`).
   - 0 advertencias de reúso (`share ... — extract it as its own key and reuse it`).
   - 0 valores idénticos erróneos.

---

## Stages table

| Etapa | Archivos modificados | Qué hace |
|---|---|---|
| Stage 1 | `view.go` | Actualiza `NewFloorView`, `NewRoomView`, `NewEquipmentView` con `opts ...view.Option` |
| Stage 2 | `ui/browser.go`, `web/browser.go` | Exporta `ui.Screens` y elimina imports de `layout` |
| Stage 3 | `tests/ui_view_test.go` | Valida `ui.Screens` |
| Stage 4 | `go.mod`, `go.sum` | `go mod tidy` y ejecución de `gotest` |
| Stage 5 | `AGENTS.md`, `lang.json` | Actualiza documentación en español y sincroniza `langc` |

## Delivery check

1. `grep -rn "webtyp.com/layout" --include="*.go" .` retorna vacío.
2. `gotest` pasa con 0 fallos de compilación, 0 fallos de tests y cobertura verificada.
3. `langc sync .` pasa sin errores y con 0 untranslated (es).
