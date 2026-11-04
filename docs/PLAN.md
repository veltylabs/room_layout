---
PLAN: "feat!: room days from the installation zone (Deps.Zone instead of a zone name)"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — room_layout: dates and hours from an explicit `time.Zone`

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/time@v0.6.3 webtyp.com/appkit@v0.2.7
```

Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, `go mod tidy` after `go get`.

## Why

`webtyp.com/time` v0.6.3 removed every implicit local time (`FormatDate`, `FormatTime`, `FormatDateTime`, `IsToday`,
`LocalMinutesToUnixUTC`) and renamed what is UTC (`Date` → `DateUTC`, `CivilDate` → `CivilUTC`, `Weekday` →
`WeekdayUTC`, `ParseDate` → `ParseDateUTC`). `webtyp.com/appkit` v0.2.7 hands every module the installation's zone,
already loaded: `appkit.Deps.Zone` on the server and `appkit.BrowserDeps{Caller, Zone}` to the screens
(`ModuleBrowser.Screens func(d appkit.BrowserDeps) []view.Screen`). This repo stops compiling until it adopts them.

## The rule (apply it at every site below)

- A **date key** — a day stored as its midnight UTC, with no hour (a holiday, a shift's specific date, a calendar day
  filter) — uses the UTC functions: `time.ParseDateUTC`, `time.WeekdayUTC`, `time.CivilUTC`.
- An **instant a person reads** (when something happened, a booking's hour, "today") uses the installation's zone:
  `zone.Date(u)`, `zone.Clock(u)`, `zone.DateTime(u)`, `zone.Minutes(u)`, `zone.Weekday(u)`, `zone.IsToday(u)`.
- Never call `time.LoadZone` in a module: the zone arrives in `Deps.Zone` / `BrowserDeps.Zone`.

## APIs used (time v0.6.3 `zone.go`, `api.go`; appkit v0.2.7)

| Symbol | Signature |
|---|---|
| `time.Zone` | `Date(u Unix) string` "YYYY-MM-DD"; `Clock(u Unix) string` "HH:MM"; `DateTime(u Unix) string`; `Minutes(u Unix) int`; `Weekday(u Unix) int` (0 = Sunday); `IsToday(u Unix) bool`; `At(date string, minutes int) (Unix, error)`; `Day(date string) (from, to Unix, err error)` |
| `time.ParseDateUTC`, `time.CivilUTC`, `time.WeekdayUTC`, `time.DateUTC` | the old `ParseDate`, `CivilDate`, `Weekday`, `Date`, renamed, same behaviour |
| `time.LoadZone` (tests and demos only) | `func LoadZone(name string) (Zone, error)` |
| `time.MachineZone` | `func MachineZone() Zone` — this computer's zone; only where stated below |
| `appkit.Deps.Zone` | `time.Zone` (server) |
| `appkit.BrowserDeps` | `struct { Caller router.Caller; Zone time.Zone }` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1 — the sites

- This module's `Deps.Timezone string` → `Deps.Zone time.Zone` (Spanish comment: "la zona de la instalación,
  appkit.Deps.Zone"); `module/server.go:29` passes `Zone: d.Zone`.
- `day.go:13-26` `localAt` becomes:
  ```go
  // localAt: el día local (como clave: su medianoche UTC) y el minuto de ese día, en la zona de la instalación.
  func (m *Module) localAt(sec time.Unix) (time.Unix, int) {
  	D, _ := time.ParseDateUTC(m.deps.Zone.Date(sec))
  	return D, m.deps.Zone.Minutes(sec)
  }
  ```
- `day.go:110`, `shift.go:67,204,208,353,441` `time.Weekday(<date key>)` → `time.WeekdayUTC(...)` (they are date keys).
- `convert.go:12` `time.ParseDate(` → `time.ParseDateUTC(`; `convert.go:23` `time.FormatDate(sec)` of a **date key** →
  `y, mo, d, _, _, _ := time.CivilUTC(sec); return fmt.Sprintf("%04d-%02d-%02d", y, mo, d)` (FormatDate shifted it by
  the machine offset: a negative offset showed the previous day).
- `module/browser.go:14` → `Screens: func(d appkit.BrowserDeps) []view.Screen { return ui.Screens(d.Caller) }`.

## Stage 2 — tests, seed and demo

Every test, seed and `web/` demo that builds what changed passes a zone: `zone, _ := time.LoadZone("America/Santiago")`.
New test: `localAt(2019-07-16 01:30 UTC)` with Santiago → day key of `2019-07-15`, minute `1290`.

## Delivery check

1. `gotest` passes (a PR that does not compile, or whose tests could not run, is a failed delivery).
2. `grep -rn -E "time\.(FormatDate|FormatTime|FormatDateTime|IsToday|LocalMinutesToUnixUTC|CivilDate|ParseDate|Weekday)\(" --include=*.go .` → empty
   (the `*UTC` names do not match).
3. Comments and docs of this repo stay **in Spanish** (veltylabs rule); add one line to `docs/ARCHITECTURE.md`:
   "Las fechas que ve una persona salen de la zona de la instalación (`Deps.Zone` / `BrowserDeps.Zone`); un día
   guardado como clave usa las funciones UTC."

Commit after every stage. Never edit a generated `*_orm.go`. Never run `git reset`, `git restore`, `git checkout --
<file>` or `git stash` on your own work. Stuck 30 minutes: open the PR anyway listing what fails.
