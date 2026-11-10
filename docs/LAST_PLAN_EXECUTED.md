---
PLAN: "refactor!: room_layout on the crud-by-declaration framework — full migration (floors, rooms, equipment, categories, artifacts, shifts, cancellations, read models and screens)"
EXECUTOR: local
REVIEWER: cesar
---

> Read AGENTS.md at the repo root before any change.

# Plan: `room_layout` on the crud-by-declaration framework (Full Migration)

Combines Plan 1 (floors, rooms, equipment, categories, artifacts and read models) and Plan 2 (shift writers, `RoomShift` and `RoomShiftCancellation` rules, store methods and shift screen).

## Stages

| # | What | Files touched |
|---|---|---|
| 1 | tools, `definition.go`, errors | `definition.go`, `go.mod`, `go.sum` |
| 2 | `Module.store`, `storeFor`, row helpers, order, conversion and shift tx helpers | `module.go`, `store.go`, `rows.go`, `order.go`, `convert.go`, `shift_tx.go` |
| 3 | generate; shared checks; rule files (floors, rooms, equipment, categories, shifts, cancellations, read models) | `checks.go`, `rule_*.go` |
| 4 | store methods on floors, rooms, equipment, categories, shifts, day; `ops.go` deleted | `floor.go`, `room.go`, `equipment.go`, `category.go`, `shift.go`, `day.go`, `ops.go` (deleted) |
| 5 | `view.go` (all presenters including shifts), `ui/data.go`, `module/`, `seed/demo.go`, `web/` deleted | `view.go`, `ui/data.go`, `module/*.go`, `seed/demo.go`, `web/` (deleted) |
| 6 | tests (`tests/rules_test.go`, `tests/shift_rules_test.go`, porting existing), `docs/TESTS.md`, README, architecture, diagrams | `tests/*.go`, `docs/*`, `README.md` |

## Delivery check

1. `gotest` ends green on EVERY destination (`wasm` and `browser` included).
2. `langc sync .` reports zero untranslated keys for `es` and zero reuse warnings.
3. `grep -rn "reg.Operation\|view.MountSave\|view.MountDelete\|router.Context\|rpc.WriteError\|view.ListArgs\|OpListFloors\|loopback.WithTenant\|ErrShiftWritesPending" --include=*.go . | grep -v "^./definition_gen.go"` → empty; `grep -n "m\.db" *.go | grep -v "^definition_gen.go"` → only `New` and `storeFor`.
4. `ls *.py *.sh commit_message.txt 2>/dev/null` → empty; `grep -c XXXX go.sum` → 0; `test ! -e web`.
5. `git diff --stat main --diff-filter=D` lists ONLY `ops.go`, `web/browser/main.go` and the test files Stage 6 names for deletion.
