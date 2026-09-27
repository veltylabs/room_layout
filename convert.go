package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/time"
)

func parseDate(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	nano, err := time.ParseDate(s)
	if err != nil {
		return 0, ErrInvalidDate
	}
	return nano / 1_000_000_000, nil
}

func formatDate(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.FormatDate(sec * 1_000_000_000)
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
