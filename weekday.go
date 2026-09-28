package roomlayout

import "webtyp.com/fmt"

// WeekdayOptions: the choices for RoomShiftForm.day_of_week. The empty key
// means "dated shift" (the date field is used instead).
func WeekdayOptions() []fmt.KeyValue {
	return []fmt.KeyValue{
		{Key: "", Value: "— fecha específica —"},
		{Key: WeekdayMonday, Value: "Lunes"},
		{Key: WeekdayTuesday, Value: "Martes"},
		{Key: WeekdayWednesday, Value: "Miércoles"},
		{Key: WeekdayThursday, Value: "Jueves"},
		{Key: WeekdayFriday, Value: "Viernes"},
		{Key: WeekdaySaturday, Value: "Sábado"},
		{Key: WeekdaySunday, Value: "Domingo"},
	}
}
