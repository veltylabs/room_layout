# room_layout
<img src="docs/img/badges.svg">

Módulo genérico para la distribución y ocupación de espacios de atención por niveles, equipamiento, turnos y disponibilidad para el ecosistema Velty.

## Inicio Rápido

```go
import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

// Instanciación de la base de datos y módulo
db := orm.New(mem.New())
ids, _ := unixid.NewUnixID()

mod, err := roomlayout.New(db, roomlayout.Deps{
	IDs:        ids,
	Categories: categoryReader,
	Occupants:  occupantReader,
	TenantID:   "tenant-1",
	Timezone:   "America/Santiago",
})
```

## Demo

Para ejecutar la demo web interactiva:

```bash
webtyp
```

ejecutado desde la raíz del repositorio.

## Operaciones Registradas (`MountOperations`)
El módulo expone operaciones estándar usando `view.MountSave`/`view.MountDelete`. Los errores de decodificación y de negocio son propagados mediante `rpc.WriteError`. Los permisos utilizan asignación por tabla con `model.ResourceOf`.

| Operación | Recurso | Acción |
|---|---|---|
| `list_floors` | `floor` | `Read` |
| `save_floor` | `floor` | `model.Create|model.Update` |
| `delete_floor` | `floor` | `Delete` |
| `list_rooms` | `room` | `Read` |
| `get_room` | `room` | `Read` |
| `save_room` | `room` | `model.Create|model.Update` |
| `deactivate_room` | `room` | `Update` |
| `list_equipment` | `equipment` | `Read` |
| `save_equipment` | `equipment` | `model.Create|model.Update` |
| `delete_equipment` | `equipment` | `Delete` |
| `list_room_equipment` | `room` | `Read` |
| `set_room_equipment` | `room` | `Update` |
| `list_categories` | `room` | `Read` |
| `list_room_categories` | `room` | `Read` |
| `set_room_categories` | `room` | `Update` |
| `list_occupants` | `room_shift` | `Read` |
| `list_room_shifts` | `room_shift` | `Read` |
| `save_room_shift` | `room_shift` | `model.Create|model.Update` |
| `delete_room_shift` | `room_shift` | `Delete` |
| `cancel_shift_occurrence` | `room_shift` | `Update` |
| `move_shift_occurrence` | `room_shift` | `model.Create|model.Update` |
| `list_room_day` | `room_shift` | `Read` |
| `list_board` | `room` | `Read` |
| `find_free_rooms` | `room` | `Read` |
| `list_public_availability` | — | `Public` |

## Lector para otros módulos

`RoomOfOccupant(tenantID, occupantID, at)`: el box donde un profesional tiene turno en ese instante (segundos Unix). Detalle en
`docs/ARCHITECTURE.md` §5b.

## Archivos Clave

- `model.go`: Definición de modelos, parámetros y respuestas.
- `module.go`: Constructor `New` e interfaces de dependencias.
- `floor.go`, `room.go`, `equipment.go`, `category.go`, `shift.go`, `day.go`: Lógica de dominio.
- `ops.go`: Registro de operaciones `router.OperationModule`.
- `view.go`: Vistas y presentadores.
- `ui/`: Pantalla del módulo en el navegador.
- `seed/`: Cargador de datos iniciales.
- `migrate/`: Migración de esquema.
