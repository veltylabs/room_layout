# room_layout
<img src="docs/img/badges.svg">

Módulo de distribución arquitectónica y ocupación de espacios de atención (pisos, salas, equipamiento, turnos y disponibilidad) construido sobre [`crud`](https://github.com/webtyp/crud). Más detalles en [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## ¿Por qué existe?

Un centro clínico u operativo necesita saber qué espacios tiene, cómo están equipados y qué profesional ocupa cada box en cada momento.

- **Sin esto**: Cada aplicación implementa a mano la persistencia, validaciones de solapamiento horario y cálculo de disponibilidad, duplicando reglas y arriesgando dobles reservas.
- **Con esto**: La distribución física, el equipamiento, los turnos semanales y fechados, y la disponibilidad pública y privada se gestionan mediante contratos declarativos con verificación estricta de solapes y aislamiento multi-inquilino.

## Antes de comenzar

- Go 1.26+ y `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`.
- Un `*orm.DB` con las tablas de [migrate/migrate.go](migrate/migrate.go).
- `roomlayout.Deps`: `IDs` (generador de IDs), `TenantID` (identificador del inquilino), `Zone` (zona horaria del módulo, ej. `"America/Santiago"`), `Categories` (`CategoryReader`), `Occupants` (`OccupantReader`), y opcionalmente `Hours` (`HoursReader`), `Publisher` (`events.Publisher`) y `Clock` (`func() time.Time`).

## Guía paso a paso

### Paso 1: Inicialización y métodos de dominio
```go
package main

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

type mockCategories struct{}
func (mockCategories) CategoryOptions(_ string) ([]roomlayout.CategoryOption, error) {
	return []roomlayout.CategoryOption{{Id: "cat1", Label: "Medicina General"}}, nil
}
type mockOccupants struct{}
func (mockOccupants) OccupantOptions(_ string) ([]roomlayout.OccupantOption, error) {
	return []roomlayout.OccupantOption{{Id: "doc1", Label: "Dr. Soto"}}, nil
}

func main() {
	db := orm.New(mem.New())
	ids, _ := unixid.NewUnixID()
	mod, err := roomlayout.New(db, roomlayout.Deps{
		IDs: ids, TenantID: "tenant-1", Zone: "America/Santiago",
		Categories: mockCategories{}, Occupants: mockOccupants{},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	f, _ := mod.SaveFloor(roomlayout.Floor{TenantId: "tenant-1", Name: "Piso 1", Cols: 10, Rows: 10})
	r, _ := mod.SaveRoom(roomlayout.Room{TenantId: "tenant-1", FloorId: f.Id, Code: "B1", Name: "Box 1", IsActive: true})
	rooms, _ := mod.ListRooms("tenant-1", false)
	fmt.Println("Salas creadas:", len(rooms), r.Code)
}
```

### Paso 2: Mutaciones vía crud
Las peticiones desde el cliente ejecutan operaciones declarativas sobre el cable. Ver ejemplos en [tests/rules_test.go](tests/rules_test.go).

### Paso 3: Exploración interactiva
Ejecuta la demo con interfaz gráfica desde la raíz del módulo usando el cargador en [seed/demo.go](seed/demo.go):
```bash
webtyp
```

## Mapa de documentación

| Orden | Documento | Qué pregunta responde y cuándo leerlo |
|---|---|---|
| 1 | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | ¿Cómo interactúan las piezas, almacenes y flujos del sistema? |
| 2 | [docs/TESTS.md](docs/TESTS.md) | ¿Qué regla de negocio protege cada prueba del repositorio? |
| 3 | [docs/diagrams/ROOM_CATEGORY.md](docs/diagrams/ROOM_CATEGORY.md) | ¿Cómo valida y previene errores el vínculo entre sala y categoría? |
| 4 | [docs/diagrams/SHIFT_RULES.md](docs/diagrams/SHIFT_RULES.md) | ¿Cómo se validan los horarios, límites de apertura y solapamientos de turnos? |
| 5 | [docs/diagrams/DATABASE.md](docs/diagrams/DATABASE.md) | ¿Cuál es la estructura física de tablas e índices del módulo? |

### Mapeo de operaciones («Antes» vs. «Después»)

| Operación anterior | Operación `crud` equivalente |
|---|---|
| `list_floors`, `save_floor`, `delete_floor` | `room_floor` (`Create`, `Read`, `Update`, `Delete`) |
| `list_rooms`, `get_room`, `save_room`, `deactivate_room` | `room_space` (`Create`, `Read`, `Update`, `Delete`) |
| `list_equipment`, `save_equipment`, `delete_equipment` | `room_equipment` (`Create`, `Read`, `Update`, `Delete`) |
| `list_room_artifacts`, `save_room_artifact`, `delete_room_artifact` | `room_artifact` (`Create`, `Read`, `Update`, `Delete`) |
| `list_room_equipment`, `set_room_equipment` | `room_space_equipment` (`Create`, `Read`, `Delete`) |
| `list_room_categories`, `set_room_categories` | `room_category` (`Create`, `Read`, `Delete`) |
| `list_room_shifts`, `save_room_shift`, `delete_room_shift` | `room_shift` (`Create`, `Read`, `Update`, `Delete`) |
| `cancel_shift_occurrence`, `move_shift_occurrence` | `room_shift_cancellation` (`Create`, `Read`) |
| `list_categories` | `category_option` (`Read`) |
| `list_occupants` | `occupant_option` (`Read`) |
| `list_room_day` | `day_shift` (`Read`) |
| `list_board` | `board_room` (`Read`) |
| `find_free_rooms` | `free_room` (`Read`) |
| `list_public_availability` | `public_slot` (`Read`, público) |

### Permisos por entidad y modelos de lectura

| Entidad / Proyección | Acciones | Tipo |
|---|---|---|
| `room_floor` | `Create`, `Read`, `Update`, `Delete` | Tabla persistida |
| `room_space` | `Create`, `Read`, `Update`, `Delete` | Tabla persistida |
| `room_equipment` | `Create`, `Read`, `Update`, `Delete` | Tabla persistida |
| `room_artifact` | `Create`, `Read`, `Update`, `Delete` | Tabla persistida |
| `room_space_equipment` | `Create`, `Read`, `Delete` | Tabla persistida (clave compuesta) |
| `room_category` | `Create`, `Read`, `Delete` | Tabla persistida (clave compuesta) |
| `room_shift` | `Create`, `Read`, `Update`, `Delete` | Tabla persistida |
| `room_shift_cancellation` | `Create`, `Read` | Tabla persistida |
| `category_option` | `Read` | Proyección (lectura) |
| `occupant_option` | `Read` | Proyección (lectura) |
| `day_shift` | `Read` | Proyección (lectura) |
| `board_room` | `Read` | Proyección (lectura) |
| `free_room` | `Read` | Proyección (lectura) |
| `public_slot` | `Read` (público) | Proyección (lectura) |

*Nota*: `public_slot` está declarado como recurso público (`model.Read`), permitiendo consultas de disponibilidad sin sesión autenticada.
