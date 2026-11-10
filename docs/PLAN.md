---
PLAN: "refactor!: room_layout on the crud-by-declaration framework, plan 1 of 2 — floors, rooms, equipment, categories, artifacts and every read; shift writes stay as methods"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>`, `git checkout HEAD <path>` or `git stash` on your own work: a previous session lost its whole migration with a `git checkout HEAD tests`.
> **Publishing:** you publish only by OPENING THE PULL REQUEST (a `git push` from your sandbox does not reach GitHub, and work not submitted is lost when the session ends). Open the PR after Stage 3 with whatever compiles and keep committing to it.
> **Never ask a yes/no question about progress** («am I on the right track?», «shall I proceed?», «does this approach sound correct?»): the answer is always yes. Ask only when a compile error or a signature in this plan is wrong, and name the symbol.
> Never edit `definition_gen.go`. Never write `go.mod`/`go.sum` lines by hand: only `go get` and `go mod tidy` do. No scratch files and no sed/regex/python scripts over the source: edit by hand, file by file, and run `gotest` after each file.

# Plan 1 of 2: `room_layout` on the crud-by-declaration framework

Two earlier sessions of the single plan ended with an empty or gutted pull request (the last one carried a stray `commit4.sh`, plan notes, and deleted every test). The module is the biggest of the wave (8 tables, 3 composite keys, 6 read models, the shift logic, the map UI), so the work is cut in **two plans, each ending with a tree that compiles and a green `gotest`**:

- **This plan (1)** declares EVERY table and read model in `Definition`, and migrates everything except the shift writers: floors, rooms, artifacts, equipment, the room↔equipment and room↔category links, the option lists, and the reads of shifts, day, board, free rooms and public availability.
- **Plan 2 (parked, written, dispatched after this one is closed)** turns the shift writers on (`RoomShift` and `RoomShiftCancellation` create/update/delete with rules, `SaveShift`, `DeleteShift`, `CancelShiftOccurrence`, `MoveShiftOccurrence` on the store) and ports the shift screen.

**What does NOT change in this plan, on purpose:** `shift.go` and `day.go` stay **exactly as they are** (same bodies, on `m.db`). `RoomShift` and `RoomShiftCancellation` are `Allow: model.Read` here, so over the wire the shift writes (save, delete, cancel, move) are not offered until plan 2; the screen of shifts of a room lists and shows them and answers `ErrShiftWritesPending` when a person saves or deletes one. The methods `SaveShift`, `DeleteShift`, `CancelShiftOccurrence`, `MoveShiftOccurrence` keep working for any code that calls them directly. `m.db` therefore stays in `Module` until plan 2.

After: tables `Floor`, `Room`, `RoomArtifact`, `Equipment` (full crud, rules where listed), `RoomEquipment`, `RoomCategory` (composite keys; create/read/delete), `RoomShift` and `RoomShiftCancellation` (read only); read models `category_option`, `occupant_option`, `day_shift`, `board_room`, `free_room`, `public_slot`. Every exported method keeps its name and signature.

Reference module, already migrated (copy its shape): https://github.com/veltylabs/patient_directory/blob/main — `definition.go`, `rule_patient_*.go`, `store.go`, `seed/demo.go`, `tests/setup_test.go`, `tests/crud_test.go`, `tests/demo_test.go`; and https://github.com/veltylabs/device_manager/blob/main (`view.go`, `seed/`).

## Stages

| # | File | What | Files touched |
|---|---|---|---|
| 1 | [PLAN_STAGE_1_TOOLS_DEFINITION.md](PLAN_STAGE_1_TOOLS_DEFINITION.md) | tools, `definition.go`, `errors` | `definition.go`, `go.mod`, `go.sum` |
| 2 | [PLAN_STAGE_2_MODULE_STORE.md](PLAN_STAGE_2_MODULE_STORE.md) | `Module.store`, `storeFor`, row helpers, translation table | `module.go`, `store.go`, `rows.go` |
| 3 | [PLAN_STAGE_3_RULES.md](PLAN_STAGE_3_RULES.md) | generate; shared checks; rule files | `checks.go`, `rule_*.go` |
| 4 | [PLAN_STAGE_4_METHODS.md](PLAN_STAGE_4_METHODS.md) | the kept methods of floors, rooms, equipment, categories, artifacts on the store; `ops.go` deleted | `floor.go`, `room.go`, `equipment.go`, `category.go`, `ops.go` (deleted) |
| 5 | [PLAN_STAGE_5_BROWSER.md](PLAN_STAGE_5_BROWSER.md) | `view.go`, `ui/data.go`, `module/`, `seed/demo.go` | `view.go`, `ui/data.go`, `module/*.go`, `seed/demo.go`, `web/` (deleted) |
| 6 | [PLAN_STAGE_6_TESTS_DOCS.md](PLAN_STAGE_6_TESTS_DOCS.md) | tests, `docs/TESTS.md`, README, architecture | `tests/*.go`, `docs/*`, `README.md` |

## Operation mapping (every operation of today has exactly one row)

`<entity>.<verb>` is `router.Operation{Module: ModelName, Entity: "<entity>", Action: model.<Verb>}`; `ModelName` is the constant `definitionc` writes (`"room_layout"`): never a literal.

| Today | After (plan 1) |
|---|---|
| `list_floors`, `save_floor`, `delete_floor` | `room_floor.read`, `.create` / `.update`, `.delete` |
| `list_rooms`, `get_room`, `save_room` | `room_space.read`, `.create` / `.update` |
| `deactivate_room` | `room_space.update` with `Fields: ["is_active"]` |
| `list_room_artifacts`, `save_room_artifact`, `delete_room_artifact` | `room_artifact.read` (filter `{room_id}`), `.create` / `.update`, `.delete` |
| `list_equipment`, `save_equipment`, `delete_equipment` | `room_equipment.read`, `.create` / `.update`, `.delete` |
| `list_room_equipment` / `set_room_equipment` | `room_space_equipment.read` (filter `{room_id}`) / `.create` for added and `.delete` (id `model.JoinKey(room, equipment)`) for removed |
| `list_room_categories` / `set_room_categories` | `room_category.read` (filter `{room_id}`) / `.create` and `.delete` (same) |
| `list_categories`, `list_occupants` | read models `category_option`, `occupant_option` |
| `list_room_shifts` | `room_shift.read` (filter `{room_id}`) |
| `list_room_day`, `list_board`, `find_free_rooms`, `list_public_availability` | read models `day_shift`, `board_room`, `free_room`, `public_slot` |
| `save_room_shift`, `delete_room_shift`, `cancel_shift_occurrence`, `move_shift_occurrence` | **not offered over the wire until plan 2** (the methods keep working) |

## Delivery check

1. `gotest` ends green on EVERY destination (`wasm` and `browser` included).
2. `langc sync .` reports zero untranslated keys for `es` and zero reuse warnings.
3. `grep -rn "reg.Operation\|view.MountSave\|view.MountDelete\|router.Context\|rpc.WriteError\|view.ListArgs\|OpListFloors\|loopback.WithTenant" --include=*.go . | grep -v "^./definition_gen.go"` → empty; `grep -n "m\.db" *.go | grep -v "^definition_gen.go\|^shift.go\|^day.go"` → only `New` and `storeFor`.
4. `ls *.py *.sh commit_message.txt 2>/dev/null` → empty; `grep -c XXXX go.sum` → 0; `test ! -e web`.
5. `git diff --stat main` lists the files of the stages table; `shift.go` and `day.go` do NOT appear (not even whitespace); `git diff --stat main --diff-filter=D` lists ONLY `ops.go`, `web/browser/main.go` and the test files Stage 6 names for deletion — every other test is kept and fixed, never deleted.
