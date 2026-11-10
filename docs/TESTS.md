# Pruebas

| Test | Rule |
|---|---|
| [TestFloor_CreateThroughCrud](../tests/rules_test.go) | Creación de piso vía crud con limpieza de espacios en el nombre y asignación de tenant. |
| [TestFloor_BlankNameRejected](../tests/rules_test.go) | Rechazo con ErrorInvalid en campo name si el nombre del piso está en blanco. |
| [TestFloor_DeleteWithRoomsRefused](../tests/rules_test.go) | Rechazo con ErrorConflict al intentar borrar un piso que contiene salas activas. |
| [TestRoom_CodeIsUniquePerTenant](../tests/rules_test.go) | Rechazo con ErrorConflict al intentar crear o actualizar una sala con código duplicado en el tenant. |
| [TestRoom_DeactivateIsAnUpdate](../tests/rules_test.go) | Desactivar una sala mediante actualización del campo is_active a false. |
| [TestEquipment_DeleteRemovesItsLinks](../tests/rules_test.go) | Eliminar un equipamiento desvincula automáticamente sus enlaces con salas. |
| [TestRoomEquipment_CompositeKey](../tests/rules_test.go) | Enlace sala-equipamiento con clave compuesta; creación idempotente y validación de existencia de sala. |
| [TestRoomCategory_UnknownCategoryRejected](../tests/rules_test.go) | Rechazo con ErrorInvalid en category_id si la categoría no es entregada por la app. |
| [TestRoomCategory_DeleteInUseRefused](../tests/rules_test.go) | Rechazo con ErrorConflict al borrar categoría vinculada a una sala con turnos activos. |
| [TestRoomShift_ReadByRoom](../tests/rules_test.go) | Lectura de turnos filtrados por room_id retornando sólo los de la sala solicitada. |
| [TestReadModel_Options](../tests/rules_test.go) | Los modelos de lectura category_option y occupant_option entregan las opciones provistas por las interfaces. |
| [TestReadModel_DayShift](../tests/rules_test.go) | Lectura de day_shift entrega los turnos programados para una fecha específica. |
| [TestReadModel_FreeRoomNeedsItsFilter](../tests/rules_test.go) | Lectura de free_room exige filtro de fecha y rango horario válido. |
| [TestReadModel_PublicSlotNeedsDate](../tests/rules_test.go) | Lectura de public_slot exige filtro de fecha no vacío. |
| [TestTenant_CrudSeesOnlyItsTenant](../tests/rules_test.go) | Aislamiento multi-inquilino en crud; operaciones de un tenant no ven datos de otro. |
| [TestMount_MatchesDefinition](../tests/rules_test.go) | Validación de que lo montado en crud coincide con model.Definition. |
| [TestCreate_PublishesChange](../tests/rules_test.go) | Creación vía crud publica evento de cambio con el registro creado. |
| [TestShift_CreateThroughCrud](../tests/shift_rules_test.go) | Creación de turno vía crud con conversión de datos y validaciones de rango. |
| [TestShift_RulesOfTheRange](../tests/shift_rules_test.go) | Validación de límites de inicio y término dentro de las 24 horas y orden temporal. |
| [TestShift_CategoryRules](../tests/shift_rules_test.go) | Verificación de que la sala tenga habilitada la categoría del turno asignado. |
| [TestShift_OccupantRules](../tests/shift_rules_test.go) | Exige ocupante válido mediante identificador o etiqueta textual. |
| [TestShift_OverlapRejected](../tests/shift_rules_test.go) | Rechazo de turnos solapados en horario para la misma sala o el mismo ocupante. |
| [TestShift_DatedMustFitOpeningHours](../tests/shift_rules_test.go) | Turno con fecha específica debe estar dentro del horario de apertura del establecimiento. |
| [TestShift_UpdateMergesFields](../tests/shift_rules_test.go) | Actualización de turno modifica únicamente los campos enviados y preserva el resto. |
| [TestShift_UpdateRoomChecksTheTarget](../tests/shift_rules_test.go) | Cambiar de sala un turno valida que la sala destino soporte la categoría. |
| [TestShift_DeleteRemovesItsCancellations](../tests/shift_rules_test.go) | Eliminar un turno elimina en cascada las cancelaciones de sus ocurrencias. |
| [TestCancellation_WeeklyOnly](../tests/shift_rules_test.go) | Solo turnos semanales permiten cancelar ocurrencias en fechas puntuales. |
| [TestCancellation_WrongWeekday](../tests/shift_rules_test.go) | Cancelación rechazada si la fecha no corresponde al día de la semana del turno. |
| [TestCancellation_Idempotent](../tests/shift_rules_test.go) | Cancelar la misma ocurrencia por segunda vez es una operación idempotente. |
| [TestCancellation_MoveCreatesDatedShift](../tests/shift_rules_test.go) | Mover una ocurrencia cancela la original y crea turno fechado en sala destino. |
| [TestCancellation_MoveToARoomWithoutTheCategory](../tests/shift_rules_test.go) | Mover ocurrencia a sala sin la categoría falla de forma atómica. |
| [TestShiftScreen_SavesAndDeletes](../tests/shift_rules_test.go) | La pantalla de turnos guarda y elimina registros a través de crud. |
| [TestFormToShift_Conversions](../tests/shift_rules_test.go) | Conversión bidireccional entre RoomShiftForm y RoomShift validando rangos horarios. |
| [TestShiftsOnAndBoard](../tests/board_test.go) | Consulta del estado del tablero de salas para un día y rango horario. |
| [TestMoveShiftOccurrence](../tests/board_test.go) | Mover ocurrencia de turno entre salas con reubicación atómica. |
| [TestMap_LoadsServerData](../tests/map_data_test.go) | Carga de pisos, salas y artefactos desde el servidor hacia la vista del mapa. |
| [TestMap_AddFloorIsSaved](../tests/map_data_test.go) | Agregar piso desde el mapa persiste los datos correctamente. |
| [TestMap_BadCellsAreAnError](../tests/map_data_test.go) | Celdas con formato inválido emiten notificación de error. |
| [TestMap_GridNeverHidesWhatIsDrawn](../tests/map_data_test.go) | La cuadrícula se ajusta dinámicamente para no ocultar elementos dibujados. |
| [TestEquipment_DeviceIdOptional](../tests/equipment_device_test.go) | El identificador de dispositivo de red es opcional al crear equipamiento. |
| [TestEquipment_DeviceIdKept](../tests/equipment_device_test.go) | El identificador de dispositivo asignado se conserva en consultas posteriores. |
| [TestErrorKinds](../tests/error_kinds_test.go) | Cada error de dominio implementa el tipo model.ErrorKind correspondiente (NotFound, Conflict, Invalid). |
| [TestPresenterViewsCoverage](../tests/view_test.go) | Presentadores de lista exponen capacidades de consulta y mutación. |
| [TestItemsDeclaredOnFields](../tests/view_test.go) | Declaración de ItemLabel e ItemDescription en campos del esquema. |
| [TestUI_Screens](../tests/ui_view_test.go) | Renderizado y navegación entre pantallas de administración de salas y pisos. |
| [TestDuplicateRoomCode](../tests/room_test.go) | Detección y rechazo de códigos de sala duplicados dentro del tenant. |
| [TestListRoomDay](../tests/day_test.go) | Listado consolidado de turnos semanales y puntuales para una sala en un día. |
| [TestShiftOverlapCases](../tests/overlap_test.go) | Cobertura exhaustiva de casos de solapamiento horario en turnos de sala. |
| [TestOccupantOverlap](../tests/overlap_test.go) | Detección de solapamiento horario para un mismo ocupante en salas distintas. |
| [TestRoomOfOccupant](../tests/occupant_room_test.go) | Determinación de la sala ocupada por un profesional en un instante dado. |
| [TestStages_RenderTwice](../tests/stage_rerender_test.go) | Re-renderizado consistente de componentes de interfaz sin fugas de estado. |
| [TestListPublicAvailability](../tests/public_test.go) | Disponibilidad pública de salas calculada en intervalos configurables. |
| [TestScreenSaveAndDelete](../tests/screen_save_test.go) | Flujo de guardado y eliminación desde componentes de pantalla. |
| [TestScreenSaveIgnoresTenantFromRequest](../tests/screen_save_test.go) | El tenant enviado desde cliente en formulario es ignorado o rechazado. |
| [TestFindFreeRooms](../tests/free_test.go) | Búsqueda de salas libres filtradas por categoría y rango horario. |
| [TestLocalNowTimezone](../tests/localnow_test.go) | Conversión de tiempo local respetando la zona horaria inyectada en dependencias. |
| [TestCategoryOperations](../tests/category_test.go) | Asignación y consulta de categorías de atención asociadas a salas. |
| [TestFormWidgetsRegression](../tests/widgets_test.go) | Inicialización y configuración de widgets de entrada en formularios. |
| [TestSeedLoad](../tests/seed_test.go) | Semillado carga datos de demostración con pisos, salas, equipos y turnos iniciales. |
| [TestUICSS](../tests/ui_css_test.go) | Generación y compilación de reglas de estilo CSS para componentes visuales. |
| [TestEquipmentCRUD](../tests/equipment_test.go) | Operaciones CRUD sobre catálogo de equipamiento y asignación a salas. |
| [TestArea_RectangleIsOneEntry](../tests/area_test.go) | Codificación de área rectangular en formato compacto de celdas. |
| [TestArea_LShapeRoundTrips](../tests/area_test.go) | Codificación y decodificación de área en forma de L. |
| [TestArea_EmptyAndInvalid](../tests/area_test.go) | Rechazo de cadenas de área vacías o con formato corrupto. |
| [TestCell_StringParse](../tests/area_test.go) | Análisis y extracción de coordenadas de celda desde cadena. |
| [TestRects_LShapeIsTwo](../tests/area_test.go) | Descomposición de forma en L en dos rectángulos disjuntos. |
| [TestTenantIsolation](../tests/tenant_test.go) | Aislamiento estricto de repositorios de datos entre diferentes tenants. |
| [TestDemo_Valid](../tests/demo_test.go) | Verificación de validez y ejecución de la demo del paquete seed. |
| [TestDeleteFloorInUse](../tests/floor_test.go) | Rechazo al intentar borrar un piso que aún tiene salas asignadas. |
| [TestSaveShiftValidationAndOccupantSnapshot](../tests/shift_test.go) | Guardado de turno con captura de instantánea de datos del ocupante. |
| [TestDependencyValidation](../tests/deps_test.go) | Validación de dependencias obligatorias en New (IDs, Categories, Occupants, TenantID, Zone). |
| [TestWeekdayOptions](../tests/weekday_test.go) | Opciones de selector de día de la semana ordenadas de domingo a sábado. |
| [TestORMGeneratedMethodsCoverage](../tests/orm_coverage_test.go) | Cobertura de métodos generados por definitionc para entidades y formularios. |
| [TestEventPublishingOnMutations](../tests/events_test.go) | Publicación de eventos de dominio en operaciones de mutación. |
| [TestLocalAt](../day_internal_test.go) | Cálculo de fecha y hora local para un instante Unix en la zona horaria del módulo. |
| [TestMigrate_CreatesEveryTable](../migrate/migrate_test.go) | Creación de todas las tablas de persistencia requeridas por el módulo. |
| [TestMigrate_NoCompilerIsAnError](../migrate/migrate_test.go) | Rechazo de ejecución de migración sin compilador DDL configurado. |
