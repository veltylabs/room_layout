---
PLAN: "feat: a room's equipment may point to a registered device (device_id → device_manager), optional"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 9564745239310761274
PR: https://github.com/veltylabs/room_layout/pull/10
REVIEWED_COMMIT: 37e63f702250e3f132d5a6327ebc28aad81ec6e8
---

> Read AGENTS.md at the repo root before any change.

# PLAN — R1b: equipment linked to a device (D31)

## Why

Phase R1b of `veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`, decision D31: **a device is
registered once.** A box's equipment that has a network (a computer, a printer) is already registered
in `device_manager`; the equipment row may point to it with `device_id`. Equipment without a network
(a gastroscope) leaves it empty. `room_layout` starts to depend on `device_manager` (which does not
depend on `room_layout`).

The same shape already exists: `device_manager/model.go:45-47` — `zone_id` is a **soft** reference
(no `Ref`, so an empty value is not a foreign-key violation) with `Options: &ZoneModel`, and the
screen fills its choices.

## APIs used

| Symbol | Exact | Where |
|---|---|---|
| `devicemanager.DeviceModel` | `var DeviceModel = model.Definition{Name: "device", …}` | `github.com/veltylabs/device_manager` v0.2.86 |
| `model.Field.Options` | `Options *Definition` — the table a select takes its choices from | `webtyp.com/model` (as in `device_manager/model.go:47`) |
| `input.Select` | `func Select(...fmt.KeyValue) input.Input` | `webtyp.com/input` (as in `device_manager/model.go:47`) |

## Tools

```
go get github.com/veltylabs/device_manager@v0.2.86
```

## Stage 1 — the field (`model.go`, then `ormc`)

In `EquipmentModel`, after `name`, add exactly:

```go
// Referencia blanda (sin Ref): un equipo sin red (un gastroscopio) deja device_id vacío,
// y eso no debe ser una violación de clave foránea. La pantalla llena sus opciones con los
// equipos de device_manager.
{Name: "device_id", Type: input.Select(), Label: "Device", OmitEmpty: true, Options: &devicemanager.DeviceModel},
```

Import `devicemanager "github.com/veltylabs/device_manager"`. Regenerate `model_orm.go` with the repo's
`ormc` (v0.1.32 or later). `Equipment` gets `DeviceId string`.

## Stage 2 — translations (`lang.json`)

`langc sync .`; the key `Device` must have the same values `device_manager/lang.json` gives it
(es «Equipo», zh «设备», hi «डिवाइस», ar «جهاز», pt «Dispositivo», fr «Appareil», de «Gerät», ru
«Устройство»). `langc sync .` → nothing pending, no reuse warnings, no identical values.

## Stage 3 — tests (`tests/`)

New `tests/equipment_device_test.go`:

- `TestEquipment_DeviceIdOptional`: save an equipment with no `DeviceId` through `m.SaveEquipment` → no
  error; read it back → `DeviceId == ""`.
- `TestEquipment_DeviceIdKept`: save one with `DeviceId: "dev-1"` → read back → `"dev-1"`.
- `TestEquipment_DeviceOptionsFromDeviceManager`: the `device_id` field of `roomlayout.EquipmentModel`
  has `Options == &devicemanager.DeviceModel`.

Use the same setup the existing equipment tests use (`tests/ops_exec_test.go` / `newTestEnv`).

## Stage 4 — docs

`README.md` (Spanish): «Un equipamiento puede apuntar a un equipo registrado en `device_manager`
(`device_id`, opcional): un aparato se registra una vez. El que no tiene red deja el campo vacío.»

## Delivery check

```
GOPRIVATE=webtyp.com/*,github.com/veltylabs/* GOFLAGS=-mod=mod gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only: `model.go`, `model_orm.go`, `lang.json`,
  `tests/equipment_device_test.go`, `README.md`, `go.mod`, `go.sum`, this plan. No stray scripts.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `model.go`, `model_orm.go`, `go.mod`, `go.sum` | compiles |
| 2 | `lang.json` | `langc sync .` clean |
| 3 | `tests/equipment_device_test.go` | suite green |
| 4 | `README.md` | text present |
