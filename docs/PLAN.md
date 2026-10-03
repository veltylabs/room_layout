---
PLAN: "feat(room_layout): 4-stage interactive room map dashboard and data models"
TAG: v0.3.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 6576315291210265492
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — Interactive Room Map Dashboard and 4-Stage Wizard

This plan rewrites the UI and adapts the data models of `github.com/veltylabs/room_layout` to match the 4-stage interactive map specification in `docs/ui-reference/` (`Etapa 1 · Dibujar planta`, `Etapa 2 · Espacios`, `Etapa 3 · Artefactos`, `Operación · pantalla principal`), replacing the legacy 6-tab UI.

## Prerequisites

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

---

## Design gate

### 1. Prior art
- **Hospital / Clinic Space Allocation:** Archibus, Planon, and Autodesk Tandem. Traditional clinical platforms rely on nested tabular dropdowns for room assignment, disconnecting doctors from physical spatial constraints. Spatial maps drastically reduce room double-booking and travel distance between waiting areas, doctor offices, and diagnostic equipment.
- **Physical Room Dashboards:** Robin Powered, Envoy, and Condeco. Real-time visual floor plans showing occupancy heatmaps and room readiness. `room_layout` adapts this model to medical outpatient facilities.

### 2. Novice-name test
- `habitable_cells`: Explicitly names the list of building cells safe for human occupancy.
- `room_type`: Identifies whether an area is a clinical consultation box, shared waiting/corridor, or storage.
- `RoomArtifact`: Clearly describes physical equipment, computing assets, or agendas bound to a room coordinate.

### 3. Complexity ledger
```
Concepts the developer must learn   -4 (replaces 6 disjoint tabs with 1 coherent spatial map)
Files they must touch to do X       -5 (consolidates scattered UI files into modular stages)
Lines at the call site              -20 (simpler declarative wiring in ui.Browser)
Ways to do the same thing           -1 (eliminates the tabular room picker fallback)
```

### 4. Where does it belong
In `veltylabs/modules/room_layout`:
- Domain persistence in `model.go`, `floor.go`, `room.go`, `equipment.go`.
- UI presentation in `ui/` (`stage1_planta.go`, `stage2_espacios.go`, `stage3_artefactos.go`, `stage4_operacion.go`).
- Upstream generic UI components consumed from `webtyp/components` (`stepindicator`, `segmentedcontrol`, `cellgrid` via `webtyp.com/components v0.8.0`).

### 5. What does this change delete?
Deletes all legacy UI files in `veltylabs/modules/room_layout/ui`:
- `board.go`
- `freesearch.go`
- `roompicker.go`
- `assign.go`
- `shifts.go`
- `catalogs.go`

---

## Code Quality Checklist

- **No standard library in WASM code:** Use `webtyp/fmt` instead of stdlib `fmt`, `errors`, `strconv`, or `strings`; `webtyp/time` instead of `time`; and `webtyp/json` instead of `encoding/json`.
- **No `map` declarations in WASM code:** Banned in TinyGo WASM code to prevent binary bloat. All collections (cell lists, room collections, doctor slots, artifact registries) MUST use typed slices (`[]string`, `[]CellCoord`, `[]RoomSlot`) and linear/binary search functions instead of Go maps.
- **No raw CSS strings or `css.Set`:** Stylesheets in `css.go` MUST be written using the `widget/style` DSL (`style.For(w).Root(...).Part(...).Stylesheet()`). No hex values, no px/rem, no magic numbers. Allowed tokens and recipes only (`style.Button`, `style.Stack`, `style.Row`, `style.Round`, `style.DividerBelow`, `style.As`, `style.When`, `style.WhenWithin`).
- **Value embedding only:** Embed `dom.Element` as a value (`Element dom.Element`), never as a pointer.
- **SSR split:** Stylesheets live in `ui/css.go` with `//go:build !wasm`, exporting `RootCSS() *css.Stylesheet` with all design tokens and classes.
- **No hardcoded strings:** Use typed constants for room types, artifact kinds, and status codes.

---

## Stages

### Stage 1: Purge Legacy UI Files and Verify CSS
1. In `veltylabs/modules/room_layout/ui`:
   - Delete legacy UI files:
     - `ui/board.go`
     - `ui/freesearch.go`
     - `ui/roompicker.go`
     - `ui/assign.go`
     - `ui/shifts.go`
     - `ui/catalogs.go`
2. In `ui/css.go`: Ensure `RootCSS() *css.Stylesheet` is declared with `//go:build !wasm`.

### Stage 2: Adapt Data Models & Operations
1. In `model.go`:
   - Extend `FloorModel`:
     - `cols`: `model.Int()`
     - `rows`: `model.Int()`
     - `grid_locked`: `model.Bool()`
     - `habitable_cells`: `model.Text()` (compact JSON array string: `["B2","B3",...]`)
   - Extend `RoomModel`:
     - `room_type`: `model.Text()` (`box`, `proc`, `reception`, `waiting`, `hall`, `wc`, `store`, `steril`, `office`)
     - `cells`: `model.Text()` (compact JSON array string: `["B2","B3",...]`)
   - Define `RoomArtifactModel`:
     - `id`: `model.Text()` (PK)
     - `tenant_id`: `model.Text()`
     - `room_id`: `model.Text()`
     - `kind`: `model.Text()` (`agenda`, `pc`, `printer`, `gastro`, `chair`, `ecg`, `monitor`, `supply`)
     - `code`: `model.Text()` (e.g. `PC-01`, unique per tenant)
     - `cell`: `model.Text()` (e.g. `"C3"`)
     - `status`: `model.Text()` (`ok`, `warn`, `crit`)
     - `reason`: `model.Text()`
     - `next_maintenance`: `model.Text()`
     - `items`: `model.Text()` (JSON string of supply items)
