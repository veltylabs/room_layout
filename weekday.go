package roomlayout

import "webtyp.com/fmt"

// WeekdayOptions: the choices for RoomShiftForm.day_of_week. The empty key
// means "dated shift" (the date field is used instead).
func WeekdayOptions() []fmt.KeyValue {
	return []fmt.KeyValue{
		{Key: "", Value: "— specific date —"},
		{Key: WeekdayMonday, Value: "Monday"},
		{Key: WeekdayTuesday, Value: "Tuesday"},
		{Key: WeekdayWednesday, Value: "Wednesday"},
		{Key: WeekdayThursday, Value: "Thursday"},
		{Key: WeekdayFriday, Value: "Friday"},
		{Key: WeekdaySaturday, Value: "Saturday"},
		{Key: WeekdaySunday, Value: "Sunday"},
	}
}
