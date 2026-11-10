package roomlayout

// sortFloors ordena los pisos por posición y, a igual posición, por nombre.
func sortFloors(list []Floor) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && floorBefore(list[j], list[j-1]); j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

func floorBefore(a, b Floor) bool {
	if a.Position != b.Position {
		return a.Position < b.Position
	}
	return a.Name < b.Name
}

// sortRoomsByCode ordena las salas por código.
func sortRoomsByCode(list []Room) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].Code < list[j-1].Code; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// sortEquipmentByName ordena el equipamiento por nombre.
func sortEquipmentByName(list []Equipment) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].Name < list[j-1].Name; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}