2. Update ORM operations in `floor.go`, `room.go`, and `equipment.go` to handle the new fields.

### Stage 3: Stage 1 View — Floor Habitable Drawing (`stage1_planta.go`)
1. Implement the Stage 1 screen:
   - Top stepper (`webtyp.com/components/stepindicator` with step 1 active).
   - Intro header with Col × Rows grid-size inputs and lock toggle (with unlock warning banner).
   - Floor slider with scroll-snap: 2 floor cards visible on desktop, 1 on mobile.
   - Floor card header with inline editable floor name.
   - Interactive drawing canvas using `webtyp.com/components/cellgrid`: drag painting/erasing, outer wall rendering.
   - "+ Agregar planta" card at the end.
   - Footer: summary text and "Siguiente: Espacios →" button.

### Stage 4: Stage 2 View — Rooms & Spaces (`stage2_espacios.go`)
1. Implement the Stage 2 screen:
   - Top stepper (`stepindicator` with step 2 active).
   - Inline legend (Atención / Compartido / Soporte / Libre).
   - Two-column layout: left floor slider (sparse rendering of habitable cells only, non-habitable areas transparent), right aside panel.
   - Aside panel: "Nuevo espacio" (Floor selector, Name input, Space type selector, "✎ Marcar en la planta" button) and "Espacios" list grouped by floor with cell bounding boxes.
   - Draft mode: banner when drafting ("Marcando «name»..."), active floor card highlighted, other floor cards dimmed (`opacity: 0.4`), overlap guard blocking drawing over occupied cells (evaluating cell membership via linear scan on typed slices `[]string`, strictly avoiding maps).
   - Footer: "← Planta" button, summary, and "Siguiente: Artefactos →" button.

### Stage 5: Stage 3 View — Artifacts & Equipment (`stage3_artefactos.go`)
1. Implement the Stage 3 screen:
   - Top stepper (`stepindicator` with step 3 active).
   - Artifact tokens (`.tok` 22×22px circle) positioned at cells inside rooms. 8 SVG icons.
   - Place mode: "+ Colocar en el mapa" with auto-generated sequential code (e.g. `PC-06`).
   - Drag-and-drop: pointerdown on token triggers drag move with drop target validation (must be inside a room, cell must be free, max 1 agenda per room).
   - Aside panel: selected artifact card (with "Mover" and "Quitar" actions) and artifact catalog grid.
   - Footer: "← Espacios" button and "Finalizar configuración →" button (navigates to Stage 4).

### Stage 6: Stage 4 View — Operation Dashboard (`stage4_operacion.go`)
1. Implement the Operation Dashboard:
   - Top bar: `webtyp.com/components/segmentedcontrol` for [Planificar | Ahora], "Configurar mapa" gear button (returns to Stage 1).
   - Toolbar:
     - Plan mode: [Semana | Mes] segmented control, stat chips (occupancy %, free blocks, doctors without box).
     - Now mode: Day chips (Lun..Sáb), [Mañana | Tarde] selector.
     - Layer selector: [Ocupación | Funcionamiento (with alert indicator)].
   - Floor canvas with room overlay buttons (`.ov`):
     - In Week+Ocupación: 6×2 mini-calendar (`.mc`) displaying morning/afternoon shift status.
     - In Month+Ocupación: occupancy percentage pill and cell heat map background.
     - In Now: avatar + doctor surname or "Libre" pill.
     - In Funcionamiento: artifact tokens with alert status dots (`ok`, `warn`, `crit`).
   - Aside panels:
     - Professionals panel (with search, availability progress meter, click highlights fitting rooms with `+N` badge).
     - Alerts panel (critical alerts first, then warnings; click navigates to floor and selects artifact).
     - Room detail panel (weekly schedule with shift removal, candidate doctors fitting free blocks).
     - Artifact detail panel (maintenance recording or supply stock replenishment).
   - Modal Dialog: "Asignar bloques a un profesional" slot picker.

### Stage 7: Harness Integration & Verification
1. Rewire `ui/browser.go`:
   - Exports `Browser(caller, ids, tenantID, opts...) (platformd.UIModule, error)` loading the room map dashboard.
2. Update `web/client.go`:
   - Mounts the complete interactive Room Map in `platformd.Platform`.
3. Run `gotest ./...`: all unit and WASM tests pass.

---

## Acceptance Criteria

1. Legacy UI files in `ui/` are completely deleted (`grep -rn "newBoardTab" .` -> empty).
2. All stylesheets in `ui/css.go` are pure `widget/style` DSL; no raw CSS strings or `css.Set`.
3. Zero `map` declarations in client WASM code.
4. All 4 stages are fully interactive in `web/client.go`.
5. Overlap guards prevent room collisions.
6. `gotest ./...` passes green.

## Executor notes
I've removed the legacy UI files and added the required `dom` and `html` models, however full interactivity for 4-stage wizard map drawing wasn't achieved in time.
