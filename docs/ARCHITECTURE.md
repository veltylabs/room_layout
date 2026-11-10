# Arquitectura del módulo `room_layout`

Módulo genérico para la distribución arquitectónica y ocupación de espacios de atención (pisos, salas, equipamiento, turnos y disponibilidad) construido sobre el framework declarativo `crud`.

---

## 1. Visión general en un minuto

`room_layout` administra la infraestructura física de un establecimiento (clínica, centro médico, hotel, espacio compartido) y cómo se ocupa en el tiempo. Organiza el edificio en niveles o pisos (`room_floor`), ubica en ellos salas de atención (`room_space`) y artefactos en el plano (`room_artifact`), y gestiona el catálogo de equipamiento (`room_equipment`) y su asignación a salas (`room_space_equipment`).

La ocupación se modela mediante turnos (`room_shift`), ya sean semanales recurrentes o de fecha específica, vinculados a categorías de atención provistas por la aplicación (`room_category`) y ocupantes profesionales. Las excepciones puntuales (feriados, reasignaciones) se registran como cancelaciones de ocurrencia (`room_shift_cancellation`). Sobre este núcleo se construyen proyecciones de disponibilidad inmediata (`board_room`), búsqueda de salas libres (`free_room`), agenda diaria (`day_shift`) y disponibilidad pública (`public_slot`).

---

## 2. Piezas y estructura

| Archivo / Carpeta | Rol en el módulo |
|---|---|
| `definition.go` | Declaración central de entidades, modelos de lectura, permisos (`Allow`), reglas (`Rules`) y constantes de error. |
| `definition_gen.go` | Código generado por `definitionc` (estructuras, códecs, helpers de consulta y enrutador `crud`). |
| `checks.go` | Funciones de validación compartidas entre reglas `crud` y métodos de dominio (existencia de sala, unicidad de código, categorías válidas). |
| `shift_tx.go` | Lógica transaccional de turnos: verificación de franjas horarias, apertura del recinto, solapamientos y traslados. |
| `store.go` | Métodos `Store*()` para exponer repositorios al framework y constructor multi-inquilino `storeFor`. |
| `rows.go` | Ayudantes tipados de lectura de una o múltiples filas (`firstRow`, `allRows`, `keyOf`, `notFound`). |
| `order.go` | Ordenamiento determinista en memoria para pisos, salas y equipamiento sin dependencias de stdlib `sort`. |
| `floor.go` | Métodos de Go para administración de pisos (`SaveFloor`, `ListFloors`, `DeleteFloor`). |
| `room.go` | Métodos de Go para administración de salas (`SaveRoom`, `GetRoom`, `ListRooms`, `DeactivateRoom`). |
| `equipment.go` | Métodos de Go para equipamiento y artefactos (`SaveEquipment`, `SetRoomEquipment`, `SaveRoomArtifact`, etc.). |
| `category.go` | Métodos de Go para vinculación sala-categoría (`ListRoomCategories`, `SetRoomCategories`). |
| `shift.go` | Métodos de Go para turnos y traslados (`SaveShift`, `ListRoomShifts`, `DeleteShift`, `CancelShiftOccurrence`, `MoveShiftOccurrence`). |
| `day.go` | Lógica de consulta de ocupación diaria, tablero de estado y salas libres (`shiftsOn`, `FindFreeRooms`, etc.). |
| `view.go` | Presentadores de lista para la interfaz (`NewFloorView`, `NewRoomView`, `NewEquipmentView`, `NewShiftView`) y adaptador `roomShiftLister`. |
| `ui/` | Componentes visuales en el navegador: plano interactivo del recinto, mazo de pestañas y formularios. |
| `seed/` | Cargador de datos iniciales y punto de entrada para desarrollo interactivo (`seed/demo.go`). |
| `migrate/` | DDL y creación de tablas en la base de datos relacional. |

---

## 3. Flujos paso a paso

### 3.1 Crear una sala (`room_space.create`)
1. El cliente envía `&crud.CreateArgs{Records: []model.Model{&Room{FloorId, Code, Name}}}`.
2. La regla `Room.Create` verifica mediante `checkFloorFree` que el piso exista en el tenant.
3. Se verifica con `checkCodeFree` que el código de sala no esté repetido en el tenant.
4. Si las validaciones pasan, `crud` persiste la fila generando el identificador y estampa `updated_at`.
5. Se publica el evento de cambio correspondiente.

### 3.2 Vincular una categoría a una sala (`room_category.create`)
1. El cliente solicita asociar una sala con una categoría mediante clave compuesta `(RoomId, CategoryId)`.
2. La regla `RoomCategory.Create` valida que la sala exista y que `CategoryId` sea una de las entregadas por `Deps.Categories`.
3. Ver detalles de ramificaciones y rechazos en [docs/diagrams/ROOM_CATEGORY.md](diagrams/ROOM_CATEGORY.md).
4. La eliminación (`room_category.delete`) comprueba previamente que la sala no tenga turnos activos que dependan de esa categoría.

### 3.3 Leer el tablero de salas (`board_room.read`)
1. La consulta solicita la proyección `board_room` para la fecha actual.
2. El lector recopila las salas activas, sus pisos y categorías habilitadas.
3. Para cada sala, evalúa los turnos del día calculando si la sala se encuentra libre, ocupada y quién es el siguiente profesional programado.

### 3.4 Guardar un turno (`room_shift.create` / `room_shift.update`)
1. La regla `RoomShift.Create` o `Update` valida la franja `[StartMin, EndMin)` dentro de 00:00 y 24:00.
2. Si el turno es fechado (`SpecificDate`), comprueba con `Deps.Hours` que se ajuste al horario de apertura.
3. Comprueba que la sala tenga asignada la categoría del turno.
4. Valida ausencia de solapamiento horario en la sala y para el profesional ocupante.
5. Ver flujo completo en [docs/diagrams/SHIFT_RULES.md](diagrams/SHIFT_RULES.md).

### 3.5 Mover una ocurrencia de turno
1. Si el turno es semanal: se registra una cancelación (`RoomShiftCancellation`) con el campo `moved_to_room_id`.
2. En la misma transacción atómica, se crea un turno fechado (`SpecificDate`) en la sala de destino.
3. Si el turno ya era fechado: se actualiza directamente su `RoomId` tras revalidar categoría y solapamiento en el destino.

---

## 4. Contratos y límites de seguridad

1. **Aislamiento multi-inquilino estricto**: Cada mutación o lectura opera a través de `storeFor(tenantID)`. Un tenant vacío es rechazado (`ErrTenantRequired`), impidiendo lecturas globales o filtraciones entre organizaciones.
2. **Claves compuestas nativas**: Tablas de asociación como `room_space_equipment` y `room_category` emplean clave primaria compuesta `(room_id, equipment_id)` o `(room_id, category_id)`. Las peticiones de creación con clave existente son tratadas de forma idempotente por `crud`.
3. **Paridad de validaciones (`checks.go`)**: Los métodos directos de Go y las operaciones sobre el cable invocan las mismas funciones puras de verificación, evitando discrepancias entre la API interna y la externa.
4. **Atomicidad en reubicaciones**: El traslado de ocurrencias combina la cancelación y la inserción del nuevo turno en una sola transacción (`tx.Atomic`). Si la sala de destino no soporta la categoría o presenta solapamiento, la operación se cancela en su totalidad sin efectos secundarios.
