package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/time"
)

func parseDate(s string) (time.Unix, error) {
	if s == "" {
		return 0, nil
	}
	sec, err := time.ParseDateUTC(s)
	if err != nil {
		return 0, ErrInvalidDate
	}
	return sec, nil
}

func formatDate(sec time.Unix) string {
	if sec <= 0 {
		return ""
	}
	y, mo, d, _, _, _ := time.CivilUTC(sec)
	return fmt.Sprintf("%04d-%02d-%02d", y, mo, d)
}

func parseHour(s string) (int, error) {
	min, err := time.ParseTime(s)
	if err != nil {
		return 0, ErrInvalidRange
	}
	return int(min), nil
}

func formatHour(min int) string {
	h := min / 60
	m := min % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func shiftToForm(s RoomShift) RoomShiftForm {
	f := RoomShiftForm{
		Id:            s.Id,
		RoomId:        s.RoomId,
		CategoryId:    s.CategoryId,
		OccupantId:    s.OccupantId,
		OccupantLabel: s.OccupantLabel,
		Start:         formatHour(int(s.StartMin)),
		End:           formatHour(int(s.EndMin)),
	}
	if s.SpecificDate > 0 {
		f.Date = formatDate(s.SpecificDate)
	} else {
		f.DayOfWeek = fmt.Sprintf("%d", s.DayOfWeek)
	}
	return f
}

// weekdayOf convierte el día de la semana del formulario («0» domingo … «6» sábado) en su número.
// Ejemplo: n, err := weekdayOf(WeekdayMonday)   // 1
func weekdayOf(s string) (int, error) {
	days := []string{WeekdaySunday, WeekdayMonday, WeekdayTuesday, WeekdayWednesday, WeekdayThursday, WeekdayFriday, WeekdaySaturday}
	for i, d := range days {
		if s == d {
			return i, nil
		}
	}
	return 0, ErrInvalidWeekday
}

// formToShift convierte el formulario de la pantalla (horas «HH:MM», fecha «AAAA-MM-DD» o día de la semana) en el turno
// (minutos y segundos). No revisa reglas de negocio: eso lo hace checkShift.
// Ejemplo: s, err := formToShift(RoomShiftForm{RoomId: "r1", CategoryId: "cat1", DayOfWeek: WeekdayMonday, Start: "09:00", End: "10:00"})
func formToShift(f RoomShiftForm) (RoomShift, error) {
	start, errStart := parseHour(f.Start)
	end, errEnd := parseHour(f.End)
	if errStart != nil || errEnd != nil {
		return RoomShift{}, ErrInvalidRange
	}
	s := RoomShift{
		Id: f.Id, RoomId: f.RoomId, CategoryId: f.CategoryId, OccupantId: f.OccupantId,
		OccupantLabel: fmt.TrimSpace(f.OccupantLabel), StartMin: int64(start), EndMin: int64(end),
	}
	if f.Date == "" {
		dow, err := weekdayOf(f.DayOfWeek)
		if err != nil {
			return RoomShift{}, err
		}
		s.DayOfWeek = int64(dow)
		return s, nil
	}
	date, err := parseDate(f.Date)
	if err != nil || date <= 0 {
		return RoomShift{}, ErrInvalidDate
	}
	s.SpecificDate = date
	return s, nil
}
