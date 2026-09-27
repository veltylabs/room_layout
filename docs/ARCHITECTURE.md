# Arquitectura del módulo `room_layout`

Módulo genérico para la distribución y ocupación de espacios de atención por niveles, equipamiento, turnos y disponibilidad.

---

## 1. Alcance y Vocabulario del Dominio

El módulo gestiona espacios físicos en establecimientos (p. ej. consultorios, spas, hoteles) organizados en niveles/pisos. Cada espacio puede ser habilitado para ciertas categorías de atención y equipado con infraestructura. La ocupación de los espacios se asigna a ocupantes mediante turnos (semanales o fechados).

**Vocabulario del dominio:**

| Concepto del negocio (consultorio) | Nombre en el dominio | Tabla |
|---|---|---|
| Piso | `floor` | `floor` |
| Box / Sala | `room` | `room` |
| Infraestructura / Equipamiento | `equipment` | `equipment`, `room_equipment` |
| Área / Especialidad | `category` (id blando) | `room_category` |
| Profesional / Doctor | `occupant` (id blando) | — |
| Turno de box | `room_shift` | `room_shift` |
| Cancelación de ocurrencia | `room_shift_cancellation` | `room_shift_cancellation` |

---

## 2. Entidades Principales

1. **`Floor` (`floor`)**: Representa un piso o nivel físico (`id`, `tenant_id`, `name`, `position`, `updated_at`).
2. **`Room` (`room`)**: Espacio de atención dentro de un nivel (`id`, `tenant_id`, `floor_id`, `code`, `name`, `notes`, `is_active`, `updated_at`).
3. **`Equipment` (`equipment`)**: Equipamiento o infraestructura disponible (`id`, `tenant_id`, `name`, `updated_at`).
4. **`RoomEquipment` (`room_equipment`)**: Relación N:M entre un espacio y su equipamiento (`tenant_id`, `room_id`, `equipment_id`).
5. **`RoomCategory` (`room_category`)**: Relación N:M entre un espacio y las categorías para las que está habilitado (`tenant_id`, `room_id`, `category_id`).
6. **`RoomShift` (`room_shift`)**: Turno de ocupación recurrente (semanal por `day_of_week`) o fechado (`specific_date`).
7. **`RoomShiftCancellation` (`room_shift_cancellation`)**: Excepción que invalida la ocurrencia de un turno semanal en una fecha específica.

---

## 3. Reglas de Negocio

- **Validación de solapamiento**: Dos turnos en el mismo espacio (o para el mismo ocupante) no pueden solaparse en sus franjas horarias `[start_min, end_min)`.
- **Límites de establecimiento (`DayBounds`)**: Los turnos fechados deben encontrarse dentro del horario de apertura del día.
- **Transacciones e Integridad**: Los cambios en categorías y equipamiento de un espacio se realizan de forma atómica en transacciones. No se puede deshabilitar una categoría si el espacio tiene turnos activos vinculados a ella.
- **Aislamiento Multi-Tenant**: Toda consulta y mutación filtra estrictamente por `tenant_id`.

---

## 4. Tabla de Operaciones (Ops)

| Op | Recurso | Acción | Descripción |
|---|---|---|---|
| `list_floors` | `floor` | `model.Read` | Lista los niveles del tenant |
| `save_floor` | `floor` | `model.Create\|model.Update` | Crea o actualiza un nivel |
| `delete_floor` | `floor` | `model.Delete` | Borra un nivel (si no tiene espacios) |
| `list_rooms` | `room` | `model.Read` | Lista espacios del tenant |
| `get_room` | `room` | `model.Read` | Obtiene un espacio por ID |
| `save_room` | `room` | `model.Create\|model.Update` | Crea o actualiza un espacio |
| `deactivate_room` | `room` | `model.Update` | Desactiva un espacio |
| `list_equipment` | `equipment` | `model.Read` | Lista equipamiento disponible |
| `save_equipment` | `equipment` | `model.Create\|model.Update` | Crea o actualiza equipamiento |
| `delete_equipment` | `equipment` | `model.Delete` | Borra equipamiento |
| `list_room_equipment` | `room` | `model.Read` | Lista equipamiento asignado a un espacio |
| `set_room_equipment` | `room` | `model.Update` | Reemplaza el equipamiento de un espacio |
| `list_categories` | `room` | `model.Read` | Lista categorías globales desde `CategoryReader` |
| `list_room_categories` | `room` | `model.Read` | Lista categorías asignadas a un espacio |
| `set_room_categories` | `room` | `model.Update` | Reemplaza categorías asociadas a un espacio |
| `list_occupants` | `room_shift` | `model.Read` | Lista ocupantes desde `OccupantReader` |
| `list_room_shifts` | `room_shift` | `model.Read` | Lista turnos activos de un espacio |
| `save_room_shift` | `room_shift` | `model.Create\|model.Update` | Crea o actualiza un turno |
| `delete_room_shift` | `room_shift` | `model.Delete` | Borra un turno |
| `cancel_shift_occurrence` | `room_shift` | `model.Update` | Cancela una ocurrencia de turno semanal |
| `move_shift_occurrence` | `room_shift` | `model.Create\|model.Update` | Mueve una ocurrencia a otro espacio |
| `list_room_day` | `room_shift` | `model.Read` | Ocupación resuelta de un día |
| `list_board` | `room` | `model.Read` | Estado actual/siguiente del tablero de espacios |
| `find_free_rooms` | `room` | `model.Read` | Busca espacios libres en un rango horario |
| `list_public_availability` | — | `.Public()` | Consulta pública de slots disponibles |

---

## 5. Ejemplo de Raíz de Composición

```go
package main

import (
	roomlayout "github.com/veltylabs/room_layout"
	"webtyp.com/fmt"
	"webtyp.com/orm"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
)

type categoryAdapter struct{}

func (c *categoryAdapter) CategoryOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{
		{Key: "general", Value: "Medicina General"},
		{Key: "cardio", Value: "Cardiología"},
	}, nil
}

type occupantAdapter struct{}

func (o *occupantAdapter) OccupantOptions(tenantID string) ([]fmt.KeyValue, error) {
	return []fmt.KeyValue{
		{Key: "perez", Value: "Dr. Juan Pérez"},
	}, nil
}

func main() {
	db := orm.New(mem.New())
	ids, _ := unixid.NewUnixID()

	mod, err := roomlayout.New(db, roomlayout.Deps{
		IDs:        ids,
		Categories: &categoryAdapter{},
		Occupants:  &occupantAdapter{},
		TenantID:   "tenant-1",
		Timezone:   "America/Santiago",
	})
	if err != nil {
		panic(err)
	}

	router := loopback.WithTenant("tenant-1", mod)
	_ = router
}
```
