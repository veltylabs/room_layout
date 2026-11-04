---
PLAN: "feat!: room_layout reads opening hours as time.DayLimits; contracts in interfaces.go"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — room_layout asks business_calendar for DayLimits and keeps its contracts in interfaces.go

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
go get webtyp.com/time@v0.6.4 webtyp.com/appkit@v0.2.8
```

## Why

`time` v0.6.4 renamed `DayBounds` → `DayLimits` (fields `Open bool`, `OpenMin int`, `CloseMin int`, unchanged) and
`Unbounded` → `Unlimited`. `business_calendar` renames `GetDayBounds` → `DayLimits`. Today `module/server.go` never
fills `Deps.Bounds`, so in an app room days ignore the clinic's hours.

## Design gate

1. Prior art: consumer-declared narrow interface (same as today).
2. Names: `HoursReader` with `DayLimits` = the method business_calendar offers.
3. Ledger: renames only.
4. Belongs here: unchanged.
5. Deletes: `BoundsReader`, `GetDayBounds` calls, `DayBounds`/`Unbounded`.

## Stage 1 — interfaces.go

Move `CategoryReader`, `OccupantReader` and `BoundsReader` from `module.go` (lines ~11–21) to a new `interfaces.go`.
Rename `BoundsReader` → `HoursReader`:
```go
// HoursReader es la parte de business_calendar que este módulo usa: si la clínica abre ese día y entre qué minutos.
// Ejemplo: lim, err := h.DayLimits(date)
type HoursReader interface {
	DayLimits(date time.Unix) (time.DayLimits, error)
}
```
`Deps.Bounds` → `Deps.Hours HoursReader`; every `m.deps.Bounds.GetDayBounds(x)` (`shift.go:129,430`, `day.go:91,292,372,472`)
→ `m.deps.Hours.DayLimits(x)`; `time.DayBounds` → `time.DayLimits`, `time.Unbounded` → `time.Unlimited` everywhere.
Keep the existing nil checks on the dependency unchanged (renamed).

## Stage 2 — module/server.go

After `occ`: `hours, _ := appkit.Find[roomlayout.HoursReader](d)` (optional: without business_calendar the whole day
is open, as today) and pass `Hours: hours` in `roomlayout.Deps`.

## Stage 3 — tests

`tests/setup_test.go` `fakeBounds.GetDayBounds` → `fakeHours.DayLimits`; all tests pass.

## Stage 4 — docs (Spanish)

`README.md`, `docs/ARCHITECTURE.md`: `HoursReader` (`DayLimits`, from business_calendar); contracts in `interfaces.go`.

## Delivery check

1. `gotest` passes.
2. `grep -rn "DayBounds\|Unbounded\|BoundsReader\|\.Bounds\b" --include=*.go --include=*.md .` → empty.
3. `grep -n "interface {" module.go` → empty.

Commit after every stage. Never edit `model_orm.go` by hand; a build error there means the hand-written code is wrong.
Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work. Stuck 30 minutes: open
the PR anyway listing what fails. A PR that does not compile, or whose tests could not run, is a failed delivery.
Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, `go mod tidy` before `ormc`.
If a signature in the published tag differs from the APIs table, stop and name the symbol in the PR; never adapt it by guessing.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `interfaces.go`, `module.go`, `shift.go`, `day.go` | compiles |
| 2 | `module/server.go` | compiles |
| 3 | `tests/*` | tests pass |
| 4 | `README.md`, `docs/ARCHITECTURE.md` | done |
